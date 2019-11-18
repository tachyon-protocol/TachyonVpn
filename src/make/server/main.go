package main

import (
	"github.com/tachyon-protocol/udw/udwConsole"
	"tachyonVpnServer"
)

//kmg make sshDeploy -PkgPath make/server -Ip [ip] -Command server
//kmg make sshDeploy -PkgPath make/server -Ip [ip] -Command 'server -UseRelay -RelayServerIp [ip]'
func main() {
	server := &tachyonVpnServer.Server{}
	udwConsole.MustRunCommandLineFromFuncV2(server.Run)
}
