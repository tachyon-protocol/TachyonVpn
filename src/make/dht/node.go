package dht

import (
	"encoding/binary"
	"github.com/tachyon-protocol/udw/udwCryptoSha3"
	"github.com/tachyon-protocol/udw/udwLog"
	"github.com/tachyon-protocol/udw/udwRand"
	"math"
	"sync"
)

type peerNode struct {
	id         uint64
	lock       sync.RWMutex
	keyMap     map[uint64][]byte
	knownNodes map[uint64]bool
}

func newPeerNode(id uint64, bootstrapNodeIds ...uint64) *peerNode {
	if id == 0 {
		id = udwRand.MustCryptoRandUint64()
	}
	n := &peerNode{
		id:         id,
		keyMap:     map[uint64][]byte{},
		knownNodes: map[uint64]bool{},
	}
	for _, id := range bootstrapNodeIds {
		n.knownNodes[id] = true
	}
	rpcInMemoryRegister(n)
	n.findNode(n.id)
	return n
}

func hash(v []byte) uint64 {
	digest := udwCryptoSha3.Sum224(v)
	return binary.LittleEndian.Uint64(digest[:])
}

func (node *peerNode) find(targetId uint64, isValue bool) (closestId uint64, value []byte) {
	closestId, value = node.findLocal(node.id, targetId, isValue)
	if isValue && value != nil {
		return closestId, value
	}
	if !isValue && targetId == closestId {
		return closestId, nil
	}
	for {
		closestNode := rpcInMemoryGetNode(closestId)
		_closestId, _value := closestNode.findLocal(node.id, targetId, isValue)
		if _closestId != node.id {
			node.lock.Lock()
			_, exist := node.knownNodes[_closestId]
			if !exist {
				if debugLog {
					udwLog.Log("[findNode]", node.id, "add new id", _closestId)
				}
				node.knownNodes[_closestId] = true
			}
			node.lock.Unlock()
		}
		if isValue && _value != nil {
			return _closestId, _value
		}
		if _closestId == closestId {
			return closestId, nil
		}
		closestId = _closestId
		if closestId == targetId {
			return targetId, nil
		}
		if closestId == node.id {
			return node.id, nil
		}
	}
}

func (node *peerNode) findLocal(callerId uint64, targetId uint64, isValue bool) (closestId uint64, value []byte) {
	if isValue {
		node.lock.RLock()
		v, exist := node.keyMap[targetId]
		node.lock.RUnlock()
		if exist {
			return targetId, v
		}
	}
	var min uint64 = math.MaxUint64
	var minId = node.id
	node.lock.RLock()
	for id := range node.knownNodes {
		_min := targetId ^ id
		if _min < min {
			min = _min
			minId = id
		}
	}
	node.lock.RUnlock()
	if callerId == node.id {
		return minId, nil
	}
	if callerId == targetId {
		node.lock.Lock()
		_, exist := node.knownNodes[callerId]
		if !exist {
			node.knownNodes[callerId] = true
			if debugLog {
				udwLog.Log("[findLocal]", node.id, "add new id", callerId)
			}
		}
		node.lock.Unlock()
	}
	if minId^targetId < node.id^targetId {
		if debugLog {
			udwLog.Log(node.id, "[findLocal]", targetId, "from caller", callerId, "closest:", minId)
		}
		return minId, nil
	}
	if debugLog {
		udwLog.Log("[findLocal]", node.id, "closest is itself, target", targetId)
	}
	return node.id, nil
}

//TODO ping

func (node *peerNode) store(v []byte) {
	node.lock.Lock()
	node.keyMap[hash(v)] = v
	node.lock.Unlock()
}

func (node *peerNode) findNode(targetId uint64) (closestId uint64) {
	closestId, _ = node.find(targetId, false)
	return closestId
}

func (node *peerNode) findValue(key uint64) (value []byte) {
	_, value = node.find(key, true)
	return value
}
