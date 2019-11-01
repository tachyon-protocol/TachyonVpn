package tachyonSimpleVpnProtocol

import (
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwRand/udwRandNewId"
	"github.com/tachyon-protocol/udw/udwTest"
	"testing"
)

func Test_encode_decode_cmdData(t *testing.T) {
	clientId := udwRandNewId.NextUint64Id()
	packet := &VpnPacket{
		Cmd:            CmdData,
		ClientIdSender: clientId,
		Data:           []byte{0x00},
	}
	bufW := &udwBytes.BufWriter{}
	n := packet.Encode(bufW)
	udwTest.Equal(n, 1+16+1)
	packet.Reset()
	_buf := bufW.GetBytes()
	err := packet.Decode(_buf)
	udwTest.Equal(err, nil)
	udwTest.Equal(packet.Cmd, CmdData)
	udwTest.Equal(packet.ClientIdSender, clientId)
	udwTest.Equal(packet.Data, []byte{0x00})
}

func Test_encode_decode_cmdForward(t *testing.T) {
	clientId := udwRandNewId.NextUint64Id()
	packet := &VpnPacket{
		Cmd:              CmdForward,
		ClientIdReceiver: clientId,
		Data:             []byte{0x00},
	}
	bufW := &udwBytes.BufWriter{}
	n := packet.Encode(bufW)
	udwTest.Equal(n, 1+16+1)
	packet.Reset()
	_buf := bufW.GetBytes()
	err := packet.Decode(_buf)
	udwTest.Equal(err, nil)
	udwTest.Equal(packet.Cmd, CmdForward)
	udwTest.Equal(packet.ClientIdReceiver, clientId)
	udwTest.Equal(packet.Data, []byte{0x00})
}
