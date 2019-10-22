package SimpleVpnPacket

import (
	"encoding/binary"
	"errors"
)

func decode(packet *vpnPacket, buf []byte) error {
	if len(buf) < 1 {
		return errors.New("len(buf) < 1")
	}
	packet.cmd = buf[0]
	switch packet.cmd {
	case cmdData:
		packet.clientIdFrom = binary.BigEndian.Uint64(buf[1:])
	case cmdForward:
		packet.clientIdForwardTo = binary.BigEndian.Uint64(buf[1:])
	}
	packet.data = buf[9:]
	return nil
}
