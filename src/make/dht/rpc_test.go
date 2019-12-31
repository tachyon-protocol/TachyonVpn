package dht

import (
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwTest"
	"testing"
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
