package main

import (
	"github.com/tachyon-protocol/udw/udwConsole"
	"tachyonVpnServer"
)

func main() {
	server := &tachyonVpnServer.Server{}
	udwConsole.MustRunCommandLineFromFuncV2(server.Run)
}
