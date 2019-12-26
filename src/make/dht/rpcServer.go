package dht

import (
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwLog"
	"net"
	"strconv"
)

const rpcPort = 19283

func (node *peerNode) StartRpcServer() {
	packetConn, err := net.ListenPacket("udp", ":"+strconv.Itoa(rpcPort))
	udwErr.PanicIfError(err)
	rBuf := make([]byte, 2<<10)
	wBuf := udwBytes.NewBufWriter(nil)
	for {
		n, addr, err := packetConn.ReadFrom(rBuf)
		if err != nil {
			udwLog.Log("[g7ath8f3dq]", err)
			continue
		}
		request := rpcMessage{}
		err = request.decode(rBuf[:n])
		if err != nil {
			udwLog.Log("[xj4w3w2yh9]", err)
			continue
		}
		switch request.cmd {
		case cmdStore:
			node.store(request.data)
			response := rpcMessage{
				cmd:      cmdOk,
				idSender: node.id,
			}
			wBuf.Reset()
			response.encode(wBuf)
			_, err = packetConn.WriteTo(wBuf.GetBytes(),addr)
			if err != nil {
				udwLog.Log("[m3v73uce68]", err)
				continue
			}
		default:
			udwLog.Log("[8yty9m5r2v] unknown cmd[" + strconv.Itoa(int(request.cmd)) + "]")
			continue
		}
	}
}
