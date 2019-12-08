package tachyonVpnProtocol

import (
	"fmt"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwChan"
	"io"
	"net"
	"sync"
	"time"
	"tlsPacketDebugger"
)

const DebugInternalConnection = false

func NewInternalConnectionDual(closeFnLeft func(), closeFnRight func()) (rBwA net.Conn, rAwB net.Conn) {
	a := &internalConnectionSingle{
		pipe:      udwChan.MakeChan(256),
		debugName: "A",
	}
	b := &internalConnectionSingle{
		pipe:      udwChan.MakeChan(256),
		debugName: "B",
	}
	return &internalConnectionPeer{
			readConn:  b,
			writeConn: a,
			closeFn:   closeFnLeft,
		}, &internalConnectionPeer{
			readConn:  a,
			writeConn: b,
			closeFn:   closeFnRight,
		}
}

type internalConnectionPeer struct {
	readConn  *internalConnectionSingle
	writeConn *internalConnectionSingle
	closeFn   func()
}

func (conn *internalConnectionPeer) LocalAddr() net.Addr {
	return &net.TCPAddr{
		IP:   net.ParseIP("1.1.1.1"),
		Port: 1111,
		Zone: "",
	}
}

func (conn *internalConnectionPeer) RemoteAddr() net.Addr {
	return &net.TCPAddr{
		IP:   net.ParseIP("1.1.1.1"),
		Port: 2222,
		Zone: "",
	}
}

func (conn *internalConnectionPeer) SetDeadline(_ time.Time) error {
	return nil
}

func (conn *internalConnectionPeer) SetReadDeadline(_ time.Time) error {
	return nil
}

func (conn *internalConnectionPeer) SetWriteDeadline(_ time.Time) error {
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
	if conn.closeFn != nil {
		conn.closeFn()
	}
	return nil
}

type internalConnectionSingle struct {
	debugName string
	pipe      *udwChan.Chan
	bufPool   udwBytes.BufWriterPool
	lockerR   sync.Mutex
	bufR      *udwBytes.BufWriter
	readIndex int
}

func (conn *internalConnectionSingle) Read(buf []byte) (n int, err error) {
	conn.lockerR.Lock()
	if conn.bufR == nil {
		conn.bufR = udwBytes.NewBufWriter(nil)
	}
	if conn.readIndex == conn.bufR.GetLen() {
		conn.bufR.Reset()
		_bufI, isClose := conn.pipe.Receive()
		if isClose {
			conn.lockerR.Unlock()
			return 0, io.ErrClosedPipe
		}
		_buf := _bufI.(*udwBytes.BufWriter)
		conn.bufR.Write_(_buf.GetBytes())
		conn.bufPool.Put(_buf)
		conn.readIndex = 0
	}
	n = copy(buf, conn.bufR.GetBytes()[conn.readIndex:])
	if DebugInternalConnection {
		fmt.Println(conn.debugName, "read", n)
	}
	conn.readIndex += n
	conn.lockerR.Unlock()
	return n, nil
}

func (conn *internalConnectionSingle) Write(buf []byte) (n int, err error) {
	if DebugInternalConnection {
		tlsPacketDebugger.Dump(conn.debugName, buf)
	}
	_bufW := conn.bufPool.GetAndCloneFromByteSlice(buf)
	isClose := conn.pipe.Send(_bufW)
	if isClose {
		return 0, io.ErrClosedPipe
	}
	return len(buf), nil
}

func (conn *internalConnectionSingle) Close() error {
	conn.pipe.Close()
	return nil
}
