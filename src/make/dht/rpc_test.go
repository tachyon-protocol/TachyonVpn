package dht

import (
	"errors"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwTest"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestRpcNodeStore(t *testing.T) {
	node := newPeerNode(1)
	closeRpcServer := node.StartRpcServer()
	defer closeRpcServer()
	rNode := rpcNode{
		id: node.id,
		ip: "127.0.0.1",
	}
	const data = "Hyperion"
	key := hash([]byte(data))
	err := rNode.store([]byte(data))
	udwTest.Equal(err, nil)
	node.lock.RLock()
	v := node.keyMap[key]
	node.lock.RUnlock()
	udwTest.Equal(string(v), data)
}

func TestRpcNodeFindNode(t *testing.T) {
	node1 := newPeerNode(1)
	node2 := newPeerNode(2, node1.id)
	closeRpcServer := node2.StartRpcServer()
	defer closeRpcServer()
	rNode := rpcNode{
		id: node2.id,
		ip: "127.0.0.1",
	}
	closestId, err := rNode.findNode(1)
	udwErr.PanicIfError(err)
	udwTest.Equal(closestId, uint64(1))
}

func TestRpcNodeFindValue(t *testing.T) {
	const data = "Hyperion"
	key := hash([]byte(data))
	node1 := newPeerNode(key)
	node1.store([]byte(data))
	node2 := newPeerNode(2, node1.id)
	closeRpcServerNode2 := node2.StartRpcServer()
	rNode2 := rpcNode{
		id: node2.id,
		ip: "127.0.0.1",
	}
	closestId, value, err := rNode2.findValue(key)
	udwErr.PanicIfError(err)
	udwTest.Equal(closestId, node1.id)
	udwTest.Equal(value, nil)
	closeRpcServerNode2()

	closeRpcServerNode1 := node1.StartRpcServer()
	defer closeRpcServerNode1()
	rNodeClosest := rpcNode{
		id: closestId,
		ip: "127.0.0.1",
	}
	closestId, value, err = rNodeClosest.findValue(key)
	udwErr.PanicIfError(err)
	udwTest.Equal(closestId, node1.id)
	udwTest.Equal(string(value), data)
}

var responseTimeoutError = errors.New("timeout")

func sendBinaryToLocalRpcServer(request []byte, afterWrite func(conn net.Conn) (isReturn bool)) (response []byte, err error) {
	conn, err := net.Dial("udp", "127.0.0.1:"+strconv.Itoa(rpcPort))
	udwErr.PanicIfError(err)
	_, err = conn.Write(request)
	udwErr.PanicIfError(err)
	if afterWrite != nil {
		isReturn := afterWrite(conn)
		if isReturn {
			return
		}
	}
	buf := make([]byte, 2<<10)
	err = conn.SetDeadline(time.Now().Add(time.Millisecond * 300))
	udwErr.PanicIfError(err)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, responseTimeoutError
	}
	return buf[:n], nil
}

func TestRpcNodeClientError(t *testing.T) {
	node := newPeerNode(0)
	closeRpcServer := node.StartRpcServer()
	defer closeRpcServer()
	_, err := sendBinaryToLocalRpcServer([]byte("1"), nil)
	udwTest.Equal(err.Error(), responseTimeoutError.Error())
	_, err = sendBinaryToLocalRpcServer([]byte{0x02, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, nil)
	udwTest.Equal(err.Error(), responseTimeoutError.Error())
	_, err = sendBinaryToLocalRpcServer([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, nil)
	udwTest.Equal(err.Error(), responseTimeoutError.Error())
	_, err = sendBinaryToLocalRpcServer([]byte{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}, func(conn net.Conn) bool {
		_ = conn.Close()
		return true
	})
	udwTest.Equal(err, nil)
}
