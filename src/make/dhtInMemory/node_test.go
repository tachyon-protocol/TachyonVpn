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

//BUG
//func TestRandomNetwork(t *testing.T) {
//	for i := 0; i < 1; i++ {
//		rpcInMemoryReset()
//		node0 := newNode(0)
//		node1 := newNode(0, node0.id)
//		node2 := newNode(0, node1.id)
//		node3 := newNode(0, node0.id)
//		rpcInMemoryPrintlAllNode()
//
//		const data = "Poseidon"
//		key := hash([]byte(data))
//		node3.store([]byte(data))
//		storeNodeId := node3.FindNode(key)
//		fmt.Println("will store in:", storeNodeId)
//		storeNode := rpcInMemoryGetNode(storeNodeId)
//		storeNode.store([]byte(data))
//
//		v := node2.FindValue(key)
//		udwTest.Ok(string(v) == data)
//	}
//}

func TestFindLoop(t *testing.T) {
	rpcInMemoryReset()
	node0 := newNode(2013408581626216689)
	node1 := newNode(4246694672849243900, node0.id)
	node2 := newNode(6321635280997390418, node1.id)
	node3 := newNode(16775675729505829361, node0.id)
	rpcInMemoryPrintlAllNode()

	const data = "Poseidon"
	key := hash([]byte(data))
	node3.store([]byte(data))
	storeNodeId := node3.FindNode(key)
	fmt.Println("will store in:", storeNodeId)
	storeNode := rpcInMemoryGetNode(storeNodeId)
	storeNode.store([]byte(data))

	v := node2.FindValue(key)
	udwTest.Ok(string(v) == data)
}

func TestJoiningTheNetwork(t *testing.T) {
	node5New := newTestNetwork()
	closestId := node5New.FindNode(node5New.id)
	udwTest.Ok(closestId == node5New.id)
}

func TestFindNode(t *testing.T) {
	node5New := newTestNetwork()
	closestId := node5New.FindNode(1)
	udwTest.Equal(closestId, uint64(1))
}

func TestStoreAndFindValue(t *testing.T) {
	node5 := newTestNetwork()
	const data = "prometheus"
	key := hash([]byte(data))
	closestId := node5.FindNode(key)
	fmt.Println("will store in:", closestId)
	closestNode := rpcInMemoryGetNode(closestId)
	closestNode.store([]byte(data))

	node3 := rpcInMemoryGetNode(3)
	v := node3.FindValue(key)
	udwTest.Ok(string(v) == data)
}

func TestStoreAndFindValue2(t *testing.T) {
	node5 := newTestNetwork()
	const data = "Oceanus"
	key := hash([]byte(data))
	storeNodeId := node5.FindNode(key)
	fmt.Println("will store in:", storeNodeId)
	storeNode := rpcInMemoryGetNode(storeNodeId)
	storeNode.store([]byte(data))
	v := storeNode.FindValue(key)
	udwTest.Ok(string(v) == data)
}

func TestStoreAndFindValue3(t *testing.T) {
	node5 := newTestNetwork()
	const data = "Hyperion"
	key := hash([]byte(data))
	storeNodeId := node5.FindNode(key)
	fmt.Println("will store in:", storeNodeId)
	storeNode := rpcInMemoryGetNode(storeNodeId)
	storeNode.store([]byte(data))
	node3 := rpcInMemoryGetNode(3)
	v := node3.FindValue(key)
	udwTest.Ok(string(v) == data)
}

func TestStoreAndFindValue4(t *testing.T) {
	node5 := newTestNetwork()
	const data = "Hyperion"
	key := hash([]byte(data))
	storeNodeId := node5.FindNode(key)
	fmt.Println("will store in:", storeNodeId)
	storeNode := rpcInMemoryGetNode(storeNodeId)
	storeNode.store([]byte(data))
	node6 := newNode(6) //isolated
	v := node6.FindValue(key)
	udwTest.Ok(v == nil)
}
