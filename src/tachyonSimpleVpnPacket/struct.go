package tachyonSimpleVpnPacket

const (
	overheadEncrypt      = 0
	overheadVpnHeader    = 1
	overheadIpHeader     = 20
	overheadUdpHeader    = 8
	overheadTcpHeaderMax = 60
	Mtu                  = 1460 - (overheadEncrypt + overheadVpnHeader + overheadIpHeader + overheadUdpHeader)
	Mss                  = Mtu - (overheadTcpHeaderMax - overheadUdpHeader)
)

const (
	CmdData    byte = 1
	CmdForward byte = 2
)

type VpnPacket struct {
	Cmd               byte
	ClientIdFrom      uint64
	ClientIdForwardTo uint64
	Data              []byte
}

func (packet *VpnPacket) Reset() {
	packet.Cmd = 0
	packet.ClientIdFrom = 0
	packet.ClientIdForwardTo = 0
	packet.Data = packet.Data[:0]
}
