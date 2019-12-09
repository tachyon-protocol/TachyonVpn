package main

import (
	"github.com/tachyon-protocol/udw/udwConsole"
	"tachyonVpnRouteServer"
)

func main(){
	udwConsole.MustRunCommandLineFromFuncV2(tachyonVpnRouteServer.RouteServerRunCmd)
}
