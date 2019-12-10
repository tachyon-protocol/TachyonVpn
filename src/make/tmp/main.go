package main

import (
	"github.com/tachyon-protocol/udw/udwDebug"
	"fmt"
	"github.com/tachyon-protocol/udw/tyVpnRouteServer/tyVpnRouteClient"
	"github.com/tachyon-protocol/udw/tyVpnProtocol"
)

func main(){
	//udwRpc2Tester.BuildAndTest()
	routeC:=tyVpnRouteClient.Rpc_NewClient(tyVpnProtocol.PublicRouteServerAddr)
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
