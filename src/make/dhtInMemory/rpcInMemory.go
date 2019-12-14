package dhtInMemory

import "sync"

var (
	gRpcInMemoryNodeMapLock sync.RWMutex
	gRpcInMemoryNodeMap     = map[uint64]*node{}
)

func rpcInMemoryGetNode(id uint64) *node {
	gRpcInMemoryNodeMapLock.RLock()
	n := gRpcInMemoryNodeMap[id]
	gRpcInMemoryNodeMapLock.RUnlock()
	return n
}

func rpcInMemoryRegister(n *node) {
	gRpcInMemoryNodeMapLock.Lock()
	gRpcInMemoryNodeMap[n.id] = n
	gRpcInMemoryNodeMapLock.Unlock()
}
