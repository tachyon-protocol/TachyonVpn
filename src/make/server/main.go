package main

import (
	"github.com/tachyon-protocol/udw/udwConsole"
	"tachyonVpnServer"
)

//relay server
//sshDeploy -PkgPath make/server -Ip ip -Command 'server -SelfTKey relay123'
//vpe server
//sshDeploy -PkgPath make/server -Ip vpeServerIp -Command 'server -SelfTKey exit123 -UseRelay -RelayServerIp relayServerIp -RelayServerTKey relay123'
func main() {
	server := &tachyonVpnServer.Server{}
	udwConsole.MustRunCommandLineFromFuncV2(server.Run)
}
