package dht

import (
	"encoding/binary"
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
		response := rpcMessage{
			cmd:        cmdOk,
			idSender:   node.id,
			_idMessage: request._idMessage,
		}
		switch request.cmd {
		case cmdStore:
			node.store(request.data)
		case cmdFindNode, cmdFindValue:
			if len(request.data) != 8 {
				response.cmd = cmdError
				response.data = []byte("[95hs5hzw68] len(request.data) != 8")
				break
			}
			isValue := request.cmd == cmdFindValue
			targetId := binary.BigEndian.Uint64(request.data)
			closestId, value := node.findLocal(request.idSender, targetId, isValue)
			response.data = make([]byte, 8+len(value))
			binary.BigEndian.PutUint64(response.data,closestId)
			if len(value) > 0 {
				copy(response.data[8:],value)
			}
		default:
			udwLog.Log("[8yty9m5r2v] unknown cmd[" + strconv.Itoa(int(request.cmd)) + "]")
			continue
		}
		wBuf.Reset()
		response.encode(wBuf)
		_, err = packetConn.WriteTo(wBuf.GetBytes(), addr)
		if err != nil {
			udwLog.Log("[m3v73uce68]", addr, err)
			continue
		}
	}
}
