package tachyonSimpleVpnProtocol

import (
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwRand/udwRandNewId"
	"github.com/tachyon-protocol/udw/udwTest"
	"testing"
)

func Test_encode_decode_cmdData (t *testing.T) {
	clientId := udwRandNewId.NextUint64Id()
	packet := &VpnPacket{
		Cmd:          CmdData,
		ClientIdFrom: clientId,
		Data:         []byte{0x00},
	}
	bufW := &udwBytes.BufWriter{}
	n := Encode(packet,bufW)
	udwTest.Equal(n,10)
	packet.Reset()
	_buf := bufW.GetBytes()
	err := Decode(packet, _buf)
	udwTest.Equal(err,nil)
	udwTest.Equal(packet.Cmd, CmdData)
	udwTest.Equal(packet.ClientIdFrom, clientId)
	udwTest.Equal(packet.Data, []byte{0x00})
}

func Test_encode_decode_cmdForward (t *testing.T) {
	clientId := udwRandNewId.NextUint64Id()
	packet := &VpnPacket{
		Cmd:               CmdForward,
		ClientIdForwardTo: clientId,
		Data:              []byte{0x00},
	}
	bufW := &udwBytes.BufWriter{}
	n := Encode(packet,bufW)
	udwTest.Equal(n,10)
	packet.Reset()
	_buf := bufW.GetBytes()
	err := Decode(packet, _buf)
	udwTest.Equal(err,nil)
	udwTest.Equal(packet.Cmd, CmdForward)
	udwTest.Equal(packet.ClientIdForwardTo, clientId)
	udwTest.Equal(packet.Data, []byte{0x00})
}
