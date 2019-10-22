package SimpleVpnPacket

import (
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwRand/udwRandNewId"
	"github.com/tachyon-protocol/udw/udwTest"
	"testing"
)

func Test_encode_decode_cmdData (t *testing.T) {
	clientId := udwRandNewId.NextUint64Id()
	packet := &vpnPacket{
		cmd:          cmdData,
		clientIdFrom: clientId,
		data:         []byte{0x00},
	}
	bufW := &udwBytes.BufWriter{}
	n := encode(packet,bufW)
	udwTest.Equal(n,10)
	packet.reset()
	_buf := bufW.GetBytes()
	err := decode(packet, _buf)
	udwTest.Equal(err,nil)
	udwTest.Equal(packet.cmd, cmdData)
	udwTest.Equal(packet.clientIdFrom, clientId)
	udwTest.Equal(packet.data, []byte{0x00})
}

func Test_encode_decode_cmdForward (t *testing.T) {
	clientId := udwRandNewId.NextUint64Id()
	packet := &vpnPacket{
		cmd:               cmdForward,
		clientIdForwardTo: clientId,
		data:              []byte{0x00},
	}
	bufW := &udwBytes.BufWriter{}
	n := encode(packet,bufW)
	udwTest.Equal(n,10)
	packet.reset()
	_buf := bufW.GetBytes()
	err := decode(packet, _buf)
	udwTest.Equal(err,nil)
	udwTest.Equal(packet.cmd, cmdForward)
	udwTest.Equal(packet.clientIdForwardTo, clientId)
	udwTest.Equal(packet.data, []byte{0x00})
}
