package main

import (
	"github.com/tachyon-protocol/udw/udwConsole"
	tachyonVpnClient "tachyonVpnServer"
)

//kmg make sshDeploy -PkgPath make/server -Ip [ip] -Command server
//kmg make sshDeploy -PkgPath make/server -Ip [ip] -Command 'server -UseRelay -RelayServerIp [ip]'
func main() {
	server := &tachyonVpnClient.Server{}
	udwConsole.MustRunCommandLineFromFuncV2(server.Run)
}
