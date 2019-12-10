package main

import (
	"github.com/tachyon-protocol/udw/udwConsole"
	"github.com/tachyon-protocol/udw/tyVpnClient"
)

//sshDeploy -PkgPath make/client -Ip clientIp -Command 'client -ServerIp serverIp -ServerTKey relay123'
//sshDeploy -PkgPath make/client -Ip clientIp -Command 'client -IsRelay -ServerIp relayServerIp -ServerTKey relay123 -ExitServerTKey exit123 -ExitServerClientId clientId'
func main() {
	client := tyVpnClient.Client{}
	udwConsole.MustRunCommandLineFromFuncV2(client.Run)
}
