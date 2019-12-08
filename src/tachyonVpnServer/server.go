package tachyonVpnServer

import (
	"crypto/tls"
	"fmt"
	"github.com/tachyon-protocol/udw/udwBinary"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwClose"
	"github.com/tachyon-protocol/udw/udwConsole"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwIpPacket"
	"github.com/tachyon-protocol/udw/udwLog"
	"github.com/tachyon-protocol/udw/udwNet"
	"github.com/tachyon-protocol/udw/udwNet/udwTapTun"
	"github.com/tachyon-protocol/udw/udwTlsSelfSignCertV2"
	"net"
	"strconv"
	"sync"
	"tachyonVpnProtocol"
	"tyTls"
)

type ServerRunReq struct {
	UseRelay        bool
	RelayServerIp   string
	RelayServerTKey string

	SelfTKey string
}

type Server struct {
	clientId               uint64
	tun                    *udwTapTun.TunTapObj
	relayConnKeepAliveChan chan uint64

	lock           sync.Mutex
	clientMap      map[uint64]*vpnClient
	vpnIpList      [maxCountVpnIp]*vpnClient
	nextVpnIpIndex int
	relayConn      net.Conn

	req ServerRunReq
}

func (s *Server) Run(req ServerRunReq) {
	tyTls.EnableTlsVersion13()
	s.req = req
	s.clientId = tachyonVpnProtocol.GetClientId(0)
	fmt.Println("ClientId:", s.clientId)
	tun, err := udwTapTun.NewTun("")
	udwErr.PanicIfError(err)
	err = udwTapTun.SetP2PIpAndUp(udwTapTun.SetP2PIpRequest{
		IfaceName: tun.Name(),
		SrcIp:     udwNet.Ipv4AddAndCopyWithBuffer(READONLY_vpnIpStart, 2, nil),
		DstIp:     udwNet.Ipv4AddAndCopyWithBuffer(READONLY_vpnIpStart, 1, nil),
		Mtu:       tachyonVpnProtocol.Mtu,
		Mask:      net.CIDRMask(16, 32),
	})
	udwErr.PanicIfError(err)
	s.tun = tun
	networkConfig()
	tlsServerCert:=udwTlsSelfSignCertV2.GetTlsCertificate()
	sTlsConfig,errMsg:=tyTls.NewServerTlsConfigWithChk(tyTls.NewServerTlsConfigWithChkReq{
		ServerCert: *tlsServerCert,
	})
	udwErr.PanicIfErrorMsg(errMsg)
	fmt.Println("ServerChk: "+tyTls.MustHashChkFromTlsCert(tlsServerCert))
	fmt.Println("Server started ✔")
	//read thread from TUN
	go func() {
		bufR := make([]byte, 16*1024)
		bufW := udwBytes.NewBufWriter(nil)
		vpnPacket := &tachyonVpnProtocol.VpnPacket{
			ClientIdSender:   s.clientId,
			Cmd:              tachyonVpnProtocol.CmdData,
		}
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
			if udwNet.IsPrivateNetwork(ipPacket.GetSrcIp()) {
				//udwLog.Log("[ye723euu1ah] private IP address is not allowed", ipPacket.GetSrcAddrString())
				continue
			}
			client := s.getClientByVpnIp(ip)
			if client == nil {
				udwLog.Log("[r1tp9rk84m] TUN Read no such client", ipPacket.GetSrcAddrString())
				continue
			}
			//vpnPacket := &tachyonVpnProtocol.VpnPacket{
			//	ClientIdSender:   s.clientId,
			//	ClientIdReceiver: client.id,
			//	Cmd:              tachyonVpnProtocol.CmdData,
			//}
			vpnPacket.ClientIdReceiver = client.id
			if tachyonVpnProtocol.Debug {
				fmt.Println("read from tun, write to client", vpnPacket.ClientIdSender, "->", vpnPacket.ClientIdReceiver)
			}
			ipPacket.SetDstIp__NoRecomputeCheckSum(READONLY_vpnIpClient)
			ipPacket.TcpFixMss__NoRecomputeCheckSum(tachyonVpnProtocol.Mss)
			ipPacket.RecomputeCheckSum()
			vpnPacket.Data = ipPacket.SerializeToBuf()
			bufW.Reset()
			vpnPacket.Encode(bufW)
			_ = udwBinary.WriteByteSliceWithUint32LenNoAllocV2(client.getConnToClient(), bufW.GetBytes()) //TODO
		}
	}()
	closer := udwClose.NewCloser()
	//two methods to accept new vpn conn
	if req.UseRelay {
		err := s.connectToRelay()
		udwErr.PanicIfError(err)
		s.relayConnKeepAliveThread()
	} else {
		_close := udwNet.TcpNewListener(":"+strconv.Itoa(tachyonVpnProtocol.VpnPort), func(conn net.Conn) {
			conn = tls.Server(conn, sTlsConfig)
			s.clientTcpConnHandle(conn)
		})
		closer.AddOnClose(_close)
	}
	udwConsole.WaitForExit()
	closer.Close()
}

func (s *Server) clientTcpConnHandle(connToClient net.Conn) {
	vpnPacket := &tachyonVpnProtocol.VpnPacket{}
	bufW := udwBytes.NewBufWriter(nil)
	for {
		bufW.Reset()
		err := udwBinary.ReadByteSliceWithUint32LenToBufW(connToClient, bufW)
		if err != nil {
			udwLog.Log("[tw1me5hux3] close conn", err, connToClient.RemoteAddr())
			_ = connToClient.Close()
			return
		}
		err = vpnPacket.Decode(bufW.GetBytes())
		if err != nil {
			udwLog.Log("[m1ds6vv2n8] close conn", err)
			_ = connToClient.Close()
			return
		}
		switch vpnPacket.Cmd {
		case tachyonVpnProtocol.CmdPing, tachyonVpnProtocol.CmdKeepAlive:
			bufW.Reset()
			vpnPacket.ClientIdReceiver = vpnPacket.ClientIdSender
			vpnPacket.ClientIdSender = s.clientId
			vpnPacket.Encode(bufW)
			err := udwBinary.WriteByteSliceWithUint32LenNoAllocV2(connToClient, bufW.GetBytes())
			if err != nil {
				udwLog.Log("[2cpj1sbv37s] close conn", err)
				_ = connToClient.Close()
				return
			}
		case tachyonVpnProtocol.CmdHandshake:
			if s.req.SelfTKey == "" {
				s.newOrUpdateClientFromDirectConn(vpnPacket.ClientIdSender, connToClient)
				udwLog.Log("[4z734vc9pn] New client sent handshake ✔ server not require TKey", connToClient.RemoteAddr())
			} else if len(s.req.SelfTKey) == len(string(vpnPacket.Data)) && s.req.SelfTKey == string(vpnPacket.Data) {
				s.newOrUpdateClientFromDirectConn(vpnPacket.ClientIdSender, connToClient)
				udwLog.Log("[agz7rzq1kr9] New client TKey matched ✔", connToClient.RemoteAddr())
			} else {
				_ = connToClient.Close()
				udwLog.Log("[wzh56ty1bur] New client TKey not match ✘ close conn", connToClient.RemoteAddr())
			}
		case tachyonVpnProtocol.CmdData:
			client := s.getClient(vpnPacket.ClientIdSender)
			if client == nil {
				_ = connToClient.Close()
				udwLog.Log("[k692xqw1d2n] CmdData close conn cause no such client", vpnPacket.ClientIdSender, connToClient.RemoteAddr())
				return
			}
			//client := s.getOrNewClientFromDirectConn(vpnPacket.ClientIdSender, connToClient)
			ipPacket, errMsg := udwIpPacket.NewIpv4PacketFromBuf(vpnPacket.Data)
			if errMsg != "" {
				udwLog.Log("[txd5xn4ex7] close conn", errMsg, "ipPacket.IsIpv4:", ipPacket.IsIpv4(), "ipPacket.Ipv4HasMoreFragments:", ipPacket.Ipv4HasMoreFragments())
				_ = connToClient.Close()
				return
			}
			vpnIp := udwNet.Ipv4AddAndCopyWithBuffer(READONLY_vpnIpStart, uint32(client.vpnIpOffset), bufW)
			ipPacket.SetSrcIp__NoRecomputeCheckSum(vpnIp)
			ipPacket.TcpFixMss__NoRecomputeCheckSum(tachyonVpnProtocol.Mss)
			ipPacket.RecomputeCheckSum()
			_, err = s.tun.Write(ipPacket.SerializeToBuf())
			if err != nil {
				udwLog.Log("[x8z73fm1x5] close conn", err)
				_ = connToClient.Close()
				return
			}
		case tachyonVpnProtocol.CmdForward:
			client := s.getClient(vpnPacket.ClientIdSender)
			if client == nil {
				_ = connToClient.Close()
				udwLog.Log("[be8meu1vhm1d] CmdForward close conn cause no such client", vpnPacket.ClientIdSender, connToClient.RemoteAddr())
				return
			}
			nextPeer := s.getClient(vpnPacket.ClientIdReceiver)
			if nextPeer == nil {
				udwLog.Log("[4tz1d2932g] forward failed nextPeer[", vpnPacket.ClientIdReceiver, "] == nil")
				continue
			}
			err := udwBinary.WriteByteSliceWithUint32LenNoAllocV2(nextPeer.getConnToClient(), bufW.GetBytes()) //TLS layer
			if err != nil {
				udwLog.Log("[va1gz58zm3] forward failed", err)
				continue
			}
		default:
			_ = connToClient.Close()
			udwLog.Log("[rjb3nay1ezg] Cmd unknown", vpnPacket.Cmd, "close conn", connToClient.RemoteAddr())
			return
		}
	}
}
