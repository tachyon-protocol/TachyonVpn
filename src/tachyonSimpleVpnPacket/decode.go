package tachyonSimpleVpnPacket

import (
	"encoding/binary"
	"errors"
)

func (packet *VpnPacket) Decode(buf []byte) error {
	if len(buf) < 1 {
		return errors.New("len(buf) < 1")
	}
	packet.Cmd = buf[0]
	switch packet.Cmd {
	case CmdData:
		packet.ClientIdFrom = binary.BigEndian.Uint64(buf[1:])
	case CmdForward:
		packet.ClientIdForwardTo = binary.BigEndian.Uint64(buf[1:])
	}
	packet.Data = buf[9:]
	return nil
}
