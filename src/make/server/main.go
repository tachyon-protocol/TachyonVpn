package main

import (
	"github.com/tachyon-protocol/udw/udwConsole"
	"tachyonVpnServer"
)

//relay server
//sshDeploy -PkgPath make/server -Ip ip -Command 'server -Token relay123'
//vpe server
//sshDeploy -PkgPath make/server -Ip vpeServerIp -Command 'server -Token exit123 -UseRelay -RelayServerIp relayServerIp -RelayServerToken relay123'
func main() {
	server := &tachyonVpnServer.Server{}
	udwConsole.MustRunCommandLineFromFuncV2(server.Run)
}
