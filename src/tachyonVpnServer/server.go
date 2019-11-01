package tachyonVpnClient

import (
	"crypto/tls"
	"fmt"
	"github.com/tachyon-protocol/udw/udwBinary"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwConsole"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwIpPacket"
	"github.com/tachyon-protocol/udw/udwLog"
	"github.com/tachyon-protocol/udw/udwNet"
	"github.com/tachyon-protocol/udw/udwNet/udwTapTun"
	"github.com/tachyon-protocol/udw/udwRand"
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

	locker        sync.Mutex
	connToClient  net.Conn
	connRelaySide net.Conn
}

type ServerRunReq struct {
	UseRelay      bool
	RelayServerIp string
}

type server struct {
	clientId uint64
	locker         sync.Mutex
	clientMap      map[uint64]*vpnClient
	vpnIpList      [maxCountVpnIp]*vpnClient
	nextVpnIpIndex int
}

func (s *server) Run(req ServerRunReq) {
	s.clientId = tachyonSimpleVpnProtocol.GetClientId() //TODO fixed clientId
	fmt.Println("ClientId:", s.clientId)
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
	fmt.Println("Server started ✔")

	//read thread from TUN
	go func() {
		bufR := make([]byte, 3<<10)
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
			client := s.getClientByVpnIp(ip)
			if client == nil {
				//noinspection SpellCheckingInspection
				udwLog.Log("[rdtp9rk84m] TUN Read no such client")
				continue
			}
			responseVpnPacket := &tachyonSimpleVpnProtocol.VpnPacket{
				ClientIdFrom: s.clientId,
				Cmd:          tachyonSimpleVpnProtocol.CmdData,
			}
			ipPacket.SetDstIp__NoRecomputeCheckSum(READONLY_vpnIpClient)
			ipPacket.TcpFixMss__NoRecomputeCheckSum(tachyonSimpleVpnProtocol.Mss)
			ipPacket.RecomputeCheckSum()
			responseVpnPacket.Data = ipPacket.SerializeToBuf()
			bufW.Reset()
			responseVpnPacket.Encode(bufW)
			_ = udwBinary.WriteByteSliceWithUint32LenNoAllocV2(client.connToClient, bufW.GetBytes()) //TODO
		}
	}()

	var (
		acceptPipe = make(chan net.Conn, 10<<10)
	)
	//read thread from vpn conn
	go func() {
		certs := []tls.Certificate{
			*udwTlsSelfSignCertV2.GetTlsCertificate(),
		}
		for {
			conn := <-acceptPipe
			conn = tls.Server(conn, &tls.Config{
				Certificates: certs,
				NextProtos:   []string{"http/1.1"},
			})
			go func() {
				bufR := make([]byte, 3<<10)
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
					client := s.getOrNewClient(vpnPacket.ClientIdFrom, conn, nil)
					switch vpnPacket.Cmd {
					case tachyonSimpleVpnProtocol.CmdData:
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
					case tachyonSimpleVpnProtocol.CmdForward:
						nextPeer := s.getClient(vpnPacket.ClientIdForwardTo)
						if nextPeer == nil {
							fmt.Println("[4tz1d2932g] nextPeer == nil")
							continue
						}
						_, err := nextPeer.connToClient.Write(vpnPacket.Data) //TLS layer
						if err != nil {
							fmt.Println("[va1gz58zm3]", err)
							continue
						}
					}
				}
			}()
		}
	}()

	//two methods to accept new vpn conn
	if req.UseRelay {
		relayConn, err := net.Dial("tcp", req.RelayServerIp+":"+strconv.Itoa(tachyonSimpleVpnProtocol.VpnPort))
		udwErr.PanicIfError(err)
		fmt.Println("Server connected to relay server[", req.RelayServerIp, "] ✔")
		relayConn = tls.Client(relayConn, &tls.Config{
			ServerName:         udwRand.MustCryptoRandToReadableAlpha(5) + ".com",
			InsecureSkipVerify: true,
			NextProtos:         []string{"http/1.1", "h2"},
		})
		go func() {
			vpnPacket := &tachyonSimpleVpnProtocol.VpnPacket{}
			buf := udwBytes.NewBufWriter(nil)
			for {
				err := udwBinary.ReadByteSliceWithUint32LenToBufW(relayConn, buf)
				udwErr.PanicIfError(err)
				err = vpnPacket.Decode(buf.GetBytes())
				udwErr.PanicIfError(err)
				if vpnPacket.Cmd == tachyonSimpleVpnProtocol.CmdForward {
					if vpnPacket.ClientIdForwardTo == s.clientId {
						client := s.getOrNewClient(vpnPacket.ClientIdFrom, nil,relayConn)
						acceptPipe <- client.connToClient
						_, err := client.connRelaySide.Write(vpnPacket.Data)
						if err != nil {
							udwLog.Log("[dy11zv1eg6]", err)
						}
					} else {
						fmt.Println("[vw9tm9rv2s] not forward to self")
					}
				} else {
					fmt.Println("[d39e7d859m]Unexpected Cmd[", vpnPacket.Cmd, "]")
				}
			}
		}()
	} else {
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(tachyonSimpleVpnProtocol.VpnPort))
		udwErr.PanicIfError(err)
		go func() {
			for {
				conn, err := ln.Accept()
				udwErr.PanicIfError(err)
				acceptPipe <- conn
			}
		}()
	}
	udwConsole.WaitForExit()
}
