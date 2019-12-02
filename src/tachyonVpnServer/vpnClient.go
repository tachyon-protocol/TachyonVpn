package tachyonVpnServer

import (
	"crypto/tls"
	"fmt"
	"github.com/tachyon-protocol/udw/udwBinary"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwLog"
	"github.com/tachyon-protocol/udw/udwTlsSelfSignCertV2"
	"net"
	"sync"
	"tachyonVpnProtocol"
	"tlsPacketDebugger"
)

type vpnClient struct {
	id          uint64
	vpnIpOffset int
	vpnIp       net.IP

	connLock      sync.Mutex
	connToClient  net.Conn
	connRelaySide net.Conn
}

func (vc *vpnClient) getConnToClient() net.Conn {
	vc.connLock.Lock()
	conn := vc.connToClient
	vc.connLock.Unlock()
	return conn
}

func (s *Server) getClient(clientId uint64) *vpnClient {
	s.lock.Lock()
	if s.clientMap == nil {
		s.clientMap = map[uint64]*vpnClient{}
	}
	client := s.clientMap[clientId]
	s.lock.Unlock()
	return client
}

func (s *Server) newOrUpdateClientFromDirectConn(clientId uint64, connToClient net.Conn) {
	s.lock.Lock()
	if s.clientMap == nil {
		s.clientMap = map[uint64]*vpnClient{}
	}
	client := s.clientMap[clientId]
	if client != nil {
		client.connLock.Lock()
		client.connToClient = connToClient //reconnect
		client.connLock.Unlock()
		s.lock.Unlock()
		return
	}
	client = &vpnClient{
		id:           clientId,
		connToClient: connToClient,
	}
	s.clientMap[client.id] = client
	err := s.clientAllocateVpnIp_NoLock(client)
	s.lock.Unlock()
	if err != nil {
		panic("[ub4fm53v26] " + err.Error())
	}
	return
}

func (s *Server) getOrNewClientFromRelayConn(clientId uint64) *vpnClient {
	s.lock.Lock()
	if s.clientMap == nil {
		s.clientMap = map[uint64]*vpnClient{}
	}
	client := s.clientMap[clientId]
	if client != nil {
		s.lock.Unlock()
		return client
	}
	client = &vpnClient{
		id: clientId,
	}
	left, right := tachyonVpnProtocol.NewInternalConnectionDual(func() {
		s.lock.Lock()
		delete(s.clientMap, clientId)
		s.lock.Unlock()
	}, nil)
	right = tls.Server(right, &tls.Config{
		Certificates: []tls.Certificate{ //TODO optimize allocate
			*udwTlsSelfSignCertV2.GetTlsCertificate(),
		},
		NextProtos: []string{"http/1.1"},
		MinVersion: tls.VersionTLS12,
	})
	client.connToClient = right
	client.connRelaySide = left
	s.clientMap[client.id] = client
	err := s.clientAllocateVpnIp_NoLock(client)
	go s.clientTcpConnHandle(client.getConnToClient())
	s.lock.Unlock()
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
			err = udwBinary.WriteByteSliceWithUint32LenNoAllocV2(s.getRelayConn(), bufW.GetBytes()) //TODO lock
			if err != nil {
				udwLog.Log("[ar1nr4wf3s]", err)
				continue
			}
		}
	}()
	return client
}
