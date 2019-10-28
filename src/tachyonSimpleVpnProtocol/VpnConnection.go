package tachyonSimpleVpnProtocol

import (
	"crypto/tls"
	"errors"
	"github.com/tachyon-protocol/udw/udwBinary"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwRand"
	"net"
	"sync"
)

func VpnConnectionDial(address string) (*VpnConnection, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		//noinspection SpellCheckingInspection
		return nil, errors.New("[ubb6g6pyjw]" + err.Error())
	}
	conn = tls.Client(conn, &tls.Config{
		ServerName:         udwRand.MustCryptoRandToReadableAlpha(5) + ".com",
		InsecureSkipVerify: true,
		NextProtos:         []string{"http/1.1", "h2"},
	})
	return &VpnConnection{
		rawConn: conn,
		bufR:    udwBytes.NewBufWriter(nil),
		bufW:    udwBytes.NewBufWriter(nil),
	}, nil
}

type VpnConnection struct {
	rawConn net.Conn
	bufW    *udwBytes.BufWriter
	bufR    *udwBytes.BufWriter
	locker  sync.Mutex
}

func (conn *VpnConnection) Read(packet *VpnPacket) error {
	conn.bufR.Reset()
	err := udwBinary.ReadByteSliceWithUint32LenToBufW(conn.rawConn, conn.bufR)
	if err != nil {
		return errors.New("[qz2qq4n43m]" + err.Error())
	}
	err = packet.Decode(conn.bufR.GetBytes())
	if err != nil {
		//noinspection SpellCheckingInspection
		return errors.New("[sjub59zv6y]" + err.Error())
	}
	return nil
}

func (conn *VpnConnection) Write(packet *VpnPacket) error {
	conn.locker.Lock()
	conn.bufW.Reset()
	packet.Encode(conn.bufW)
	err := udwBinary.WriteByteSliceWithUint32LenNoAllocV2(conn.rawConn, conn.bufW.GetBytes())
	conn.locker.Unlock()
	if err != nil {
		return errors.New("[v6ca32w5z7]" + err.Error())
	}
	return nil
}
