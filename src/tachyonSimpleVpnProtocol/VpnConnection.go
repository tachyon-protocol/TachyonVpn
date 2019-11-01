package tachyonSimpleVpnProtocol

import (
	"crypto/tls"
	"errors"
	"github.com/tachyon-protocol/udw/udwBinary"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwNet"
	"github.com/tachyon-protocol/udw/udwRand"
	"net"
	"sync"
)

type VpnConnection struct {
	req       VpnConnectionNewReq
	locker    sync.Mutex
	buf       *udwBytes.BufWriter
	vpnPacket *VpnPacket
}

type VpnConnectionNewReq struct {
	ClientIdFrom      uint64
	ClientIdForwardTo uint64
	IsRelay           bool
	RawConn           net.Conn
}

func VpnConnectionNew(req VpnConnectionNewReq) net.Conn {
	req.RawConn = tls.Client(req.RawConn, &tls.Config{
		ServerName:         udwRand.MustCryptoRandToReadableAlpha(5) + ".com",
		InsecureSkipVerify: true,
		NextProtos:         []string{"http/1.1", "h2"},
	})
	return udwNet.NewRwcOverConn(&VpnConnection{
		req: req,
	}, req.RawConn)
}

//TODO not complete implement of stream
func (conn *VpnConnection) Read(buf []byte) (n int, err error) {
	conn.locker.Lock()
	if conn.buf == nil {
		conn.buf = udwBytes.NewBufWriter(nil)
	}
	conn.buf.Reset()
	err = udwBinary.ReadByteSliceWithUint32LenToBufW(conn.req.RawConn, conn.buf)
	if err != nil {
		conn.locker.Unlock()
		return 0, errors.New("[qz2qq4n43m]" + err.Error())
	}
	if conn.vpnPacket == nil {
		conn.vpnPacket = &VpnPacket{}
	}
	err = conn.vpnPacket.Decode(conn.buf.GetBytes())
	if err != nil {
		conn.locker.Unlock()
		//noinspection SpellCheckingInspection
		return 0, errors.New("[sjub59zv6y]" + err.Error())
	}
	n = copy(buf, conn.vpnPacket.Data)
	conn.locker.Unlock()
	return n, nil
}

//TODO not complete implement of stream
func (conn *VpnConnection) Write(buf []byte) (n int, err error) {
	conn.locker.Lock()
	if conn.buf == nil {
		conn.buf = udwBytes.NewBufWriter(nil)
	}
	conn.buf.Reset()
	if conn.vpnPacket == nil {
		conn.vpnPacket = &VpnPacket{}
	}
	conn.vpnPacket.ClientIdSender = conn.req.ClientIdFrom
	conn.vpnPacket.Cmd = CmdData
	if conn.req.IsRelay {
		conn.vpnPacket.ClientIdReceiver = conn.req.ClientIdForwardTo
		conn.vpnPacket.Cmd = CmdForward
	}
	n = copy(conn.vpnPacket.Data, buf)
	conn.vpnPacket.Encode(conn.buf)
	err = udwBinary.WriteByteSliceWithUint32LenNoAllocV2(conn.req.RawConn, conn.buf.GetBytes())
	conn.locker.Unlock()
	if err != nil {
		return 0, errors.New("[v6ca32w5z7]" + err.Error())
	}
	return 0, nil
}

func (conn *VpnConnection) Close() error {
	return conn.req.RawConn.Close()
}
