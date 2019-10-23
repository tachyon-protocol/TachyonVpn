package tachyonSimpleVpnPacket

import "github.com/tachyon-protocol/udw/udwBytes"

func (packet *VpnPacket) Encode(buf *udwBytes.BufWriter) (n int) {
	buf.WriteByte_(packet.Cmd)
	switch packet.Cmd {
	case CmdData:
		buf.WriteBigEndUint64(packet.ClientIdFrom)
	case CmdForward:
		buf.WriteBigEndUint64(packet.ClientIdForwardTo)
	}
	buf.Write_(packet.Data)
	return buf.GetLen()
}
