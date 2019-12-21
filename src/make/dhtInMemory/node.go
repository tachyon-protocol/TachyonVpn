package dhtInMemory

import (
	"encoding/binary"
	"github.com/tachyon-protocol/udw/udwCryptoSha3"
	"github.com/tachyon-protocol/udw/udwLog"
	"github.com/tachyon-protocol/udw/udwRand"
	"math"
	"sync"
)

type node struct {
	id         uint64
	lock       sync.RWMutex
	keyMap     map[uint64][]byte
	knownNodes map[uint64]bool
}

func newNode(id uint64,bootstrapNodeIds ...uint64) *node {
	if id == 0 {
		id = udwRand.MustCryptoRandUint64()
	}
	n := &node{
		id:         id,
		keyMap:     map[uint64][]byte{},
		knownNodes: map[uint64]bool{},
	}
	for _, id := range bootstrapNodeIds {
		n.knownNodes[id] = true
	}
	rpcInMemoryRegister(n)
	return n
}

func hash(v []byte) uint64 {
	digest := udwCryptoSha3.Sum224(v)
	return binary.LittleEndian.Uint64(digest[:])
}

func (n *node) store(v []byte) {
	n.lock.Lock()
	n.keyMap[hash(v)] = v
	n.lock.Unlock()
}

func (n *node) findNode(targetId uint64) (closestId uint64) {
	closestId = n.findNodeLocal(n.id, targetId)
	if targetId == closestId {
		return targetId
	}
	for {
		closestNode := rpcInMemoryGetNode(closestId)
		_closestId := closestNode.findNodeLocal(n.id, targetId)
		if _closestId == 0 {
			return closestId
		}
		n.lock.Lock()
		n.knownNodes[_closestId] = true
		if debugLog {
			udwLog.Log("[findNode]", n.id, "add new id", _closestId)
		}
		n.lock.Unlock()
		if _closestId == closestId {
			return closestId
		}
		closestId = _closestId
		if closestId == targetId {
			return targetId
		}
	}
}

func (n *node) findNodeLocal(callerId uint64, targetId uint64) (closestId uint64) {
	var min uint64 = math.MaxUint64
	var minId uint64
	n.lock.RLock()
	for id := range n.knownNodes {
		_min := targetId ^ id
		if _min < min {
			minId = id
		}
	}
	n.lock.RUnlock()
	if debugLog {
		udwLog.Log("[findNodeLocal]", n.id, "closest:", minId, "target:", targetId)
	}
	if callerId != n.id && callerId == targetId {
		n.lock.Lock()
		n.knownNodes[callerId] = true
		n.lock.Unlock()
		if debugLog {
			udwLog.Log("[findNodeLocal]", n.id, "add new id", callerId)
		}
	}
	return minId
}

func (n *node) findValue(key uint64) (value []byte) {
	n.lock.RLock()
	v, exist := n.keyMap[key]
	n.lock.RUnlock()
	if exist {
		return v
	}
	var min uint64 = math.MaxUint64
	var minId uint64
	n.lock.RLock()
	for id := range n.knownNodes {
		_min := key ^ id
		if _min < min {
			minId = id
		}
	}
	n.lock.RUnlock()
	_node := rpcInMemoryGetNode(minId)
	return _node.findValue(key)
}
