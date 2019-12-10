package main

import (
	"github.com/tachyon-protocol/udw/udwConsole"
	"github.com/tachyon-protocol/udw/tyVpnRouteServer"
)

func main(){
	udwConsole.MustRunCommandLineFromFuncV2(tyVpnRouteServer.RouteServerRunCmd)
}
