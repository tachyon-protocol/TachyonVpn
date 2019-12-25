package dhtInMemory

import (
	"fmt"
	"github.com/tachyon-protocol/udw/udwTest"
	"testing"
)

func newTestNetwork() *node {
	rpcInMemoryReset()
	node1 := newNode(1)
	node2 := newNode(2, node1.id)
	node3 := newNode(3, node2.id)
	node4 := newNode(4, node3.id)
	node5New := newNode(5, node4.id)
	rpcInMemoryPrintlAllNode()
	return node5New
}

func TestJoiningTheNetwork(t *testing.T) {
	node5New := newTestNetwork()
	closestId := node5New.findNode(node5New.id)
	udwTest.Ok(closestId == node5New.id)
}

func TestFindNode(t *testing.T) {
	node5New := newTestNetwork()
	closestId := node5New.findNode(1)
	udwTest.Equal(closestId, uint64(1))
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

func TestStoreAndFindValue2(t *testing.T) {
	node5 := newTestNetwork()
	const data = "Oceanus"
	key := hash([]byte(data))
	storeNodeId := node5.findNode(key)
	fmt.Println("will store in:", storeNodeId)
	storeNode := rpcInMemoryGetNode(storeNodeId)
	storeNode.store([]byte(data))
	v := storeNode.findValue(key)
	udwTest.Ok(string(v) == data)
}

func TestStoreAndFindValue3(t *testing.T) {
	node5 := newTestNetwork()
	const data = "Hyperion"
	key := hash([]byte(data))
	storeNodeId := node5.findNode(key)
	fmt.Println("will store in:", storeNodeId)
	storeNode := rpcInMemoryGetNode(storeNodeId)
	storeNode.store([]byte(data))
	node3 := rpcInMemoryGetNode(3)
	v := node3.findValue(key)
	udwTest.Ok(string(v) == data)
}

func TestStoreAndFindValue4(t *testing.T) {
	node5 := newTestNetwork()
	const data = "Hyperion"
	key := hash([]byte(data))
	storeNodeId := node5.findNode(key)
	fmt.Println("will store in:", storeNodeId)
	storeNode := rpcInMemoryGetNode(storeNodeId)
	storeNode.store([]byte(data))
	node6 := newNode(6) //isolated
	v := node6.findValue(key)
	udwTest.Ok(v == nil)
}
