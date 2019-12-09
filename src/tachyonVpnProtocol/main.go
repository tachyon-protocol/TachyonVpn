package tachyonVpnProtocol

import (
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwFile"
	"github.com/tachyon-protocol/udw/udwRand"
	"github.com/tachyon-protocol/udw/udwStrconv"
	"strconv"
)

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

const VpnPort = 29444

const (
	CmdData      byte = 1
	CmdForward   byte = 2
	CmdHandshake byte = 3
	CmdPing      byte = 4
	CmdKeepAlive byte = 5
)

const PublicRouteServerAddr = "35.223.105.46:24587"

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

func GetClientId(index int) uint64 {
	clientIdPath := "/usr/local/etc/tachyonClientId"+strconv.Itoa(index)
	var id uint64
	udwErr.PanicToErrorMsgWithStackAndLog(func() {
		b := udwFile.MustReadFile(clientIdPath)
		id = udwStrconv.MustParseUint64(string(b))
	})
	if id == 0 {
		id = udwRand.MustCryptoRandUint64()
		udwFile.MustWriteFileWithMkdir(clientIdPath, []byte(udwStrconv.FormatUint64(id)))
	}
	return id
}
