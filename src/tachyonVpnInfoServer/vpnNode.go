package tachyonVpnInfoServer

import (
	"tachyonVpnInfoServer/tachyonVpnInfoClient"
	"time"
	"github.com/tachyon-protocol/udw/udwJson"
	"tachyonVpnClient"
)

func (serverRpcObj) RegisterAsVpnNode(req tachyonVpnInfoClient.RegisterAsVpnNodeReq){
	startTime:=time.Now()
	thisNode:=ServerNode{
		Ip: req.Ip,
		ServerCertPem: req.ServerCertPem,
		UpdateTime: startTime,
	}
	tachyonVpnClient.Ping(tachyonVpnClient.PingReq{

	})
	getDb().MustSet(k1VpnNodeIp,req.Ip,udwJson.MustMarshalToString(thisNode))
}

type ServerNode struct{
	Ip string
	ServerCertPem string
	UpdateTime time.Time
}
const k1VpnNodeIp = "k1VpnNodeIp2"