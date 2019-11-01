package tachyonVpnClient

import (
	"github.com/tachyon-protocol/udw/udwBinary"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwLog"
	"net"
	"tachyonSimpleVpnProtocol"
)

func (s *server) getClient(clientId uint64) *vpnClient {
	s.locker.Lock()
	if s.clientMap == nil {
		s.clientMap = map[uint64]*vpnClient{}
	}
	client := s.clientMap[clientId]
	s.locker.Unlock()
	return client
}

func (s *server) getOrNewClientFromDirectConn(clientId uint64, connToClient net.Conn) *vpnClient {
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

func (s *server) getOrNewClientFromRelayConn(clientId uint64, relayConn net.Conn, acceptPipe <-chan net.Conn) *vpnClient {
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
	}
	cipher, plain := tachyonSimpleVpnProtocol.NewInternalConnectionDual()
	client.connToClient = plain
	client.connRelaySide = cipher
	s.clientMap[client.id] = client
	err := s.clientAllocateVpnIp_NoLock(client)
	s.locker.Unlock()
	if err != nil {
		panic("[ub4fm53v26] " + err.Error())
	}
	go func() {
		vpnPacket := &tachyonSimpleVpnProtocol.VpnPacket{
			Cmd:               tachyonSimpleVpnProtocol.CmdForward,
			ClientIdFrom:      s.clientId,
			ClientIdForwardTo: clientId,
		}
		buf := make([]byte, 3<<10)
		bufW := udwBytes.NewBufWriter(nil)
		for {
			n, err := client.connRelaySide.Read(buf)
			udwErr.PanicIfError(err)
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

