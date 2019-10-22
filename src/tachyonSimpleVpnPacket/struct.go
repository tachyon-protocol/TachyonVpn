package tachyonSimpleVpnPacket

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

func (packet *VpnPacket) Reset(){
	packet.Cmd = 0
	packet.ClientIdFrom = 0
	packet.ClientIdForwardTo = 0
	packet.Data = packet.Data[:0]
}
