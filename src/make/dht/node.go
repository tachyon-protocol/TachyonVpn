package dht

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

func newNode(id uint64, bootstrapNodeIds ...uint64) *node {
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
	n.findNode(n.id)
	return n
}

func hash(v []byte) uint64 {
	digest := udwCryptoSha3.Sum224(v)
	return binary.LittleEndian.Uint64(digest[:])
}

func (n *node) find(targetId uint64, isValue bool) (closestId uint64, value []byte) {
	closestId, value = n.findLocal(n.id, targetId, isValue)
	if isValue && value != nil {
		return closestId, value
	}
	if !isValue && targetId == closestId {
		return closestId, nil
	}
	for {
		closestNode := rpcInMemoryGetNode(closestId)
		_closestId, _value := closestNode.findLocal(n.id, targetId, isValue)
		if _closestId != n.id {
			n.lock.Lock()
			_, exist := n.knownNodes[_closestId]
			if !exist {
				if debugLog {
					udwLog.Log("[findNode]", n.id, "add new id", _closestId)
				}
				n.knownNodes[_closestId] = true
			}
			n.lock.Unlock()
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
		if closestId == n.id {
			return n.id, nil
		}
	}
}

func (n *node) findLocal(callerId uint64, targetId uint64, isValue bool) (closestId uint64, value []byte) {
	if isValue {
		n.lock.RLock()
		v, exist := n.keyMap[targetId]
		n.lock.RUnlock()
		if exist {
			return targetId, v
		}
	}
	var min uint64 = math.MaxUint64
	var minId = n.id
	n.lock.RLock()
	for id := range n.knownNodes {
		_min := targetId ^ id
		if _min < min {
			min = _min
			minId = id
		}
	}
	n.lock.RUnlock()
	if callerId == n.id {
		return minId, nil
	}
	if callerId == targetId {
		n.lock.Lock()
		_, exist := n.knownNodes[callerId]
		if !exist {
			n.knownNodes[callerId] = true
			if debugLog {
				udwLog.Log("[findLocal]", n.id, "add new id", callerId)
			}
		}
		n.lock.Unlock()
	}
	if minId^targetId < n.id^targetId {
		if debugLog {
			udwLog.Log(n.id, "[findLocal]", targetId, "from caller", callerId, "closest:", minId)
		}
		return minId, nil
	}
	if debugLog {
		udwLog.Log("[findLocal]", n.id, "closest is itself, target", targetId)
	}
	return n.id, nil
}

//TODO ping

func (n *node) store(v []byte) {
	n.lock.Lock()
	n.keyMap[hash(v)] = v
	n.lock.Unlock()
}

func (n *node) findNode(targetId uint64) (closestId uint64) {
	closestId, _ = n.find(targetId, false)
	return closestId
}

func (n *node) findValue(key uint64) (value []byte) {
	_, value = n.find(key, true)
	return value
}
