package dht

import (
	"github.com/tachyon-protocol/udw/udwTest"
	"testing"
)

func TestRpcNode (t *testing.T){
	node := newPeerNode(1)
	go node.StartRpcServer()
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
	udwTest.Equal(string(v),data)
}

