package tachyonVpnClient

import (
	"errors"
	"fmt"
	"github.com/tachyon-protocol/udw/udwConsole"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwIo"
	"github.com/tachyon-protocol/udw/udwIpPacket"
	"github.com/tachyon-protocol/udw/udwLog"
	"github.com/tachyon-protocol/udw/udwNet"
	"github.com/tachyon-protocol/udw/udwNet/udwIPNet"
	"github.com/tachyon-protocol/udw/udwNet/udwTapTun"
	"io"
	"net"
	"strconv"
	"sync"
	"tachyonSimpleVpnProtocol"
)

type ClientRunReq struct {
	IsForward           bool
	ForwardRelayIp      string
	ForwardExitClientId string

	ExitIp string
}

func ClientRun(req ClientRunReq) {
	vpnServerIp := req.ExitIp
	if req.IsForward {
		vpnServerIp = req.ForwardRelayIp
	}
	tun, err := clientCreateTun(vpnServerIp)
	udwErr.PanicIfError(err)
	vpnConn, err := tachyonSimpleVpnProtocol.VpnConnectionDial(vpnServerIp + ":" + strconv.Itoa(tachyonSimpleVpnProtocol.VpnPort))
	udwErr.PanicIfError(err)
	serverType := "EXIT"
	if req.IsForward {
		serverType = "RELAY"
	}
	fmt.Println("Connected to", serverType, "Server ✔")
	if req.IsForward {

	}
	clientId := tachyonSimpleVpnProtocol.GetClientId()
	go func() {
		bufR := make([]byte, 2<<20)
		vpnPacket := &tachyonSimpleVpnProtocol.VpnPacket{}
		for {
			n, err := tun.Read(bufR)
			udwErr.PanicIfError(err)
			vpnPacket.Cmd = tachyonSimpleVpnProtocol.CmdData
			vpnPacket.ClientIdFrom = clientId
			vpnPacket.Data = bufR[:n]
			err = vpnConn.Write(vpnPacket)
			udwErr.PanicIfError(err)
		}
	}()
	go func() {
		vpnPacket := &tachyonSimpleVpnProtocol.VpnPacket{}
		for {
			//out, err := udwBinary.ReadByteSliceWithUint32LenNoAllocLimitMaxSize(conn, bufR, uint32(len(bufR)))
			//udwErr.PanicIfError(err)
			//err = vpnPacket.Decode(out)
			//udwErr.PanicIfError(err)
			err := vpnConn.Read(vpnPacket)
			udwErr.PanicIfError(err)
			ipPacket, errMsg := udwIpPacket.NewIpv4PacketFromBuf(vpnPacket.Data)
			if errMsg != "" {
				panic("parse IPv4 failed:" + errMsg)
			}
			_, err = tun.Write(ipPacket.SerializeToBuf())
			if err != nil {
				//noinspection SpellCheckingInspection
				udwLog.Log("[wmwa2fyr9e] TUN Write error", err)
			}
		}
	}()
	udwConsole.WaitForExit()
}

func clientCreateTun(vpnServerIp string) (tun io.ReadWriteCloser, err error) {
	vpnClientIp := net.ParseIP("172.21.0.1")
	includeIpNetSet := udwIPNet.NewAllPassIpv4Net()
	includeIpNetSet.RemoveIpString(vpnServerIp)
	tunCreateCtx := &udwTapTun.CreateIpv4TunContext{
		SrcIp:        vpnClientIp,
		DstIp:        vpnClientIp,
		FirstIp:      vpnClientIp,
		DhcpServerIp: vpnClientIp,
		Mtu:          tachyonSimpleVpnProtocol.Mtu,
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
