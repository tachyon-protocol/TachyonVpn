package dht

import (
	"encoding/binary"
	"github.com/tachyon-protocol/udw/udwCryptoSha3"
	"github.com/tachyon-protocol/udw/udwLog"
	"github.com/tachyon-protocol/udw/udwMath"
	"github.com/tachyon-protocol/udw/udwRand"
	"github.com/tachyon-protocol/udw/udwSortedMap"
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

func sizeOfCommonPrefix(a, b uint64) int {
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

func (node *peerNode) find(targetId uint64, isValue bool) (closestIdList []uint64, value []byte) {
	closestIdList, value = node.findLocal(node.id, targetId, isValue)
	if len(closestIdList) == 0 {
		return nil, value
	}
	if isValue && value != nil {
		return nil, value
	}
	if !isValue {
		for _, id := range closestIdList {
			if id == targetId {
				return closestIdList, nil
			}
		}
	}
	var (
		idToDistanceMap           = udwSortedMap.NewUint64ToUint64Map()
		minDistance        uint64 = math.MaxUint64
		requestedNodeIdMap        = map[uint64]bool{}
	)
	for _, id := range closestIdList {
		distance := targetId ^ id
		idToDistanceMap.Set(id, distance)
		if distance < minDistance {
			minDistance = distance
		}
	}
	for {
		_minDistance := minDistance
		closestIdList = idToDistanceMap.KeysByValueAsc()
		for _, id := range closestIdList {
			if requestedNodeIdMap[id] {
				continue
			}
			requestedNodeIdMap[id] = true
			_node := rpcInMemoryGetNode(id)
			_closestIdList, _value := _node.findLocal(node.id, targetId, isValue)
			node.updateBuckets(_closestIdList...)
			if isValue && _value != nil {
				return _closestIdList, _value
			}
			for _, id := range _closestIdList {
				if !isValue && id == targetId {
					return _closestIdList, nil
				}
				distance := targetId ^ id
				idToDistanceMap.Set(id, distance)
				if distance < minDistance {
					minDistance = distance
				}
			}
		}
		if minDistance == _minDistance {
			return closestIdList[:udwMath.IntMin(len(closestIdList), k)], value
		}
	}
}

func (node *peerNode) findLocal(callerId uint64, targetId uint64, isValue bool) (closestIdList []uint64, value []byte) {
	if isValue {
		node.lock.RLock()
		v, exist := node.keyMap[targetId]
		node.lock.RUnlock()
		if exist {
			return nil, v
		}
	}
	idToDistanceMap := udwSortedMap.NewUint64ToUint64Map()
	node.lock.RLock()
	for _, km := range node.kBuckets {
		for id := range km {
			idToDistanceMap.Set(id, targetId^id)
		}
	}
	node.lock.RUnlock()
	closestIdList = idToDistanceMap.KeysByValueAsc()
	if debugDhtLog {
		udwLog.Log("[findLocal]", node.id, "target", targetId, "closest id rank: caller", callerId)
		for _, id := range closestIdList {
			distance, _ := idToDistanceMap.Get(id)
			udwLog.Log("		", id, "distance", distance)
		}
	}
	if callerId == targetId {
		node.updateBuckets(callerId)
	}
	return closestIdList[:udwMath.IntMin(len(closestIdList), k)], nil
}

func (node *peerNode) updateBuckets(ids ...uint64) {
	node.lock.Lock()
	for _, id := range ids {
		if id == node.id {
			continue
		}
		cps := sizeOfCommonPrefix(id, node.id)
		if debugDhtLog {
			udwLog.Log("[updateBuckets]", node.id, id,"cps", cps)
		}
		m := node.kBuckets[cps]
		if m == nil {
			m = map[uint64]bool{}
		}
		if !m[id] {
			m[id] = true
			node.kBuckets[cps] = m
			if debugDhtLog {
				udwLog.Log("[updateBuckets]", node.id, "add new id", id)
			}
		}
	}
	node.lock.Unlock()
}

//TODO ping

func (node *peerNode) store(v []byte) {
	node.lock.Lock()
	node.keyMap[hash(v)] = v
	node.lock.Unlock()
}

func (node *peerNode) findNode(targetId uint64) (closestIdList []uint64) {
	closestIdList, _ = node.find(targetId, false)
	return closestIdList
}

func (node *peerNode) findValue(key uint64) (value []byte) {
	_, value = node.find(key, true)
	return value
}
