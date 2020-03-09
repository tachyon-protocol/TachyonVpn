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

	targetId           uint64
	closestRpcNodeList []*rpcNode
	value              []byte
}

func rpcMessageEncode(buf *udwBytes.BufWriter, message rpcMessage) {
	buf.WriteByte_(message.cmd)
	buf.WriteBigEndUint32(message._idMessage)
	buf.WriteBigEndUint64(message.idSender)
	switch message.cmd {
	case cmdFindNode, cmdFindValue:
		buf.WriteBigEndUint64(message.targetId)
	case cmdOkClosestRpcNodeList:
		buf.WriteByte_(byte(len(message.closestRpcNodeList)))
		for _, rNode := range message.closestRpcNodeList {
			buf.WriteBigEndUint64(rNode.Id)
			buf.WriteByte_(byte(len(rNode.Ip)))
			buf.Write_(rNode.Ip)
			buf.WriteBigEndUint16(rNode.Port)
		}
	}
}

func rpcMessageDecode(buf []byte) (message rpcMessage, err error) {
	minSize := 13
	if len(buf) < minSize {
		return message, errors.New("[d5tkk1grb1rk] input too short " + strconv.Itoa(len(buf)))
	}
	message.cmd = buf[0]
	message._idMessage = binary.BigEndian.Uint32(buf[1:5])
	message.idSender = binary.BigEndian.Uint64(buf[5:13])
	switch message.cmd {
	case cmdFindNode, cmdFindValue:
		if len(buf) < minSize+8 {
			return message, errors.New("[bpc1cpn8d2h] input too short " + strconv.Itoa(len(buf)))
		}
		message.targetId = binary.BigEndian.Uint64(buf[13 : 13+8])
	case cmdOkClosestRpcNodeList:
		if len(buf) < minSize+1 {
			return message, errors.New("[bdk1fs7q1kkr]")
		}
		minSize += 1
		index := 14
		nodeSize := int(buf[13])
		message.closestRpcNodeList = make([]*rpcNode, 0, nodeSize)
		for i := 0; i < nodeSize; i++ {
			rNode := &rpcNode{}
			if len(buf) < index+8 {
				return message, errors.New("mvd4hpy1tpf")
			}
			rNode.Id = binary.BigEndian.Uint64(buf[index : index+8])
			index += 8
			if len(buf) < index {
				return message, errors.New("s4g6wak1zcy")
			}
			ipSize := int(buf[index])
			index += 1
			if len(buf) < index+ipSize {
				return message, errors.New("3d4675k29f")
			}
			rNode.Ip = make([]byte, ipSize)
			copy(rNode.Ip, buf[index:index+ipSize])
			index += ipSize
			if len(buf) < index+2 {
				return message, errors.New("dyn9hcd1j8d")
			}
			rNode.Port = binary.BigEndian.Uint16(buf[index : index+2])
			index += 2
			message.closestRpcNodeList = append(message.closestRpcNodeList, rNode)
		}
	}
	return message, nil
}

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
	Ip   net.IP
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
			udwLog.Log("[rpcNode call] new conn to", rNode.Ip.To4().String())
		}
		conn, err := net.Dial("udp", rNode.Ip.To4().String()+":"+strconv.Itoa(int(rNode.Port)))
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
		udwLog.Log("[rpcNode call] send", getCmdString(request.cmd), "_idMessage:",request._idMessage)
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
			if debugRpcLog {
				udwLog.Log("[rpcNode call] receive", getCmdString(response.cmd), response._idMessage)
			}
			switch response.cmd {
			case cmdOkClosestRpcNodeList:
				return response, nil
			//case cmdError:
			//	return nil, errors.New("[mnh3apk1u8b] error[" + string(response.data) + "]")
			default:
				udwLog.Log("[rpcNode call] receive a unexpected cmd")
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
