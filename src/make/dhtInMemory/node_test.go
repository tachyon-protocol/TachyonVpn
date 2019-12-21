package dhtInMemory

import (
	"github.com/tachyon-protocol/udw/udwTest"
	"testing"
)

func Test (t *testing.T){
	node0 := newNode(0)
	node1 := newNode(node0.id)
	node2 := newNode(node1.id)
	const data = "Prometheus"
	key := hash([]byte(data))
	node0.store([]byte(data))
	v := node2.findValue(key)
	udwTest.Ok(string(v)==data)
}

func TestFindNode (t *testing.T){
	node0 := newNode(0)
	node1 := newNode(node0.id)
	node2 := newNode(node1.id)
	node3 := newNode(node2.id)
	closestId := node3.findNode(node0.id)
	udwTest.Ok(closestId==node0.id)
}
