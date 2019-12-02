package main

import (
	"github.com/tachyon-protocol/udw/udwConsole"
	"tachyonVpnClient"
)

func main() {
	client := tachyonVpnClient.Client{}
	udwConsole.MustRunCommandLineFromFuncV2(client.Run)
}
