package dht

import (
	"fmt"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwClose"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwLog"
	"net"
	"strconv"
)

func (node *peerNode) StartRpcServer() (close func()) {
	if node.port == 0 {
		panic("node.port can't be 0")
	}
	closer := udwClose.NewCloser()
	packetConn, err := net.ListenPacket("udp", ":"+strconv.Itoa(int(node.port)))
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
			request, err := rpcMessageDecode(rBuf[:n])
			if err != nil {
				udwLog.Log("[xj4w3w2yh9]", err)
				continue
			}
			fmt.Println(">>> request.targetId",request.targetId)
			response := rpcMessage{
				idSender:   node.id,
				_idMessage: request._idMessage,
			}
			switch request.cmd {
			//case cmdPing:
			//case cmdStore:
			//	node.store(request.data)
			case cmdFindNode, cmdFindValue:
				//TODO add sender to buckets
				isValue := request.cmd == cmdFindValue
				closestIdList, value := node.findLocal(request.targetId, isValue)
				if isValue && value != nil {
					response.cmd = cmdOkValue
					response.value = value
				} else {
					response.cmd = cmdOkClosestRpcNodeList
					response.closestRpcNodeList = node.getRpcNodeList(closestIdList)
				}
				//node.getRpcNodeList(closestIdList)
				//bufSize := 1 + 8*len(closestIdList) + len(value)
				//response.data = make([]byte, bufSize)
				//response.data[0] = byte(len(closestIdList))
				//for i, id := range closestIdList {
				//	binary.BigEndian.PutUint64(response.data[1+i*8:1+i*8+8], id)
				//}
				//if len(value) > 0 {
				//	copy(response.data[1+8*len(closestIdList):], value)
				//}
			default:
				udwLog.Log("[8yty9m5r2v] unknown cmd[" + strconv.Itoa(int(request.cmd)) + "]")
				continue
			}
			wBuf.Reset()
			rpcMessageEncode(wBuf, response)
			_, err = packetConn.WriteTo(wBuf.GetBytes(), addr)
			if err != nil {
				udwLog.Log("[m3v73uce68]", addr, err)
				continue
			}
		}
	}()
	return closer.Close
}
