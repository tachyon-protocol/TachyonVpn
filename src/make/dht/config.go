package dht

import "time"

var k = 4

const (
	timeoutRpcRead          = time.Second * 5
	timeoutRpcNodeInBuckets = time.Minute
	intervalGcBuckets       = time.Minute
)

//debug flags
const (
	debugDhtLog     = true
	debugRpcLog     = true
	debugMemoryMode = false
)
