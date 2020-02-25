package dht

import (
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwTest"
	"net"
	"testing"
)

func TestRpcMessageEncodeDecode(t *testing.T) {
	message := rpcMessage{
		cmd: cmdOkClosestRpcNodeList,
		closestRpcNodeList: []*rpcNode{
			{
				Id:   1,
				Ip:   net.ParseIP("1.1.1.1"),
				Port: 80,
			},
			{
				Id:   2,
				Ip:   net.ParseIP("1.1.1.2"),
				Port: 443,
			},
		},
	}
	buf := udwBytes.NewBufWriter(nil)
	rpcMessageEncode(buf, message)
	_message, err := rpcMessageDecode(buf.GetBytes())
	udwTest.Equal(err, nil)
	udwTest.Equal(_message.cmd, cmdOkClosestRpcNodeList)
	udwTest.Equal(len(_message.closestRpcNodeList), 2)
	udwTest.Equal(_message.closestRpcNodeList[0].Id, uint64(1))
	udwTest.Equal(_message.closestRpcNodeList[0].Ip.To4().String(), "1.1.1.1")
	udwTest.Equal(_message.closestRpcNodeList[0].Port, uint16(80))
	udwTest.Equal(_message.closestRpcNodeList[1].Id, uint64(2))
	udwTest.Equal(_message.closestRpcNodeList[1].Ip.To4().String(), "1.1.1.2")
	udwTest.Equal(_message.closestRpcNodeList[1].Port, uint16(443))
}
