package tachyonVpnClient

import (
	"errors"
	"fmt"
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
	"io"
	"net"
	"os"
	"strconv"
	"sync"
	"tachyonSimpleVpnProtocol"
)

func ClientRun() {
	if len(os.Args) != 2 {
		panic("Usage: client 123.123.123.123")
	}
	vpnServerIp := os.Args[1]
	tun, err := clientCreateTun(vpnServerIp)
	udwErr.PanicIfError(err)
	conn, err := net.Dial("tcp", vpnServerIp+":"+strconv.Itoa(tachyonSimpleVpnProtocol.VpnPort))
	udwErr.PanicIfError(err)
	fmt.Println("Connected ✔")
	clientId := tachyonSimpleVpnProtocol.GetClientId()
	go func() {
		bufR := make([]byte, 2<<20)
		bufW := udwBytes.NewBufWriter(nil)
		vpnPacket := &tachyonSimpleVpnProtocol.VpnPacket{}
		for {
			n, err := tun.Read(bufR)
			udwErr.PanicIfError(err)
			vpnPacket.Cmd = tachyonSimpleVpnProtocol.CmdData
			vpnPacket.ClientIdFrom = clientId
			vpnPacket.Data = bufR[:n]
			bufW.Reset()
			vpnPacket.Encode(bufW)
			err = udwBinary.WriteByteSliceWithUint32LenNoAllocV2(conn, bufW.GetBytes())
			udwErr.PanicIfError(err)
		}
	}()
	go func() {
		bufR := make([]byte, 3<<20)
		vpnPacket := &tachyonSimpleVpnProtocol.VpnPacket{}
		for {
			out, err := udwBinary.ReadByteSliceWithUint32LenNoAllocLimitMaxSize(conn, bufR, uint32(len(bufR)))
			udwErr.PanicIfError(err)
			err = vpnPacket.Decode(out)
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
		udwNet.MustSetDnsServerAddr("8.8.8.8")
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
					udwNet.MustSetDnsServerToDefault()
				})
				if err != nil {
					udwLog.Log("error", "uninstallAllPassRoute", err.Error())
				}
			})
			return nil
		}),
	}, nil
}
