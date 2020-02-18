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
	port     uint16
	lock     sync.RWMutex
	keyMap   map[uint64][]byte
	kBuckets [64]map[uint64]*rpcNode
}

type newPeerNodeRequest struct {
	id                   uint64
	port                 uint16
	bootstrapRpcNodeList []*rpcNode
}

func newPeerNode(req newPeerNodeRequest) *peerNode {
	if req.id == 0 {
		req.id = udwRand.MustCryptoRandUint64()
	}
	n := &peerNode{
		id:       req.id,
		port:     req.port,
		keyMap:   map[uint64][]byte{},
		kBuckets: [64]map[uint64]*rpcNode{},
	}
	n.updateBuckets(req.bootstrapRpcNodeList)
	if debugMemoryMode {
		rpcInMemoryRegister(n)
	}
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

func (node *peerNode) find(targetId uint64, isFindValue bool) (closestRpcNodeList []*rpcNode, value []byte) {
	closestIdList, value := node.findLocal(targetId, isFindValue)
	if isFindValue && value != nil {
		return nil, value
	}
	if len(closestIdList) == 0 {
		return nil, nil
	}
	if !isFindValue {
		for _, id := range closestIdList {
			if id == targetId {
				return node.getRpcNodeList(closestIdList), nil
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
		for _, id := range idToDistanceMap.KeysByValueAsc() {
			if requestedNodeIdMap[id] {
				continue
			}
			requestedNodeIdMap[id] = true
			rNode := node.getRpcNode(id)
			if rNode == nil {
				udwLog.Log("[cgc1e8b2p3q] can find rpcNode", id, "on", node.id)
				continue
			}
			_closestRpcNodeList, _value, err := rNode.find(targetId, isFindValue)
			if err != nil {
				udwLog.Log("[43eav1fmk5s] ask", id, "to find", targetId, "on", node.id, "failed:", err)
				continue
			}
			node.updateBuckets(_closestRpcNodeList)
			if isFindValue && _value != nil {
				return nil, _value
			}
			for _, rNode := range _closestRpcNodeList {
				if !isFindValue && rNode.Id == targetId {
					return _closestRpcNodeList, nil
				}
				distance := targetId ^ rNode.Id
				idToDistanceMap.Set(rNode.Id, distance)
				if distance < minDistance {
					minDistance = distance
				}
			}
		}
		if minDistance >= _minDistance { //not found
			return closestRpcNodeList[:udwMath.IntMin(len(closestRpcNodeList), k)], nil
		}
	}
}

func (node *peerNode) findLocal(targetId uint64, isValue bool) (closestIdList []uint64, value []byte) {
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
			idToDistanceMap.Set(id, targetId^id) //TODO cache the result?
		}
	}
	node.lock.RUnlock()
	closestIdList = idToDistanceMap.KeysByValueAsc()
	if debugDhtLog {
		udwLog.Log("[findLocal]", node.id, "target", targetId, "closest id rank:")
		for _, id := range closestIdList {
			distance, _ := idToDistanceMap.Get(id)
			udwLog.Log("           ", id, "distance", distance)
		}
	}
	//if callerRpcNode == targetId {
	//	//TODO add caller's rpcNode
	//	//node.updateBuckets(callerId)
	//}
	return closestIdList[:udwMath.IntMin(len(closestIdList), k)], nil
}

func (node *peerNode) getRpcNodeList(idList []uint64) []*rpcNode {
	if len(idList) == 0 {
		return nil
	}
	rpcNodeList := make([]*rpcNode, 0, len(idList))
	for _, id := range idList {
		rpcNodeList = append(rpcNodeList, node.getRpcNode(id))
	}
	return rpcNodeList
}

func (node *peerNode) getRpcNode(id uint64) *rpcNode {
	cps := sizeOfCommonPrefix(id, node.id)
	node.lock.RLock()
	m := node.kBuckets[cps]
	if m != nil {
		rNode, exist := m[id]
		if exist {
			node.lock.RUnlock()
			return rNode
		}
	}
	node.lock.RUnlock()
	return nil
}

func (node *peerNode) deleteRpcNode(id uint64) {
	cps := sizeOfCommonPrefix(id, node.id)
	node.lock.Lock()
	m := node.kBuckets[cps]
	if m != nil {
		delete(m, id)
	}
	node.lock.Unlock()
}

func (node *peerNode) updateBuckets(rpcNodeList []*rpcNode) {
	node.lock.Lock()
	for _, rNode := range rpcNodeList {
		if rNode == nil {
			continue
		}
		if rNode.Id == node.id {
			continue
		}
		cps := sizeOfCommonPrefix(rNode.Id, node.id)
		m := node.kBuckets[cps]
		if m == nil {
			m = map[uint64]*rpcNode{}
		}
		if m[rNode.Id] == nil {
			m[rNode.Id] = rNode
			node.kBuckets[cps] = m
			if debugDhtLog {
				udwLog.Log("[updateBuckets]", node.id, "add new rpcNode", rNode.Id, "cps", cps)
			}
		}
	}
	node.lock.Unlock()
}

//TODO
//func (node *peerNode) gcBuckets() {
//	var checkList []*rpcNode
//	now := time.Now()
//	node.lock.RLock()
//	for _, m := range node.kBuckets {
//		if len(m) == 0 {
//			continue
//		}
//		for _, rNode := range m {
//			rNode.lock.Lock()
//			delta := now.Sub(rNode.lastResponseTime)
//			rNode.lock.Unlock()
//			if delta > timeoutRpcNodeInBuckets {
//				checkList = append(checkList, rNode)
//			}
//		}
//	}
//	node.lock.RUnlock()
//	for _, rNode := range checkList {
//		err := rNode.ping()
//		if err != nil {
//			node.deleteRpcNode(rNode.Id)
//		} else {
//			rNode.lock.Lock()
//			rNode.lastResponseTime = time.Now()
//			rNode.lock.Unlock()
//		}
//	}
//}

func (node *peerNode) store(v []byte) {
	node.lock.Lock()
	node.keyMap[hash(v)] = v
	node.lock.Unlock()
}

func (node *peerNode) findNode(targetId uint64) (closestRpcNodeList []*rpcNode) {
	closestRpcNodeList, _ = node.find(targetId, false)
	return closestRpcNodeList
}

func (node *peerNode) findValue(key uint64) (value []byte) {
	_, value = node.find(key, true)
	return value
}
