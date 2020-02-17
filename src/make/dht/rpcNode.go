package dht

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwClose"
	"github.com/tachyon-protocol/udw/udwLog"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	cmdPing      byte = 0
	cmdStore     byte = 1
	cmdFindNode  byte = 2
	cmdFindValue byte = 3
	cmdOk        byte = 4
	//cmdError     byte = 5
	cmdOkValue              byte = 6
	cmdOkClosestRpcNodeList byte = 7
)

func getCmdString(cmd byte) string {
	switch cmd {
	case cmdPing:
		return "PING"
	case cmdStore:
		return "STORE"
	case cmdFindNode:
		return "FIND_NODE"
	case cmdFindValue:
		return "FIND_VALUE"
	case cmdOk:
		return "OK"
	//case cmdError:
	//	return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type rpcMessage struct {
	cmd        byte
	_idMessage uint32 //do not set this manually
	idSender   uint64
	portSender uint16

	targetId           uint64
	closestRpcNodeList []*rpcNode
	value              []byte
}

func rpcMessageDecode(buf []byte) (message rpcMessage, err error) {
	if len(buf) < 13 {
		return message, errors.New("[d5tkk1grb1rk] input too short " + strconv.Itoa(len(buf)))
	}
	message.cmd = buf[0]
	message._idMessage = binary.BigEndian.Uint32(buf[1:5])
	message.idSender = binary.BigEndian.Uint64(buf[5:13])
	switch message.cmd {
	case cmdFindNode, cmdFindValue:
		if len(buf) < 13+8 {
			return message, errors.New("[bpc1cpn8d2h] input too short " + strconv.Itoa(len(buf)))
		}
		message.targetId = binary.BigEndian.Uint64(buf[13 : 13+8])
	}
	return message, nil
}

func rpcMessageEncode(buf *udwBytes.BufWriter, message rpcMessage) {
	buf.WriteByte_(message.cmd)
	buf.WriteBigEndUint32(message._idMessage)
	buf.WriteBigEndUint64(message.idSender)
	switch message.cmd {
	case cmdFindNode, cmdFindValue:
		buf.WriteBigEndUint64(message.targetId)
	}
}

//func (message *rpcMessage) parseData() (closestRpcNodeList []*rpcNode, value []byte, err error) {
//	if len(message.data) < 1 {
//		return nil, nil, errors.New("[88n4mc5439]")
//	}
//	switch message.cmd {
//	case cmdOk:
//		//TODO
//		return nil, nil, nil
//	case cmdOkClosestRpcNodeList:
//		const oneRpcNodeSize = 8 + 4 + 2
//		size := int(message.data[0])
//		if size > 0 {
//			closestRpcNodeList = make([]*rpcNode, 0, size)
//			for i := 0; i < size; i++ {
//				start := 1 + i*oneRpcNodeSize
//				if i >= len(message.data) || start+oneRpcNodeSize > len(message.data) {
//					udwLog.Log("[WARNING cc8t3643qe] size is", size, "but len(message.data) is", len(message.data))
//					return closestRpcNodeList, nil, nil
//				}
//				rNode := &rpcNode{
//					Id: binary.BigEndian.Uint64(message.data[start : start+8]),
//				}
//				start += 8
//				rNode.Ip = message.data[start : start+4]
//				start += 4
//				rNode.Port = binary.BigEndian.Uint16(message.data[start : start+2])
//				closestRpcNodeList = append(closestRpcNodeList, rNode)
//			}
//		}
//		return closestRpcNodeList, nil, nil
//	case cmdOkValue:
//		return nil, message.data, nil
//	default:
//		return nil, nil, errors.New("[u4ecv1aqf1cx] parse failed: unknown cmd[" + strconv.Itoa(int(message.cmd)) + "]")
//	}
//}

func newRandomMessageId() uint32 {
	var tmpBuf [4]byte
	_, err := rand.Read(tmpBuf[:])
	if err != nil {
		panic(err)
	}
	ret := binary.LittleEndian.Uint32(tmpBuf[:])
	return ret
}

type rpcNode struct {
	Id   uint64
	Ip   []byte //TODO support IPv6 address
	Port uint16

	callerId         uint64
	closer           udwClose.Closer
	lock             sync.Mutex
	conn             net.Conn
	wBuf             udwBytes.BufWriter
	rBuf             []byte
	lastResponseTime time.Time //TODO update this when any rpc request sent
}

const errorRpcCallResponseTimeout = "hgy1hkd1w7xs"

func (rNode *rpcNode) call(request rpcMessage) (response rpcMessage, err error) {
	rNode.lock.Lock()
	defer rNode.lock.Unlock()
	if rNode.conn == nil {
		if debugRpcLog {
			udwLog.Log("[rpcNode call] new conn to", net.IP(rNode.Ip).To4().String())
		}
		conn, err := net.Dial("udp", net.IP(rNode.Ip).To4().String()+":"+strconv.Itoa(rpcPort))
		if err != nil {
			return response, errors.New("[y9e4v8pvp7]" + err.Error())
		}
		rNode.conn = conn
		rNode.closer.AddOnClose(func() {
			_ = conn.Close()
		})
	}
	rNode.wBuf.Reset()
	request._idMessage = newRandomMessageId()
	//request.encode(&rNode.wBuf)
	rpcMessageEncode(&rNode.wBuf, request)
	if debugRpcLog {
		udwLog.Log("[rpcNode call] send", getCmdString(request.cmd), request._idMessage)
	}
	_, err = rNode.conn.Write(rNode.wBuf.GetBytes())
	if err != nil {
		return response, errors.New("[8srn1mzp1tkr]" + err.Error())
	}
	if rNode.rBuf == nil {
		rNode.rBuf = make([]byte, 2<<10)
	}
	err = rNode.conn.SetReadDeadline(time.Now().Add(timeoutRpcRead))
	if err != nil {
		return response, errors.New("[ds3y24s5gu]" + err.Error())
	}
	for {
		n, _err := rNode.conn.Read(rNode.rBuf)
		if _err != nil {
			return response, errors.New("[" + errorRpcCallResponseTimeout + "]" + _err.Error())
		}
		response, err := rpcMessageDecode(rNode.rBuf[:n])
		if err != nil {
			udwLog.Log("[tfq1jmc1a9v8]", err.Error())
			continue
		}
		if response._idMessage == request._idMessage {
			switch response.cmd {
			case cmdOk:
				if debugRpcLog {
					udwLog.Log("[rpcNode call] receive", getCmdString(response.cmd), response._idMessage)
				}
				return response, nil
			//case cmdError:
			//	return nil, errors.New("[mnh3apk1u8b] error[" + string(response.data) + "]")
			default:
				if debugRpcLog {
					udwLog.Log("[rpcNode call] receive", getCmdString(response.cmd), response._idMessage)
				}
				continue
			}
		}
		udwLog.Log("[7dwn1kjg1uqe] _idMessage[" + strconv.Itoa(int(response._idMessage)) + "] not match request[" + strconv.Itoa(int(request._idMessage)) + "]")
		continue
	}
}

//func (rNode *rpcNode) store(v []byte) error {
//	_, err := rNode.call(rpcMessage{
//		cmd:      cmdStore,
//		idSender: rNode.callerId,
//		data:     v,
//	})
//	if err != nil {
//		return errors.New("[fz4qqp4j9k]" + err.Error())
//	}
//	return nil
//}

//func (rNode *rpcNode) ping() error {
//	_, err := rNode.call(rpcMessage{
//		cmd:      cmdPing,
//		idSender: rNode.callerId,
//	})
//	if err != nil {
//		return errors.New("[f2red8en1bc]" + err.Error())
//	}
//	return nil
//}

func (rNode *rpcNode) find(id uint64, isFindValue bool) (closestRpcNodeList []*rpcNode, value []byte, err error) {
	cmd := cmdFindNode
	if isFindValue {
		cmd = cmdFindValue
	}
	req := rpcMessage{
		cmd:      cmd,
		idSender: rNode.callerId,
		targetId: id,
	}
	resp, err := rNode.call(req)
	if err != nil {
		return nil, nil, errors.New("[xkx1veu5dqp]" + err.Error())
	}
	return resp.closestRpcNodeList, resp.value, nil
}
