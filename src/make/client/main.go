package main

import (
	"github.com/tachyon-protocol/udw/udwConsole"
	"tachyonVpnClient"
)

//sshDeploy -PkgPath make/client -Ip clientIp -Command 'client -ServerIp serverIp -ServerToken relay123'
//sshDeploy -PkgPath make/client -Ip clientIp -Command 'client -IsRelay -ServerIp relayServerIp -ServerToken relay123 -ExitServerToken exit123 -ExitServerClientId clientId'
func main() {
	client := tachyonVpnClient.Client{}
	udwConsole.MustRunCommandLineFromFuncV2(client.Run)
}
