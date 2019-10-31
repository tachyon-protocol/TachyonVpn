package main

import tachyonVpnClient "tachyonVpnServer"

//kmg make sshDeploy -PkgPath make/server -Command server -Ip [Your server's IP]
func main() {
	tachyonVpnClient.ServerRun()
}
