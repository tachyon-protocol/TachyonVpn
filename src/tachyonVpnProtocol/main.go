package tachyonVpnProtocol

import "github.com/tachyon-protocol/udw/udwRand"

const Debug = false

const (
	overheadEncrypt      = 0
	overheadVpnHeader    = 1
	overheadIpHeader     = 20
	overheadUdpHeader    = 8
	overheadTcpHeaderMax = 60
	Mtu                  = 1460 - (overheadEncrypt + overheadVpnHeader + overheadIpHeader + overheadUdpHeader)
	Mss                  = Mtu - (overheadTcpHeaderMax - overheadUdpHeader)
)

const VpnPort = 29443

const (
	CmdData    byte = 1
	CmdForward byte = 2
	CmdHandshake byte = 3
)

type VpnPacket struct {
	Cmd              byte
	ClientIdSender   uint64
	ClientIdReceiver uint64
	Data             []byte
}

func (packet *VpnPacket) Reset() {
	packet.Cmd = 0
	packet.ClientIdSender = 0
	packet.ClientIdReceiver = 0
	packet.Data = packet.Data[:0]
}

func GetClientId() uint64 {
	return udwRand.MustCryptoRandUint64()
}
