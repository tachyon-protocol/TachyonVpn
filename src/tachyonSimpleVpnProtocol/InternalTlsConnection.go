package tachyonSimpleVpnProtocol

import (
	"github.com/tachyon-protocol/udw/udwBytes"
	"sync"
)

//left  cipher -> tls -> plain
type internalConnection struct {
	pipe      chan []byte
	locker    sync.Mutex
	buf       *udwBytes.BufWriter
	readIndex int
}

func (conn *internalConnection) Read(buf []byte) (n int, err error) {
	conn.locker.Lock()
	if conn.readIndex == conn.buf.GetLen() {
		conn.buf.Reset()
		conn.buf.Write_(<-conn.pipe)
		conn.readIndex = 0
	}
	_buf := conn.buf.GetBytes()
	n = copy(buf, _buf[conn.readIndex:])
	conn.readIndex += n
	conn.locker.Unlock()
	return n, nil
}

func (conn *internalConnection) Write(buf []byte) (n int, err error) {
	conn.pipe <- buf
	return len(buf), nil
}
