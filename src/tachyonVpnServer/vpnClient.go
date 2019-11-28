package tachyonVpnServer

import (
	"crypto/tls"
	"fmt"
	"github.com/tachyon-protocol/udw/udwBinary"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwLog"
	"github.com/tachyon-protocol/udw/udwTlsSelfSignCertV2"
	"net"
	"tachyonVpnProtocol"
	"tlsPacketDebugger"
)

func (s *Server) getClient(clientId uint64) *vpnClient {
	s.locker.Lock()
	if s.clientMap == nil {
		s.clientMap = map[uint64]*vpnClient{}
	}
	client := s.clientMap[clientId]
	s.locker.Unlock()
	return client
}

func (s *Server) getOrNewClientFromDirectConn(clientId uint64, connToClient net.Conn) *vpnClient {
	s.locker.Lock()
	if s.clientMap == nil {
		s.clientMap = map[uint64]*vpnClient{}
	}
	client := s.clientMap[clientId]
	if client != nil {
		s.locker.Unlock()
		return client
	}
	client = &vpnClient{
		id:           clientId,
		connToClient: connToClient,
	}
	s.clientMap[client.id] = client
	err := s.clientAllocateVpnIp_NoLock(client)
	s.locker.Unlock()
	if err != nil {
		panic("[ub4fm53v26] " + err.Error())
	}
	return client
}

func (s *Server) getOrNewClientFromRelayConn(clientId uint64, relayConn net.Conn) *vpnClient {
	s.locker.Lock()
	if s.clientMap == nil {
		s.clientMap = map[uint64]*vpnClient{}
	}
	client := s.clientMap[clientId]
	if client != nil {
		s.locker.Unlock()
		return client
	}
	client = &vpnClient{
		id: clientId,
	}
	left, right := tachyonVpnProtocol.NewInternalConnectionDual()
	right = tls.Server(right, &tls.Config{
		Certificates: []tls.Certificate{ //TODO optimize allocate
			*udwTlsSelfSignCertV2.GetTlsCertificate(),
		},
		NextProtos:   []string{"http/1.1"},
		MinVersion: tls.VersionTLS12,
	})
	client.connToClient = right
	client.connRelaySide = left
	s.clientMap[client.id] = client
	err := s.clientAllocateVpnIp_NoLock(client)
	go s.clientTcpConnHandle(client.connToClient)
	s.locker.Unlock()
	if err != nil {
		panic("[ub4fm53v26] " + err.Error())
	}
	go func() {
		vpnPacket := &tachyonVpnProtocol.VpnPacket{
			Cmd:              tachyonVpnProtocol.CmdForward,
			ClientIdSender:   s.clientId,
			ClientIdReceiver: clientId,
		}
		buf := make([]byte, 16*1024)
		bufW := udwBytes.NewBufWriter(nil)
		for {
			n, err := client.connRelaySide.Read(buf)
			if err != nil {
				udwLog.Log("[cz2xvv1smx] close conn", err)
				_ = client.connRelaySide.Close()
				return
			}
			if tachyonVpnProtocol.Debug {
				fmt.Println("read from connRelaySide write to relayConn", vpnPacket.ClientIdSender, "->", vpnPacket.ClientIdReceiver)
				if tachyonVpnProtocol.Debug {
					tlsPacketDebugger.Dump("---", buf[:n])
				}
			}
			vpnPacket.Data = buf[:n]
			bufW.Reset()
			vpnPacket.Encode(bufW)
			err = udwBinary.WriteByteSliceWithUint32LenNoAllocV2(relayConn, bufW.GetBytes())
			if err != nil {
				udwLog.Log("[ar1nr4wf3s]", err)
				continue
			}
		}
	}()
	return client
}
