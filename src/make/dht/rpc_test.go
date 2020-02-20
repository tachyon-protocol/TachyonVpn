package dht

import (
	"github.com/tachyon-protocol/udw/udwTest"
	"net"
	"testing"
)

//func TestRpcNodeStore(t *testing.T) {
//	node := newPeerNode(1) //TODO
//	closeRpcServer := node.StartRpcServer()
//	defer closeRpcServer()
//	rNode := rpcNode{
//		Id: node.id,
//		Ip: "127.0.0.1",
//	}
//	const data = "Hyperion"
//	key := hash([]byte(data))
//	err := rNode.store([]byte(data))
//	udwTest.Equal(err, nil)
//	node.lock.RLock()
//	v := node.keyMap[key]
//	node.lock.RUnlock()
//	udwTest.Equal(string(v), data)
//}

func TestRpcNodeFindNode_one_to_one(t *testing.T) {
	node1 := newPeerNode(newPeerNodeRequest{
		id:   1,
		port: 60001,
		bootstrapRpcNodeList: []*rpcNode{
			{
				Id:   2,
				Ip:   net.ParseIP("127.0.0.1"),
				Port: 60002,
			},
		},
	})
	close1 := node1.StartRpcServer()
	defer close1()
	//node2 := newPeerNode(newPeerNodeRequest{
	//	id:   2,
	//	port: 60002,
	//})
	//close2 := node2.StartRpcServer()
	//defer close2()
	node3 := newPeerNode(newPeerNodeRequest{
		id: 3,
		bootstrapRpcNodeList: []*rpcNode{
			{
				Id:   1,
				Ip:   net.ParseIP("127.0.0.1").To4(),
				Port: 60001,
			},
		},
	})
	closestRpcNodeList := node3.findNode(2)
	udwTest.Equal(len(closestRpcNodeList), 1)
	udwTest.Equal(closestRpcNodeList[0].Id, uint64(2))
	udwTest.Equal(closestRpcNodeList[0].Port, uint16(60002))
}

//func TestRpcNodeFindValue(t *testing.T) {
//	const data = "Hyperion"
//	key := hash([]byte(data))
//	node1 := newPeerNode(key)
//	node1.store([]byte(data))
//	node2 := newPeerNode(2, node1.id)
//	closeRpcServerNode2 := node2.StartRpcServer()
//	rNode2 := rpcNode{
//		Id: node2.id,
//		Ip: "127.0.0.1",
//	}
//	closestIdList, value, err := rNode2.find(key)
//	udwErr.PanicIfError(err)
//	udwTest.Equal(len(closestIdList), 1)
//	closestId := closestIdList[0]
//	udwTest.Equal(closestId, node1.id)
//	udwTest.Equal(value, []byte{})
//	closeRpcServerNode2()
//
//	closeRpcServerNode1 := node1.StartRpcServer()
//	defer closeRpcServerNode1()
//	rNodeClosest := rpcNode{
//		Id: closestId,
//		Ip: "127.0.0.1",
//	}
//	closestIdList, value, err = rNodeClosest.find(key)
//	udwErr.PanicIfError(err)
//	udwTest.Ok(len(closestIdList)==0)
//	udwTest.Equal(string(value), data)
//}

//var responseTimeoutError = errors.New("timeout")

//func debugClientSend(request []byte, afterWrite func(conn net.Conn) (isReturn bool)) (response []byte, err error) {
//	conn, err := net.Dial("udp", "127.0.0.1:"+strconv.Itoa(rpcPort))
//	udwErr.PanicIfError(err)
//	_, err = conn.Write(request)
//	udwErr.PanicIfError(err)
//	if afterWrite != nil {
//		isReturn := afterWrite(conn)
//		if isReturn {
//			return
//		}
//	}
//	buf := make([]byte, 2<<10)
//	err = conn.SetDeadline(time.Now().Add(time.Millisecond * 300))
//	udwErr.PanicIfError(err)
//	n, err := conn.Read(buf)
//	if err != nil {
//		return nil, responseTimeoutError
//	}
//	return buf[:n], nil
//}
//
//func TestRpcNodeErrorClient(t *testing.T) {
//	node := newPeerNode(0)
//	closeRpcServer := node.StartRpcServer()
//	defer closeRpcServer()
//	errMsg := ""
//	_, err := debugClientSend([]byte("1"), nil)
//	if err != nil {
//		errMsg = err.Error()
//	}
//	udwTest.Equal(errMsg, responseTimeoutError.Error())
//	_, err = debugClientSend([]byte{0x02, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, nil)
//	if err != nil {
//		errMsg = err.Error()
//	}
//	udwTest.Equal(errMsg, responseTimeoutError.Error())
//	_, err = debugClientSend([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, nil)
//	if err != nil {
//		errMsg = err.Error()
//	}
//	udwTest.Equal(errMsg, responseTimeoutError.Error())
//	_, err = debugClientSend([]byte{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}, func(conn net.Conn) bool {
//		_ = conn.Close()
//		return true
//	})
//	udwTest.Equal(err, nil)
//}
//
//func debugServerRespond(correctIdMessage bool, response []byte) (close func()) {
//	closer := udwClose.NewCloser()
//	packetConn, err := net.ListenPacket("udp", ":"+strconv.Itoa(rpcPort))
//	udwErr.PanicIfError(err)
//	closer.AddOnClose(func() {
//		_ = packetConn.Close()
//	})
//	go func() {
//		rBuf := make([]byte, 2<<10)
//		n, addr, err := packetConn.ReadFrom(rBuf)
//		udwErr.PanicIfError(err)
//		request := rpcMessage{}
//		err = request.rpcMessageDecode(rBuf[:n])
//		udwErr.PanicIfError(err)
//		if correctIdMessage && len(response) > 5 {
//			binary.BigEndian.PutUint32(response[1:5], request._idMessage)
//		}
//		_, err = packetConn.WriteTo(response, addr)
//		udwErr.PanicIfError(err)
//	}()
//	return closer.Close
//}
//
//func TestRpcNodeErrorServer(t *testing.T) {
//	rNode2 := rpcNode{
//		id: 1,
//		ip: "127.0.0.1",
//	}
//	errMsg := ""
//	_, err := rNode2.findNode(2)
//	if err != nil {
//		errMsg = err.Error()
//	}
//	udwTest.Ok(strings.Contains(errMsg, errorRpcCallResponseTimeout))
//
//	_close := debugServerRespond(false, []byte("1"))
//	_, err = rNode2.findNode(2)
//	if err != nil {
//		errMsg = err.Error()
//	}
//	udwTest.Ok(strings.Contains(errMsg, errorRpcCallResponseTimeout))
//	_close()
//
//	_close = debugServerRespond(true, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
//	err = rNode2.store([]byte("123"))
//	if err != nil {
//		errMsg = err.Error()
//	}
//	udwTest.Ok(strings.Contains(errMsg, errorRpcCallResponseTimeout))
//	_close()
//
//	_close = debugServerRespond(false, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
//	_, _, err = rNode2.findValue(2)
//	if err != nil {
//		errMsg = err.Error()
//	}
//	udwTest.Ok(strings.Contains(errMsg, errorRpcCallResponseTimeout))
//	_close()
//
//	_close = debugServerRespond(true, []byte{cmdOk, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
//	_, err = rNode2.findNode(2)
//	if err != nil {
//		errMsg = err.Error()
//	}
//	udwTest.Ok(strings.Contains(errMsg, "fhf1b2xk9u9"))
//	_close()
//
//	_close = debugServerRespond(true, []byte{cmdOk, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
//	_, _, err = rNode2.findValue(2)
//	if err != nil {
//		errMsg = err.Error()
//	}
//	udwTest.Ok(strings.Contains(errMsg, "kge9ma4b69"))
//	_close()
//}
