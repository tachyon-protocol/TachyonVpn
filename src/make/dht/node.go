package dht

import (
	"encoding/binary"
	"github.com/tachyon-protocol/udw/udwCryptoSha3"
	"github.com/tachyon-protocol/udw/udwLog"
	"github.com/tachyon-protocol/udw/udwMap"
	"github.com/tachyon-protocol/udw/udwRand"
	"github.com/tachyon-protocol/udw/udwSort"
	"math"
	"sync"
)

type peerNode struct {
	id       uint64
	lock     sync.RWMutex
	keyMap   map[uint64][]byte
	kBuckets [64]map[uint64]bool
}

func newPeerNode(id uint64, bootstrapNodeIds ...uint64) *peerNode {
	if id == 0 {
		id = udwRand.MustCryptoRandUint64()
	}
	n := &peerNode{
		id:       id,
		keyMap:   map[uint64][]byte{},
		kBuckets: [64]map[uint64]bool{},
	}
	for _, id := range bootstrapNodeIds {
		index := sizeOfCommonPrefix(n.id, id)
		m := n.kBuckets[index]
		if m == nil {
			m = map[uint64]bool{}
		}
		m[id] = true
		n.kBuckets[index] = m
	}
	rpcInMemoryRegister(n)
	n.findNode(n.id)
	return n
}

func sizeOfCommonPrefix(a,b uint64) int {
	pl := 64
	for {
		if a == b {
			break
		}
		a >>= 1
		b >>= 1
		pl--
	}
	return pl
}

func hash(v []byte) uint64 {
	digest := udwCryptoSha3.Sum224(v)
	return binary.LittleEndian.Uint64(digest[:])
}

func (node *peerNode) find(targetId uint64, isValue bool) (closestK map[uint64]bool, value []byte) {
	closestId, value = node.findLocal(node.id, targetId, isValue)
	if isValue && value != nil {
		return nil, value
	}
	if !isValue && targetId == closestId {
		return nil, nil
	}
	for {
		closestNode := rpcInMemoryGetNode(closestId)
		_closestId, _value := closestNode.findLocal(node.id, targetId, isValue)
		if _closestId != node.id {
			node.lock.Lock()
			_, exist := node.kBuckets[_closestId]
			if !exist {
				if debugDhtLog {
					udwLog.Log("[findNode]", node.id, "add new id", _closestId)
				}
				node.kBuckets[_closestId] = true
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

func (node *peerNode) findLocal(callerId uint64, targetId uint64, isValue bool) (closestKMap map[uint64]uint64, value []byte) {
	if isValue {
		node.lock.RLock()
		v, exist := node.keyMap[targetId]
		node.lock.RUnlock()
		if exist {
			return nil, v
		}
	}
	node.lock.RLock()
	for _, km := range node.kBuckets {
		for id := range km {
			distance := targetId ^ id
			if distance < maxDistance {
				if closestKMap == nil {
					closestKMap = map[uint64]uint64{}
				}
				closestKMap[id] = distance
				if len(closestKMap) > k {
					delete(closestKMap, maxId)
					for id := range closestKMap {
					}
				}
			}
			//if distance < min {
			//	min = distance
			//	minId = id
			//}
		}
	}
	node.lock.RUnlock()
	if callerId == node.id {
		return minId, nil
	}
	if callerId == targetId {
		node.lock.Lock()
		_, exist := node.kBuckets[callerId]
		if !exist {
			node.kBuckets[callerId] = true
			if debugDhtLog {
				udwLog.Log("[findLocal]", node.id, "add new id", callerId)
			}
		}
		node.lock.Unlock()
	}
	if minId^targetId < node.id^targetId {
		if debugDhtLog {
			udwLog.Log(node.id, "[findLocal]", targetId, "from caller", callerId, "closest:", minId)
		}
		return minId, nil
	}
	if debugDhtLog {
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
