package main

import (
	"github.com/tachyon-protocol/udw/udwConsole"
	"github.com/tachyon-protocol/udw/udwErr"
	"tachyonVpnClient"
)

func main() {
	udwConsole.MustRunCommandLineFromFuncV2(func(req tachyonVpnClient.PingReq) {
		err := tachyonVpnClient.Ping(req)
		udwErr.PanicIfError(err)
	})
}
