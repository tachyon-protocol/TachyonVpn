package tachyonVpnClient

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/tachyon-protocol/udw/udwBinary"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwLog"
	"net"
	"strconv"
	"tachyonVpnProtocol"
)

func (c *Client) connect() error {
	vpnConn, err := net.Dial("tcp", c.req.ServerIp+":"+strconv.Itoa(tachyonVpnProtocol.VpnPort))
	if err != nil {
		return errors.New("[w7syh9d1zgd] " + err.Error())
	}
	vpnConn = tls.Client(vpnConn, newInsecureClientTlsConfig())
	var (
		handshakeVpnPacket = tachyonVpnProtocol.VpnPacket{
			Cmd:            tachyonVpnProtocol.CmdHandshake,
			ClientIdSender: c.clientId,
			Data:           []byte(c.req.ServerTKey),
		}
		handshakeBuf = udwBytes.NewBufWriter(nil)
	)
	handshakeVpnPacket.Encode(handshakeBuf)
	err = udwBinary.WriteByteSliceWithUint32LenNoAllocV2(vpnConn, handshakeBuf.GetBytes())
	if err != nil {
		return errors.New("[52y73b9e89] " + err.Error())
	}
	serverType := "DIRECT"
	if c.req.IsRelay {
		serverType = "RELAY"
		var (
			connRelaySide, plain = tachyonVpnProtocol.NewInternalConnectionDual()
			relayConn            = vpnConn
		)
		vpnConn = tls.Client(plain, newInsecureClientTlsConfig())
		//read from relay conn, write to vpn conn
		go func() {
			var (
				buf       = udwBytes.NewBufWriter(nil)
				vpnPacket = &tachyonVpnProtocol.VpnPacket{}
			)
			for {
				buf.Reset()
				err := udwBinary.ReadByteSliceWithUint32LenToBufW(relayConn, buf)
				if err != nil {
					udwLog.Log("[wua1j5ps1pam] close 3 connections", err)
					_ = connRelaySide.Close()
					_ = plain.Close()
					_ = vpnConn.Close()
					return
				}
				err = vpnPacket.Decode(buf.GetBytes())
				if err != nil {
					udwLog.Log("[kj4v98z1fzc] close 3 connections", err)
					_ = connRelaySide.Close()
					_ = plain.Close()
					_ = vpnConn.Close()
					return
				}
				if vpnPacket.Cmd == tachyonVpnProtocol.CmdForward {
					_, err := connRelaySide.Write(vpnPacket.Data)
					if err != nil {
						udwLog.Log("[8gys171bvm] close 3 connections", err)
						_ = connRelaySide.Close()
						_ = plain.Close()
						_ = vpnConn.Close()
						return
					}
				} else {
					fmt.Println("[a3t7vfh1ms] Unexpected Cmd[", vpnPacket.Cmd, "]")
				}
			}
		}()
		//read from vpn conn, write to relay conn
		go func() {
			vpnPacket := &tachyonVpnProtocol.VpnPacket{
				Cmd:              tachyonVpnProtocol.CmdForward,
				ClientIdSender:   c.clientId,
				ClientIdReceiver: c.req.ExitServerClientId,
			}
			buf := make([]byte, 16*1024)
			bufW := udwBytes.NewBufWriter(nil)
			for {
				n, err := connRelaySide.Read(buf)
				if err != nil {
					udwLog.Log("[e9erq1bwd1] close 3 connections", err)
					_ = connRelaySide.Close()
					_ = plain.Close()
					_ = vpnConn.Close()
					return
				}
				vpnPacket.Data = buf[:n]
				bufW.Reset()
				vpnPacket.Encode(bufW)
				err = udwBinary.WriteByteSliceWithUint32LenNoAllocV2(relayConn, bufW.GetBytes())
				if err != nil {
					udwLog.Log("[n2cvu3w1cb] close 3 connections", err)
					_ = connRelaySide.Close()
					_ = plain.Close()
					_ = vpnConn.Close()
					return
				}
			}
		}()
		udwLog.Log("send handshake to ExitServer...")
		handshakeVpnPacket.ClientIdSender = c.clientIdToExitServer
		handshakeVpnPacket.Data = []byte(c.req.ExitServerTKey)
		handshakeBuf.Reset()
		handshakeVpnPacket.Encode(handshakeBuf)
		err = udwBinary.WriteByteSliceWithUint32LenNoAllocV2(vpnConn, handshakeBuf.GetBytes())
		if err != nil {
			return errors.New("[q3nwv1ebx1cd] " + err.Error())
		}
		udwLog.Log("sent handshake to ExitServer ✔")
	}
	fmt.Println("Connected to", serverType, "Server ✔")
	c.vpnConnLock.Lock()
	c.vpnConn = vpnConn
	c.vpnConnLock.Unlock()
	return nil
}

