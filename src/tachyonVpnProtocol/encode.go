package tachyonVpnProtocol

import "github.com/tachyon-protocol/udw/udwBytes"

func (packet *VpnPacket) Encode(buf *udwBytes.BufWriter) (n int) {
	buf.WriteByte_(packet.Cmd)
	buf.WriteBigEndUint64(packet.ClientIdSender)
	buf.WriteBigEndUint64(packet.ClientIdReceiver)
	buf.Write_(packet.Data)
	return buf.GetLen()
}
