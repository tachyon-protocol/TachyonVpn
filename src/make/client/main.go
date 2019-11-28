package main

import (
	"github.com/tachyon-protocol/udw/udwConsole"
	"tachyonVpnClient"
)

func main() {
	udwConsole.MustRunCommandLineFromFuncV2(tachyonVpnClient.ClientRun)
}
