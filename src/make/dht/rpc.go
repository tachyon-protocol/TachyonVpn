package dht

const rpcPort = 19283

func (n *node) StartRpcServer() {
	//packetConn, err := net.ListenPacket("udp", ":"+strconv.Itoa(rpcPort))
	//udwErr.PanicIfError(err)
	//readBuf := make([]byte, 2<<10)
	//for {
	//	n, addr, err := packetConn.ReadFrom(readBuf)
	//}
}
