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

func NewInternalConnectionDual() (rBwA net.Conn, rAwB net.Conn) {
	a := &internalConnectionSingle{
		pipe:      udwChan.MakeChanBytes(1 << 10),
		debugName: "A",
	}
	b := &internalConnectionSingle{
		pipe:      udwChan.MakeChanBytes(1 << 10),
		debugName: "B",
	}
	return &internalConnectionPeer{
			readConn:  b,
			writeConn: a,
		}, &internalConnectionPeer{
			readConn:  a,
			writeConn: b,
		}
}

type internalConnectionPeer struct {
	readConn  *internalConnectionSingle
	writeConn *internalConnectionSingle
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
	debugName string
	pipe      *udwChan.ChanBytes
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
		data, isClose := conn.pipe.Receive()
		if isClose {
			conn.lockerR.Unlock()
			return 0, io.ErrClosedPipe
		}
		if DebugInternalConnection {
			fmt.Println(conn.debugName, "receive", len(data))
		}
		conn.bufR.Write_(data)
		conn.readIndex = 0
	}
	_buf := conn.bufR.GetBytes()
	n = copy(buf, _buf[conn.readIndex:])
	if DebugInternalConnection {
		fmt.Println(conn.debugName, "read", n)
	}
	conn.readIndex += n
	conn.lockerR.Unlock()
	return n, nil
}

func (conn *internalConnectionSingle) Write(buf []byte) (n int, err error) {
	//const size = 8000
	//start := 0
	//for {
	//	if start >= len(buf) {
	//		return len(buf), nil
	//	}
	//	end := start + size
	//	if len(buf) < end {
	//		end = len(buf)
	//	}
	//	time.Sleep(time.Millisecond )
	//	isClose := conn.pipe.Send(buf[start:end])
	//	start += size
	//	if isClose {
	//		return 0, io.ErrClosedPipe
	//	}
	//}

	if DebugInternalConnection {
		records := tlsPacketDebugger.GetRecordList(buf)
		fmt.Println(conn.debugName, "write", len(buf))
		for _, r := range records {
			fmt.Println("	", r.ContentType, r.Length)
			for _, p := range r.ProtocolList {
				fmt.Println("		", p.HandshakeType, p.Length)
			}
		}
	}

	//const bufSize = 100
	//conn.lockerW.Lock()
	//if conn.bufW == nil {
	//	conn.bufW = udwBytes.NewBufWriter(nil)
	//}
	//conn.bufW.Write_(buf)
	//if conn.bufW.GetLen() < bufSize {
	//	conn.lockerW.Unlock()
	//	return len(buf), nil
	//}
	//isClose := conn.pipe.Send(conn.bufW.GetBytes())
	//conn.bufW.Reset()
	//conn.lockerW.Unlock()
	//if isClose {
	//	return 0, io.ErrClosedPipe
	//}
	//return len(buf), nil

	//TODO
	_buf := make([]byte, len(buf))
	copy(_buf, buf)
	isClose := conn.pipe.Send(_buf)
	if isClose {
		return 0, io.ErrClosedPipe
	}
	return len(buf), nil
}

func (conn *internalConnectionSingle) Close() error {
	conn.pipe.Close()
	return nil
}
