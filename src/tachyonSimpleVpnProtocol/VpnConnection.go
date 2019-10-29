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

type VpnConnection struct {
	req       VpnConnectionNewReq
	locker    sync.Mutex
	buf       *udwBytes.BufWriter
	vpnPacket *VpnPacket
}

type VpnConnectionNewReq struct {
	ClientId uint64
	IsRelay  bool
	RawConn  net.Conn
}

func VpnConnectionNew(req VpnConnectionNewReq) *VpnConnection {
	 req.RawConn = tls.Client(req.RawConn, &tls.Config{
		ServerName:         udwRand.MustCryptoRandToReadableAlpha(5) + ".com",
		InsecureSkipVerify: true,
		NextProtos:         []string{"http/1.1", "h2"},
	})
	return &VpnConnection{
		req: req,
	}
}

//TODO not complete implement of stream
func (conn *VpnConnection) Read(buf []byte) (n int, err error) {
	conn.buf.Reset()
	conn.locker.Lock()
	if conn.buf == nil {
		conn.buf = udwBytes.NewBufWriter(nil)
	}
	err = udwBinary.ReadByteSliceWithUint32LenToBufW(conn.req.RawConn, conn.buf)
	if err != nil {
		conn.locker.Unlock()
		return 0, errors.New("[qz2qq4n43m]" + err.Error())
	}
	if conn.vpnPacket ==  nil {
		conn.vpnPacket = &VpnPacket{}
	}
	err = conn.vpnPacket.Decode(conn.buf.GetBytes())
	if err != nil {
		//noinspection SpellCheckingInspection
		return 0, errors.New("[sjub59zv6y]" + err.Error())
	}
	n = copy(buf, conn.vpnPacket.Data)
	return n, nil
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
