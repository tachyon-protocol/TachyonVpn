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
			if err != nil {
				udwLog.Log("[m7j1pw1vr7] TUN Read failed", err)
				continue
			}
			packetBuf := bufR[:n]
			ipPacket, errMsg := udwIpPacket.NewIpv4PacketFromBuf(packetBuf)
			if errMsg != "" {
				udwLog.Log("[wj1nz633mg] TUN Read parse IPv4 failed", errMsg)
				continue
			}
			ip := ipPacket.GetDstIp()
			client := s.getClientByVpnIp(ip)
			if client == nil {
				udwLog.Log("[r1tp9rk84m] TUN Read no such client")
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
			connToClient := <-acceptPipe
			connToClient = tls.Server(connToClient, &tls.Config{
				Certificates: certs,
				NextProtos:   []string{"http/1.1"},
			})
			go func() {
				vpnPacket := &tachyonSimpleVpnProtocol.VpnPacket{}
				bufW := udwBytes.NewBufWriter(nil)
				for {
					bufW.Reset()
					err := udwBinary.ReadByteSliceWithUint32LenToBufW(connToClient, bufW)
					if err != nil {
						udwLog.Log("[tw1me5hux3] close conn", err)
						_ = connToClient.Close()
						return
					}
					err = vpnPacket.Decode(bufW.GetBytes())
					if err != nil {
						udwLog.Log("[m1ds6vv2n8] close conn", err)
						_ = connToClient.Close()
						return
					}
					//client := s.getOrNewClient(vpnPacket.ClientIdFrom, connToClient, nil)
					switch vpnPacket.Cmd {
					case tachyonSimpleVpnProtocol.CmdData:
						client := s.getClient(vpnPacket.ClientIdFrom)
						ipPacket, errMsg := udwIpPacket.NewIpv4PacketFromBuf(vpnPacket.Data)
						if errMsg != "" {
							_ = connToClient.Close()
							return
						}
						vpnIp := udwNet.Ipv4AddAndCopyWithBuffer(READONLY_vpnIpStart, uint32(client.vpnIpOffset), bufW)
						ipPacket.SetSrcIp__NoRecomputeCheckSum(vpnIp)
						ipPacket.TcpFixMss__NoRecomputeCheckSum(tachyonSimpleVpnProtocol.Mss)
						ipPacket.RecomputeCheckSum()
						_, err = tun.Write(ipPacket.SerializeToBuf())
						if err != nil {
							udwLog.Log("[x8z73fm1x5] close conn", err)
							_ = connToClient.Close()
							return
						}
					case tachyonSimpleVpnProtocol.CmdForward:
						nextPeer := s.getClient(vpnPacket.ClientIdForwardTo)
						if nextPeer == nil {
							fmt.Println("[4tz1d2932g] forward failed nextPeer == nil")
							continue
						}
						_, err := nextPeer.connToClient.Write(vpnPacket.Data) //TLS layer
						if err != nil {
							fmt.Println("[va1gz58zm3] forward failed", err)
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
		//TODO handshake with Relay Server
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
						//TODO Server will use vpnPacket.ClientIdFrom to identify different TLS connections
						//TODO vpnPacket.ClientIdFrom should not be real Client's Id
						//TODO Relay Server could replace real Client's Id with fake one
						client := s.getOrNewClientFromRelayConn(vpnPacket.ClientIdFrom, relayConn, acceptPipe)
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
