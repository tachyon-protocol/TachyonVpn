package SimpleVpnPacket

import "github.com/tachyon-protocol/udw/udwBytes"

func encode(packet *vpnPacket, buf *udwBytes.BufWriter) (n int) {
	buf.WriteByte_(packet.cmd)
	switch packet.cmd {
	case cmdData:
		buf.WriteBigEndUint64(packet.clientIdFrom)
	case cmdForward:
		buf.WriteBigEndUint64(packet.clientIdForwardTo)
	}
	buf.Write_(packet.data)
	return buf.GetLen()
}
