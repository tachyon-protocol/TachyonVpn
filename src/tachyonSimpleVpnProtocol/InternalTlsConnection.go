package tachyonSimpleVpnProtocol

import (
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwChan"
	"io"
	"net"
	"sync"
	"time"
)

//left  cipher -> tls -> plain
type internalConnection struct {
	pipe      *udwChan.ChanBytes
	locker    sync.Mutex
	buf       *udwBytes.BufWriter
	readIndex int
}

func (conn *internalConnection) LocalAddr() net.Addr {
	panic("implement me")
}

func (conn *internalConnection) RemoteAddr() net.Addr {
	panic("implement me")
}

func (conn *internalConnection) Read(buf []byte) (n int, err error) {
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

func (conn *internalConnection) Write(buf []byte) (n int, err error) {
	conn.pipe.Send(buf)
	return len(buf), nil
}

func (conn *internalConnection) Close() error {
	conn.pipe.Close()
	return nil
}

func (conn *internalConnection) SetDeadline(t time.Time) error {
	return nil
}

func (conn *internalConnection) SetReadDeadline(t time.Time) error {
	return nil
}

func (conn *internalConnection) SetWriteDeadline(t time.Time) error {
	return nil
}
