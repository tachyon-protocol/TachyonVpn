package tachyonSimpleVpnProtocol

import (
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwChan"
	"io"
	"net"
	"sync"
	"time"
)

func NewInternalConnectionDual() (cipherConn net.Conn, plainConn net.Conn) {
	left := &internalConnectionSingle{
		pipe: udwChan.MakeChanBytes(0),
		buf:  udwBytes.NewBufWriter(nil),
	}
	right := &internalConnectionSingle{
		pipe: udwChan.MakeChanBytes(0),
		buf:  udwBytes.NewBufWriter(nil),
	}
	return &internalConnectionPeer{
			readConn:  right,
			writeConn: left,
		}, &internalConnectionPeer{
			readConn:  left,
			writeConn: right,
		}
}

type internalConnectionPeer struct {
	readConn  *internalConnectionSingle
	writeConn *internalConnectionSingle
}

func (conn *internalConnectionPeer) LocalAddr() net.Addr {
	panic("implement me")
}

func (conn *internalConnectionPeer) RemoteAddr() net.Addr {
	panic("implement me")
}

func (conn *internalConnectionPeer) SetDeadline(t time.Time) error {
	return nil
}

func (conn *internalConnectionPeer) SetReadDeadline(t time.Time) error {
	return nil
}

func (conn *internalConnectionPeer) SetWriteDeadline(t time.Time) error {
	return nil
}

func (conn *internalConnectionPeer) Read(buf []byte) (n int, err error) {
	return conn.readConn.Read(buf)
}

func (conn *internalConnectionPeer) Write(buf []byte) (n int, err error) {
	return conn.writeConn.Write(buf)
}

func (conn *internalConnectionPeer) Close() (err error) {
	_ = conn.readConn.Close()
	_ = conn.writeConn.Close()
	return nil
}

type internalConnectionSingle struct {
	pipe      *udwChan.ChanBytes
	locker    sync.Mutex
	buf       *udwBytes.BufWriter
	readIndex int
}

func (conn *internalConnectionSingle) Read(buf []byte) (n int, err error) {
	conn.locker.Lock()
	if conn.readIndex == conn.buf.GetLen() {
		conn.buf.Reset()
		data, isClose := conn.pipe.Receive()
		if isClose {
			conn.locker.Unlock()
			return 0, io.ErrClosedPipe
		}
		conn.buf.Write_(data)
		conn.readIndex = 0
	}
	_buf := conn.buf.GetBytes()
	n = copy(buf, _buf[conn.readIndex:])
	conn.readIndex += n
	conn.locker.Unlock()
	return n, nil
}

func (conn *internalConnectionSingle) Write(buf []byte) (n int, err error) {
	isClose := conn.pipe.Send(buf)
	if isClose {
		return 0, io.ErrClosedPipe
	}
	return len(buf), nil
}

func (conn *internalConnectionSingle) Close() error {
	conn.pipe.Close()
	return nil
}
