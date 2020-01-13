package dht

import (
	"encoding/binary"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwClose"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwLog"
	"net"
	"strconv"
)

const rpcPort = 19283

func (node *peerNode) StartRpcServer() (close func()) {
	closer := udwClose.NewCloser()
	packetConn, err := net.ListenPacket("udp", ":"+strconv.Itoa(rpcPort))
	udwErr.PanicIfError(err)
	closer.AddOnClose(func() {
		_ = packetConn.Close()
	})
	rBuf := make([]byte, 2<<10)
	wBuf := udwBytes.NewBufWriter(nil)
	go func() {
		for {
			if closer.IsClose() {
				return
			}
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
					udwLog.Log("[95hs5hzw68] len(request.data) != 8")
					continue
				}
				isValue := request.cmd == cmdFindValue
				targetId := binary.BigEndian.Uint64(request.data)
				closestIdList, value := node.findLocal(request.idSender, targetId, isValue)
				bufSize := 1 + 8*len(closestIdList) + len(value)
				response.data = make([]byte, bufSize)
				response.data[0] = byte(len(closestIdList))
				for i, id := range closestIdList {
					binary.BigEndian.PutUint64(response.data[1+i*8:1+i*8+8], id)
				}
				if len(value) > 0 {
					copy(response.data[1+8*len(closestIdList):], value)
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
	}()
	return closer.Close
}
