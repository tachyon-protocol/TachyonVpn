package tachyonVpnServer

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"github.com/tachyon-protocol/udw/udwBinary"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwLog"
	"github.com/tachyon-protocol/udw/udwRand"
	"net"
	"strconv"
	"tachyonVpnProtocol"
	"time"
)

func (s *Server) getRelayConn() net.Conn {
	s.lock.Lock()
	conn := s.relayConn
	s.lock.Unlock()
	return conn
}

func (s *Server) connectToRelay () error {
	relayConn, err := net.Dial("tcp", s.req.RelayServerIp+":"+strconv.Itoa(tachyonVpnProtocol.VpnPort))
	if err != nil {
		return errors.New("[ytz6836s2w] "+err.Error())
	}
	udwLog.Log("Server connected to relay server[", s.req.RelayServerIp, "] ✔")
	relayConn = tls.Client(relayConn, &tls.Config{
		ServerName:         udwRand.MustCryptoRandToReadableAlpha(5) + ".com",
		InsecureSkipVerify: true,
		NextProtos:         []string{"http/1.1", "h2"},
		MinVersion:         tls.VersionTLS12,
	})
	var (
		vpnPacket = &tachyonVpnProtocol.VpnPacket{
			Cmd:            tachyonVpnProtocol.CmdHandshake,
			ClientIdSender: s.clientId,
			Data:           []byte(s.req.RelayServerTKey),
		}
		buf = udwBytes.NewBufWriter(nil)
	)
	vpnPacket.Encode(buf)
	err = udwBinary.WriteByteSliceWithUint32LenNoAllocV2(relayConn, buf.GetBytes())
	if err != nil {
		return errors.New("[tkb1nd2q3ec] "+err.Error())
	}
	vpnPacket.Reset()
	go func() {
		for {
			buf.Reset()
			err := udwBinary.ReadByteSliceWithUint32LenToBufW(relayConn, buf)
			if err != nil {
				udwLog.Log("[mka1nxd1mas1f]", err)
				return
			}
			err = vpnPacket.Decode(buf.GetBytes())
			if err != nil {
				udwLog.Log("[7ky9anc5uq]", err)
				return
			}
			switch vpnPacket.Cmd {
			case tachyonVpnProtocol.CmdForward:
				if vpnPacket.ClientIdReceiver == s.clientId {
					//TODO Server will use vpnPacket.ClientIdFrom to identify different TLS connections
					//TODO vpnPacket.ClientIdFrom should not be real Client's Id
					//TODO Relay Server could replace real Client's Id with fake one
					client := s.getOrNewClientFromRelayConn(vpnPacket.ClientIdSender)
					_, err := client.connRelaySide.Write(vpnPacket.Data) //TLS
					if err != nil {
						udwLog.Log("[dy11zv1eg6]", err)
					}
				} else {
					udwLog.Log("[vw9tm9rv2s] not forward to self", vpnPacket.ClientIdSender, "->", vpnPacket.ClientIdReceiver)
				}
			case tachyonVpnProtocol.CmdKeepAlive:
				s.relayConnKeepAliveChan<-binary.LittleEndian.Uint64(vpnPacket.Data)
			default:
				udwLog.Log("[d39e7d859m] Unexpected Cmd[", vpnPacket.Cmd, "]")
			}
			//if vpnPacket.Cmd == tachyonVpnProtocol.CmdForward {
			//	if vpnPacket.ClientIdReceiver == s.clientId {
			//		//TODO Server will use vpnPacket.ClientIdFrom to identify different TLS connections
			//		//TODO vpnPacket.ClientIdFrom should not be real Client's Id
			//		//TODO Relay Server could replace real Client's Id with fake one
			//		client := s.getOrNewClientFromRelayConn(vpnPacket.ClientIdSender)
			//		_, err := client.connRelaySide.Write(vpnPacket.Data) //TLS
			//		if err != nil {
			//			udwLog.Log("[dy11zv1eg6]", err)
			//		}
			//	} else {
			//		udwLog.Log("[vw9tm9rv2s] not forward to self", vpnPacket.ClientIdSender, "->", vpnPacket.ClientIdReceiver)
			//	}
			//} else {
			//	udwLog.Log("[d39e7d859m] Unexpected Cmd[", vpnPacket.Cmd, "]")
			//}
		}
	}()
	s.lock.Lock()
	s.relayConn = relayConn
	s.lock.Unlock()
	return nil
}

func (s *Server) relayConnKeepAliveThread() {
	s.relayConnKeepAliveChan = make(chan uint64, 10)
	go func() {
		i := uint64(0)
		vpnPacket := &tachyonVpnProtocol.VpnPacket{
			Cmd:            tachyonVpnProtocol.CmdKeepAlive,
			ClientIdSender: s.clientId,
		}
		bufW := udwBytes.NewBufWriter(nil)
		const timeout = time.Second * 2
		time.Sleep(timeout / 2)
		timer := time.NewTimer(timeout)
		for {
			bufW.Reset()
			vpnPacket.Data = vpnPacket.Data[:0]
			vpnPacket.Encode(bufW)
			bufW.WriteLittleEndUint64(i)
			err := udwBinary.WriteByteSliceWithUint32LenNoAllocV2(s.getRelayConn(), bufW.GetBytes())
			if err != nil {
				s.reconnectToRelay()
				continue
			}
			timer.Reset(timeout)
			select {
			case <-timer.C:
				udwLog.Log("[snc1hhr1ems1q] keepAlive timeout", i)
				s.reconnectToRelay()
			case _i := <-s.relayConnKeepAliveChan:
				if _i == i {
					i++
					time.Sleep(timeout / 2)
					continue
				}
				udwLog.Log("[snc1hhr1ems1q] keepAlive error: i not match, expect", i, "but got", _i)
				s.reconnectToRelay()
			}
		}
	}()
}

func (s *Server) reconnectToRelay() {
	s.lock.Lock()
	if s.relayConn != nil {
		_ = s.relayConn.Close()
	}
	s.lock.Unlock()
	for {
		udwLog.Log("[hjp1hbe1cbf1e] RECONNECT...")
		err := s.connectToRelay()
		if err != nil {
			udwLog.Log("[efv1bcw1ttm6] RECONNECT Failed", err)
			time.Sleep(time.Millisecond*500)
			continue
		}
		udwLog.Log("[ejq1knc2t8j] RECONNECT Succeed ✔")
		return
	}
}
