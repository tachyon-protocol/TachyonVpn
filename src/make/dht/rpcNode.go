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
	cmdError     byte = 5
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
	case cmdError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type rpcMessage struct {
	cmd        byte
	_idMessage uint32 //do not set this manually
	idSender   uint64
	data       []byte
}

func (packet *rpcMessage) decode(buf []byte) error {
	if len(buf) < 13 {
		return errors.New("[d5tkk1grb1rk] input too short " + strconv.Itoa(len(buf)))
	}
	packet.cmd = buf[0]
	packet._idMessage = binary.BigEndian.Uint32(buf[1:5])
	packet.idSender = binary.BigEndian.Uint64(buf[5:13])
	packet.data = buf[13:]
	return nil
}

func (packet *rpcMessage) encode(buf *udwBytes.BufWriter) {
	buf.WriteByte_(packet.cmd)
	buf.WriteBigEndUint32(packet._idMessage)
	buf.WriteBigEndUint64(packet.idSender)
	buf.Write_(packet.data)
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
	id     uint64
	ip     string
	closer udwClose.Closer
	lock   sync.Mutex
	conn   net.Conn
	wBuf   udwBytes.BufWriter
	rBuf   []byte
}

func (rNode *rpcNode) call(request rpcMessage) (response *rpcMessage, err error) {
	rNode.lock.Lock()
	defer rNode.lock.Unlock()
	if rNode.conn == nil {
		if debugRpcLog {
			udwLog.Log("[rpcNode call] new conn to", rNode.ip)
		}
		conn, err := net.Dial("udp", rNode.ip+":"+strconv.Itoa(rpcPort))
		if err != nil {
			return nil, errors.New("[y9e4v8pvp7]" + err.Error())
		}
		rNode.conn = conn
		rNode.closer.AddOnClose(func() {
			_ = conn.Close()
		})
	}
	rNode.wBuf.Reset()
	request._idMessage = newRandomMessageId()
	request.encode(&rNode.wBuf)
	if debugRpcLog {
		udwLog.Log("[rpcNode call] send", getCmdString(request.cmd), request._idMessage)
	}
	_, err = rNode.conn.Write(rNode.wBuf.GetBytes())
	if err != nil {
		return nil, errors.New("[8srn1mzp1tkr]" + err.Error())
	}
	if rNode.rBuf == nil {
		rNode.rBuf = make([]byte, 2<<10)
	}
	err = rNode.conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	if err != nil {
		return nil, errors.New("[ds3y24s5gu]" + err.Error())
	}
	for {
		n, _err := rNode.conn.Read(rNode.rBuf)
		if _err != nil {
			return nil, errors.New("[hgy1hkd1w7xs]" + _err.Error())
		}
		response = &rpcMessage{}
		err = response.decode(rNode.rBuf[:n])
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
			case cmdError:
				return nil, errors.New("[mnh3apk1u8b] error[" + string(response.data) + "]")
			default:
				return nil, errors.New("[45rau1mr258] unknown cmd[" + strconv.Itoa(int(response.cmd)) + "] data[" + string(response.data) + "]")
			}
		}
		udwLog.Log("[7dwn1kjg1uqe] _idMessage[" + strconv.Itoa(int(response._idMessage)) + "] not match request[" + strconv.Itoa(int(request._idMessage)) + "]")
		continue
	}
}

func (rNode *rpcNode) store(v []byte) error {
	_, err := rNode.call(rpcMessage{
		cmd:      cmdStore,
		idSender: rNode.id,
		data:     v,
	})
	if err != nil {
		return errors.New("[fz4qqp4j9k]" + err.Error())
	}
	return nil
}

func (rNode *rpcNode) findNode(targetId uint64) (closestId uint64, err error) {
	req := rpcMessage{
		cmd:      cmdFindNode,
		idSender: rNode.id,
		data:     make([]byte, 8),
	}
	binary.BigEndian.PutUint64(req.data, targetId)
	resp, err := rNode.call(req)
	if err != nil {
		return 0, errors.New("[7qf68n3q9g]" + err.Error())
	}
	if len(resp.data) != 8 {
		return 0, errors.New("[fhf1b2xk9u9]")
	}
	return binary.BigEndian.Uint64(resp.data), nil
}

func (rNode *rpcNode) findValue(key uint64) (closestId uint64, value []byte, err error) {
	req := rpcMessage{
		cmd:      cmdFindValue,
		idSender: rNode.id,
		data:     make([]byte, 8),
	}
	binary.BigEndian.PutUint64(req.data, key)
	resp, err := rNode.call(req)
	if err != nil {
		return 0, nil, errors.New("[xkx1veu5dqp]" + err.Error())
	}
	if len(resp.data) < 8 {
		return 0, nil, errors.New("[kge9ma4b69]")
	}
	closestId = binary.BigEndian.Uint64(resp.data)
	if len(resp.data) == 8 {
		return closestId, nil, nil
	}
	return closestId, resp.data[8:], nil
}
