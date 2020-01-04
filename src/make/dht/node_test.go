package dht

import (
	"fmt"
	"github.com/tachyon-protocol/udw/udwTest"
	"testing"
)

func newTestNetwork() *peerNode {
	rpcInMemoryReset()
	node1 := newPeerNode(1)
	node2 := newPeerNode(2, node1.id)
	node3 := newPeerNode(3, node2.id)
	node4 := newPeerNode(4, node3.id)
	node5New := newPeerNode(5, node4.id)
	rpcInMemoryPrintlAllNode()
	return node5New
}

func TestRandomNetwork(t *testing.T) {
	for i := 0; i < 30; i++ {
		rpcInMemoryReset()
		node0 := newPeerNode(0)
		node1 := newPeerNode(0, node0.id)
		node2 := newPeerNode(0, node1.id)
		node3 := newPeerNode(0, node0.id)
		rpcInMemoryPrintlAllNode()

		const data = "Poseidon"
		key := hash([]byte(data))
		node3.store([]byte(data))
		closestIdList := node3.findNode(key)
		udwTest.Ok(len(closestIdList) != 0)
		storeNodeId := closestIdList[0]
		fmt.Println("will store in:", storeNodeId)
		storeNode := rpcInMemoryGetNode(storeNodeId)
		storeNode.store([]byte(data))

		v := node2.findValue(key)
		udwTest.Ok(string(v) == data)
	}
}

//func TestFindLoop(t *testing.T) {
//	rpcInMemoryReset()
//	node0 := newPeerNode(2013408581626216689)
//	node1 := newPeerNode(4246694672849243900, node0.id)
//	node2 := newPeerNode(6321635280997390418, node1.id)
//	node3 := newPeerNode(16775675729505829361, node0.id)
//	rpcInMemoryPrintlAllNode()
//
//	const data = "Poseidon"
//	key := hash([]byte(data))
//	node3.store([]byte(data))
//	storeNodeId := node3.findNode(key)
//	fmt.Println("will store in:", storeNodeId)
//	storeNode := rpcInMemoryGetNode(storeNodeId)
//	storeNode.store([]byte(data))
//
//	v := node2.findValue(key)
//	udwTest.Ok(string(v) == data)
//}
//
//func TestJoiningTheNetwork(t *testing.T) {
//	node5New := newTestNetwork()
//	closestId := node5New.findNode(node5New.id)
//	udwTest.Ok(closestId == node5New.id)
//}
//
//func TestFindNode(t *testing.T) {
//	node5New := newTestNetwork()
//	closestId := node5New.findNode(1)
//	udwTest.Equal(closestId, uint64(1))
//}
//
//func TestStoreAndFindValue(t *testing.T) {
//	node5 := newTestNetwork()
//	const data = "prometheus"
//	key := hash([]byte(data))
//	closestId := node5.findNode(key)
//	fmt.Println("will store in:", closestId)
//	closestNode := rpcInMemoryGetNode(closestId)
//	closestNode.store([]byte(data))
//
//	node3 := rpcInMemoryGetNode(3)
//	v := node3.findValue(key)
//	udwTest.Ok(string(v) == data)
//}
//
//func TestStoreAndFindValue2(t *testing.T) {
//	node5 := newTestNetwork()
//	const data = "Oceanus"
//	key := hash([]byte(data))
//	storeNodeId := node5.findNode(key)
//	fmt.Println("will store in:", storeNodeId)
//	storeNode := rpcInMemoryGetNode(storeNodeId)
//	storeNode.store([]byte(data))
//	v := storeNode.findValue(key)
//	udwTest.Ok(string(v) == data)
//}
//
//func TestStoreAndFindValue3(t *testing.T) {
//	node5 := newTestNetwork()
//	const data = "Hyperion"
//	key := hash([]byte(data))
//	storeNodeId := node5.findNode(key)
//	fmt.Println("will store in:", storeNodeId)
//	storeNode := rpcInMemoryGetNode(storeNodeId)
//	storeNode.store([]byte(data))
//	node3 := rpcInMemoryGetNode(3)
//	v := node3.findValue(key)
//	udwTest.Ok(string(v) == data)
//}
//
//func TestStoreAndFindValue4(t *testing.T) {
//	node5 := newTestNetwork()
//	const data = "Hyperion"
//	key := hash([]byte(data))
//	storeNodeId := node5.findNode(key)
//	fmt.Println("will store in:", storeNodeId)
//	storeNode := rpcInMemoryGetNode(storeNodeId)
//	storeNode.store([]byte(data))
//	node6 := newPeerNode(6) //isolated
//	v := node6.findValue(key)
//	udwTest.Ok(v == nil)
//}
