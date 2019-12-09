package tachyonVpnRouteServer

import (
	"tachyonVpnRouteServer/tachyonVpnRouteClient"
	"time"
	"github.com/tachyon-protocol/udw/udwJson"
	"tachyonVpnClient"
	"github.com/tachyon-protocol/udw/udwRpc2"
	"github.com/tachyon-protocol/udw/udwSqlite3"
)

func (serverRpcObj) VpnNodeRegister(clientIp udwRpc2.PeerIp,thisNode tachyonVpnRouteClient.VpnNode) (errMsg string){
	startTime:=time.Now().UTC()
	thisNode.UpdateTime = startTime.Truncate(time.Second)
	if thisNode.Ip==""{
		thisNode.Ip = clientIp.Ip
	}
	err:=tachyonVpnClient.Ping(tachyonVpnClient.PingReq{
		Ip: thisNode.Ip,
		ServerChk: thisNode.ServerChk,
	})
	if err!=nil{
		return "f3pbhbjveg "+err.Error()
	}
	getDb().MustSet(k1VpnNodeIp,thisNode.Ip,udwJson.MustMarshalToString(thisNode))
	return ""
}

func (serverRpcObj) VpnNodeList() []tachyonVpnRouteClient.VpnNode{
	outList:=[]tachyonVpnRouteClient.VpnNode{}
	getDb().MustGetRangeCallback(udwSqlite3.GetRangeReq{
		K1: k1VpnNodeIp,
	},func(key string, value string){
		var thisNode tachyonVpnRouteClient.VpnNode
		udwJson.MustUnmarshalFromString(value,&thisNode)
		if isNodeTimeout(thisNode)==false{
			outList = append(outList,thisNode)
		}else{
			getDb().MustDeleteWithKv(k1VpnNodeIp,key,value)
		}
	})
	return outList
}

func (serverRpcObj) Ping(){}

func initGcVpnNode(){
	go func(){
		for{
			time.Sleep(k1VpnNodeTtl)
			getDb().MustGetRangeCallback(udwSqlite3.GetRangeReq{
				K1: k1VpnNodeIp,
			},func(key string, value string) {
				var thisNode tachyonVpnRouteClient.VpnNode
				udwJson.MustUnmarshalFromString(value,&thisNode)
				if isNodeTimeout(thisNode){
					getDb().MustDeleteWithKv(k1VpnNodeIp,key,value)
				}
			})
		}
	}()
}

func isNodeTimeout(thisNode tachyonVpnRouteClient.VpnNode) bool{
	return time.Now().Add(-k1VpnNodeTtl).After(thisNode.UpdateTime)
}

const k1VpnNodeIp = "k1VpnNodeIp2"
const k1VpnNodeTtl = time.Minute