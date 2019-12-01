package tachyonVpnClient

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"github.com/tachyon-protocol/udw/udwBinary"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwConsole"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwIo"
	"github.com/tachyon-protocol/udw/udwIpPacket"
	"github.com/tachyon-protocol/udw/udwLog"
	"github.com/tachyon-protocol/udw/udwNet"
	"github.com/tachyon-protocol/udw/udwNet/udwIPNet"
	"github.com/tachyon-protocol/udw/udwNet/udwTapTun"
	"github.com/tachyon-protocol/udw/udwRand"
	"io"
	"net"
	"sync"
	"tachyonVpnProtocol"
	"time"
	"tyTls"
)

type RunReq struct {
	ServerIp   string
	ServerTKey string

	IsRelay            bool
	ExitServerClientId uint64 //required when IsRelay is true
	ExitServerTKey     string //required when IsRelay is true
}

type Client struct {
	req                  RunReq
	clientId             uint64
	clientIdToExitServer uint64
	keepAliveChan        chan uint64
	connLock             sync.Mutex
	directVpnConn        net.Conn
	vpnConn              net.Conn
}

func (c *Client) Run(req RunReq) {
	c.req = req
	tyTls.AllowTlsVersion13()
	c.clientId = tachyonVpnProtocol.GetClientId()
	c.clientIdToExitServer = c.clientId
	if req.IsRelay {
		c.clientIdToExitServer = tachyonVpnProtocol.GetClientId()
		if req.ExitServerClientId == 0 {
			panic("ExitServerClientId can be empty when use relay mode")
		}
	}
	tun, err := createTun(req.ServerIp)
	udwErr.PanicIfError(err)
	err = c.connect()
	c.keepAliveThread()
	udwErr.PanicIfError(err)
	go func() {
		vpnPacket := &tachyonVpnProtocol.VpnPacket{
			Cmd:              tachyonVpnProtocol.CmdData,
			ClientIdSender:   c.clientIdToExitServer,
			ClientIdReceiver: req.ExitServerClientId,
		}
		buf := make([]byte, 16*1024)
		bufW := udwBytes.NewBufWriter(nil)
		c.connLock.Lock()
		vpnConn := c.vpnConn
		c.connLock.Unlock()
		for {
			n, err := tun.Read(buf)
			if err != nil {
				panic("[upe1hcb1q39h] " + err.Error())
			}
			vpnPacket.Data = buf[:n]
			bufW.Reset()
			vpnPacket.Encode(bufW)
			for {
				err = udwBinary.WriteByteSliceWithUint32LenNoAllocV2(vpnConn, bufW.GetBytes())
				if err != nil {
					c.connLock.Lock()
					_vpnConn := c.vpnConn
					c.connLock.Unlock()
					if vpnConn == _vpnConn {
						time.Sleep(time.Millisecond * 50)
					} else {
						vpnConn = _vpnConn
						udwLog.Log("[mpy2nwx1qck] tun read use new vpn conn")
					}
					continue
				}
				break
			}
		}
	}()
	go func() {
		vpnPacket := &tachyonVpnProtocol.VpnPacket{}
		buf := udwBytes.NewBufWriter(nil)
		c.connLock.Lock()
		vpnConn := c.vpnConn
		c.connLock.Unlock()
		for {
			buf.Reset()
			for {
				err := udwBinary.ReadByteSliceWithUint32LenToBufW(vpnConn, buf)
				if err != nil {
					c.connLock.Lock()
					_vpnConn := c.vpnConn
					c.connLock.Unlock()
					if vpnConn == _vpnConn {
						time.Sleep(time.Millisecond * 50)
					} else {
						vpnConn = _vpnConn
						udwLog.Log("[zdb1mbq1v1kxh] vpn conn read use new vpn conn")
					}
					continue
				}
				break
			}
			err = vpnPacket.Decode(buf.GetBytes())
			udwErr.PanicIfError(err)
			switch vpnPacket.Cmd {
			case tachyonVpnProtocol.CmdData:
				ipPacket, errMsg := udwIpPacket.NewIpv4PacketFromBuf(vpnPacket.Data)
				if errMsg != "" {
					udwLog.Log("[zdy1mx9y3h]", errMsg)
					continue
				}
				_, err = tun.Write(ipPacket.SerializeToBuf())
				if err != nil {
					udwLog.Log("[wmw12fyr9e] TUN Write error", err)
				}
			case tachyonVpnProtocol.CmdKeepAlive:
				i := binary.LittleEndian.Uint64(vpnPacket.Data)
				c.keepAliveChan <- i
			default:
				udwLog.Log("[h67hrf4kda] unexpect cmd", vpnPacket.Cmd)
			}
		}
	}()
	udwConsole.WaitForExit()
}

func createTun(vpnServerIp string) (tun io.ReadWriteCloser, err error) {
	vpnClientIp := net.ParseIP("172.21.0.1")
	includeIpNetSet := udwIPNet.NewAllPassIpv4Net()
	includeIpNetSet.RemoveIpString(vpnServerIp)
	tunCreateCtx := &udwTapTun.CreateIpv4TunContext{
		SrcIp:        vpnClientIp,
		DstIp:        vpnClientIp,
		FirstIp:      vpnClientIp,
		DhcpServerIp: vpnClientIp,
		Mtu:          tachyonVpnProtocol.Mtu,
		Mask:         net.CIDRMask(30, 32),
	}
	err = udwTapTun.CreateIpv4Tun(tunCreateCtx)
	if err != nil {
		return nil, errors.New("[3xa38g7vtd] " + err.Error())
	}
	tunNamed := tunCreateCtx.ReturnTun
	vpnGatewayIp := vpnClientIp
	err = udwErr.PanicToError(func() {
		configLocalNetwork()
		ctx := udwNet.NewRouteContext()
		for _, ipNet := range includeIpNetSet.GetIpv4NetList() {
			goIpNet := ipNet.ToGoIPNet()
			ctx.MustRouteSet(*goIpNet, vpnGatewayIp)
		}
	})
	if err != nil {
		_ = tunNamed.Close()
		return nil, errors.New("[r8y8d5ash4] " + err.Error())
	}
	var closeOnce sync.Once
	return udwIo.StructWriterReaderCloser{
		Reader: tunNamed,
		Writer: tunNamed,
		Closer: udwIo.CloserFunc(func() error {
			closeOnce.Do(func() {
				_ = tunNamed.Close()
				err := udwErr.PanicToError(func() {
					recoverLocalNetwork()
				})
				if err != nil {
					udwLog.Log("error", "uninstallAllPassRoute", err.Error())
				}
			})
			return nil
		}),
	}, nil
}

func newInsecureClientTlsConfig() *tls.Config {
	return &tls.Config{
		ServerName:         udwRand.MustCryptoRandToReadableAlpha(5) + ".com",
		InsecureSkipVerify: true,
		NextProtos:         []string{"http/1.1", "h2"},
		MinVersion:         tls.VersionTLS12,
	}
}
