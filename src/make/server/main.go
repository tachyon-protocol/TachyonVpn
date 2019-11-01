package main

import (
	"github.com/tachyon-protocol/udw/udwConsole"
	tachyonVpnClient "tachyonVpnServer"
)

//kmg make sshDeploy -PkgPath make/server -Command server -Ip [Your server's IP]
func main() {
	server := &tachyonVpnClient.Server{}
	udwConsole.MustRunCommandLineFromFuncV2(server.Run)
}
