package dht

import (
	"fmt"
	"sort"
	"sync"
)

var (
	gRpcInMemoryNodeMapLock sync.RWMutex
	gRpcInMemoryNodeMap     = map[uint64]*peerNode{}
)

func rpcInMemoryReset() {
	gRpcInMemoryNodeMapLock.Lock()
	gRpcInMemoryNodeMap = map[uint64]*peerNode{}
	gRpcInMemoryNodeMapLock.Unlock()
}

func rpcInMemoryGetNode(id uint64) *peerNode {
	gRpcInMemoryNodeMapLock.RLock()
	n := gRpcInMemoryNodeMap[id]
	gRpcInMemoryNodeMapLock.RUnlock()
	return n
}

func rpcInMemoryRegister(n *peerNode) {
	gRpcInMemoryNodeMapLock.Lock()
	gRpcInMemoryNodeMap[n.id] = n
	gRpcInMemoryNodeMapLock.Unlock()
}

func rpcInMemoryPrintAllNode() {
	fmt.Println("---------------Network---------------")
	gRpcInMemoryNodeMapLock.RLock()
	var ids []uint64
	for id := range gRpcInMemoryNodeMap {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})
	for _, id := range ids {
		node := gRpcInMemoryNodeMap[id]
		node.lock.RLock()
		fmt.Println(node.id, "k-buckets:")
		for cps, m := range node.kBuckets {
			if m == nil {
				continue
			}
			fmt.Println("	", "cps", cps)
			for id := range m {
				fmt.Println("		", id)
			}
		}
		node.lock.RUnlock()
	}
	gRpcInMemoryNodeMapLock.RUnlock()
	fmt.Println("-------------------------------------")
}
