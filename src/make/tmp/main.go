package main

import (
	"tachyonVpnRouteServer/tachyonVpnRouteClient"
	"tachyonVpnProtocol"
	"github.com/tachyon-protocol/udw/udwDebug"
)

func main(){
	//udwRpc2Tester.BuildAndTest()
	routeC:=tachyonVpnRouteClient.Rpc_NewClient(tachyonVpnProtocol.PublicRouteServerAddr)
	list,rpcErr:=routeC.VpnNodeList()
	if rpcErr!=nil{
		panic(rpcErr.Error())
	}
	udwDebug.Println(list)
}
