package tachyonVpnProtocol

import (
	"encoding/binary"
	"errors"
)

func (packet *VpnPacket) Decode(buf []byte) error {
	if len(buf) < 1 {
		return errors.New("len(buf) < 1")
	}
	packet.Cmd = buf[0]
	packet.ClientIdSender = binary.BigEndian.Uint64(buf[1:9])
	packet.ClientIdReceiver = binary.BigEndian.Uint64(buf[9:17])
	packet.Data = buf[17:]
	return nil
}
