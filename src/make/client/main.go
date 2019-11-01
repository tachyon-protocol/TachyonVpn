package main

import (
	"github.com/tachyon-protocol/udw/udwConsole"
	"tachyonVpnClient"
)

//kmg make sshDeploy -PkgPath make/server -Command client -ServerIp [ip]
//kmg make sshDeploy -PkgPath make/server -Command client -IsRelay -ServerIp [ip] -ExitClientId [clientId]
func main() {
	udwConsole.MustRunCommandLineFromFuncV2(tachyonVpnClient.ClientRun)
}
