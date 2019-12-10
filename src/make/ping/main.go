package main

import (
	"github.com/tachyon-protocol/udw/udwConsole"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/tyVpnClient"
)

func main() {
	udwConsole.MustRunCommandLineFromFuncV2(func(req tyVpnClient.PingReq) {
		err := tyVpnClient.Ping(req)
		udwErr.PanicIfError(err)
	})
}
