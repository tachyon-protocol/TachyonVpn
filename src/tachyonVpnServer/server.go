package tachyonVpnClient

import (
	"crypto/tls"
	"fmt"
	"github.com/tachyon-protocol/udw/udwBinary"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwIpPacket"
	"github.com/tachyon-protocol/udw/udwLog"
	"github.com/tachyon-protocol/udw/udwNet"
	"github.com/tachyon-protocol/udw/udwNet/udwTapTun"
	"github.com/tachyon-protocol/udw/udwTlsSelfSignCertV2"
	"net"
	"strconv"
	"sync"
	"tachyonSimpleVpnProtocol"
)

type vpnClient struct {
	id          uint64
	vpnIpOffset int
	vpnIp       net.IP

	locker sync.Mutex
	conn   net.Conn
}

var (
	gLocker         sync.Mutex
	gClientMap      map[uint64]*vpnClient
	gVpnIpList      [maxCountVpnIp]*vpnClient
	gNextVpnIpIndex int
)

func ServerRun() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(tachyonSimpleVpnProtocol.VpnPort))
	udwErr.PanicIfError(err)
	tun, err := udwTapTun.NewTun("")
	udwErr.PanicIfError(err)
	err = udwTapTun.SetP2PIpAndUp(udwTapTun.SetP2PIpRequest{
		IfaceName: tun.Name(),
		SrcIp:     udwNet.Ipv4AddAndCopyWithBuffer(READONLY_vpnIpStart, 2, nil),
		DstIp:     udwNet.Ipv4AddAndCopyWithBuffer(READONLY_vpnIpStart, 1, nil),
		Mtu:       tachyonSimpleVpnProtocol.Mtu,
		Mask:      net.CIDRMask(16, 32),
	})
	udwErr.PanicIfError(err)
	networkConfig()
	clientId := tachyonSimpleVpnProtocol.GetClientId()
	go func() {
		bufR := make([]byte, 3<<20)
		bufW := udwBytes.NewBufWriter(nil)
		for {
			n, err := tun.Read(bufR)
			udwErr.PanicIfError(err)
			packetBuf := bufR[:n]
			ipPacket, errMsg := udwIpPacket.NewIpv4PacketFromBuf(packetBuf)
			if errMsg != "" {
				//noinspection SpellCheckingInspection
				udwLog.Log("[psmddnegwg] TUN Read parse IPv4 failed", errMsg)
				return
			}
			ip := ipPacket.GetDstIp()
			client := getClientByVpnIp(ip)
			if client == nil {
				//noinspection SpellCheckingInspection
				udwLog.Log("[rdtp9rk84m] TUN Read no such client")
				continue
			}
			responseVpnPacket := &tachyonSimpleVpnProtocol.VpnPacket{
				ClientIdFrom: clientId,
				Cmd:          tachyonSimpleVpnProtocol.CmdData,
			}
			ipPacket.SetDstIp__NoRecomputeCheckSum(READONLY_vpnIpClient)
			ipPacket.TcpFixMss__NoRecomputeCheckSum(tachyonSimpleVpnProtocol.Mss)
			ipPacket.RecomputeCheckSum()
			responseVpnPacket.Data = ipPacket.SerializeToBuf()
			bufW.Reset()
			responseVpnPacket.Encode(bufW)
			_ = udwBinary.WriteByteSliceWithUint32LenNoAllocV2(client.conn, bufW.GetBytes()) //TODO
		}
	}()
	fmt.Println("Server started ✔")
	certs := []tls.Certificate{
		*udwTlsSelfSignCertV2.GetTlsCertificate(),
	}
	for {
		conn, err := ln.Accept()
		udwErr.PanicIfError(err)
		if tachyonSimpleVpnProtocol.Debug {
			udwLog.Log("New Conn", conn.RemoteAddr())
		}
		conn = tls.Server(conn, &tls.Config{
			Certificates: certs,
			NextProtos:   []string{"http/1.1"},
		})
		go func() {
			bufR := make([]byte, 3<<20)
			vpnPacket := &tachyonSimpleVpnProtocol.VpnPacket{}
			vpnIpBufW := udwBytes.NewBufWriter(nil)
			for {
				out, err := udwBinary.ReadByteSliceWithUint32LenNoAllocLimitMaxSize(conn, bufR, uint32(len(bufR)))
				if err != nil {
					_ = conn.Close()
					return
				}
				err = vpnPacket.Decode(out)
				if err != nil {
					_ = conn.Close()
					return
				}
				client := getClient(vpnPacket.ClientIdFrom, conn)
				ipPacket, errMsg := udwIpPacket.NewIpv4PacketFromBuf(vpnPacket.Data)
				if errMsg != "" {
					_ = conn.Close()
					return
				}
				vpnIp := udwNet.Ipv4AddAndCopyWithBuffer(READONLY_vpnIpStart, uint32(client.vpnIpOffset), vpnIpBufW)
				ipPacket.SetSrcIp__NoRecomputeCheckSum(vpnIp)
				ipPacket.TcpFixMss__NoRecomputeCheckSum(tachyonSimpleVpnProtocol.Mss)
				ipPacket.RecomputeCheckSum()
				_, err = tun.Write(ipPacket.SerializeToBuf())
				if err != nil {
					_ = conn.Close()
					return
				}
			}
		}()
	}
}
