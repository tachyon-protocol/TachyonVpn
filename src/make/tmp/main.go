package main

import (
	"tachyonVpnRouteServer/tachyonVpnRouteClient"
	"tachyonVpnProtocol"
	"github.com/tachyon-protocol/udw/udwDebug"
	"fmt"
)

func main(){
	//udwRpc2Tester.BuildAndTest()
	routeC:=tachyonVpnRouteClient.Rpc_NewClient(tachyonVpnProtocol.PublicRouteServerAddr)
	fmt.Println("start 1")
	rpcErr:=routeC.Ping()
	if rpcErr!=nil{
		panic(rpcErr.Error())
	}
	fmt.Println("start 2")
	list,rpcErr:=routeC.VpnNodeList()
	if rpcErr!=nil{
		panic(rpcErr.Error())
	}
	udwDebug.Println(list)
}
