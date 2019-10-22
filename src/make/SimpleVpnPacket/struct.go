package SimpleVpnPacket

const (
	cmdData byte = 1
	cmdForward byte = 2
)

type vpnPacket struct {
	cmd byte
	clientIdFrom uint64
	clientIdForwardTo uint64
	data []byte
}

func (packet *vpnPacket) reset (){
	packet.cmd = 0
	packet.clientIdFrom = 0
	packet.clientIdForwardTo = 0
	packet.data	= packet.data[:0]
}
