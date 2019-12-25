package dhtInMemory

import (
	"fmt"
	"github.com/tachyon-protocol/udw/udwTest"
	"testing"
)

func newTestNetwork() *node {
	node1 := newNode(1)
	node2 := newNode(2, node1.id)
	node3 := newNode(3, node2.id)
	node4 := newNode(4, node3.id)
	node5New := newNode(5, node4.id)
	return node5New
}

func TestJoiningTheNetwork(t *testing.T) {
	node5New := newTestNetwork()
	closestId := node5New.findNode(node5New.id)
	udwTest.Ok(closestId != node5New.id)
	closestId = node5New.findNode(node5New.id)
	udwTest.Ok(closestId == node5New.id)
}

func TestFindNode(t *testing.T) {
	node5New := newTestNetwork()
	udwTest.Ok(len(node5New.knownNodes) == 1)
	closestId := node5New.findNode(1)
	udwTest.Ok(closestId == 1)
	udwTest.Ok(len(node5New.knownNodes) == 4)
}

func TestStoreAndFindValue(t *testing.T) {
	node5 := newTestNetwork()
	const data = "prometheus"
	key := hash([]byte(data))
	closestId := node5.findNode(key)
	fmt.Println("will store in:", closestId)
	closestNode := rpcInMemoryGetNode(closestId)
	closestNode.store([]byte(data))

	node3 := rpcInMemoryGetNode(3)
	v := node3.findValue(key)
	udwTest.Ok(string(v) == data)
}
