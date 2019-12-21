package dhtInMemory

import (
	"github.com/tachyon-protocol/udw/udwTest"
	"testing"
)

func Test (t *testing.T){
	node0 := newNode(0,0)
	node1 := newNode(0,node0.id)
	node2 := newNode(0,node1.id)
	const data = "Prometheus"
	key := hash([]byte(data))
	node0.store([]byte(data))
	v := node2.findValue(key)
	udwTest.Ok(string(v)==data)
}

func TestFindNode (t *testing.T){
	node0 := newNode(0,0)
	node1 := newNode(0,node0.id)
	node2 := newNode(0,node1.id)
	node3 := newNode(0,node2.id)
	udwTest.Ok(len(node3.knownNodes)==1)
	closestId := node3.findNode(node0.id)
	udwTest.Ok(closestId==node0.id)
	udwTest.Ok(len(node3.knownNodes)==3)
}

func TestJoiningTheNetwork (t *testing.T) {
	node1 := newNode(1,0)
	node2 := newNode(2,node1.id)
	node3 := newNode(3,node2.id)
	node4 := newNode(4,node3.id)
	node5New := newNode(5,node4.id)
	closestId := node5New.findNode(node5New.id)
	udwTest.Ok(closestId != node5New.id)
	closestId = node5New.findNode(node5New.id)
	udwTest.Ok(closestId == node5New.id)
}
