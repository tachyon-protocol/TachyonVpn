package main

import (
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwIo"
	"github.com/tachyon-protocol/udw/udwLog"
	"github.com/tachyon-protocol/udw/udwNet"
	"github.com/tachyon-protocol/udw/udwNet/udwTapTun"
	"net"
	"sync"
)

func main() {
	vpnServerIp := ""
	vpnClientIp := net.ParseIP("172.21.0.1")
	tunCreateCtx := &udwTapTun.CreateIpv4TunContext{
		SrcIp:        vpnClientIp,
		DstIp:        vpnClientIp,
		FirstIp:      vpnClientIp,
		DhcpServerIp: vpnClientIp,
		Mtu:          1300, //TODO
		Mask:         net.CIDRMask(30, 32),
	}
	err := udwTapTun.CreateIpv4Tun(tunCreateCtx)
	if err != nil {
		panic(err)
	}
	tunNamed := tunCreateCtx.ReturnTun
	vpnGatewayIp := vpnClientIp
	err = udwErr.PanicToError(func() {
		udwNet.MustSetDnsServerAddr("8.8.8.8")
		ctx := udwNet.NewRouteContext()
		for _, ipNet := range setting.IncludeIpNetSet.GetIpv4NetList() {
			goIpNet := ipNet.ToGoIPNet()
			ctx.MustRouteSet(*goIpNet, vpnGatewayIp)
		}
	})
	if err != nil {
		tunNamed.Close()
		return nil, err
	}
	var closeOnce sync.Once
	return udwIo.StructWriterReaderCloser{
		Reader: tunNamed,
		Writer: tunNamed,
		Closer: udwIo.CloserFunc(func() error {
			closeOnce.Do(func() {
				tunNamed.Close()
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
