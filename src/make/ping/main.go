package main

import (
	"github.com/tachyon-protocol/udw/udwConsole"
	"github.com/tachyon-protocol/udw/udwErr"
	"tachyonVpnClient"
)

func main() {
	udwConsole.MustRunCommandLineFromFuncV2(func(req struct{
		Ip string
		Count int
	}) {
		err := tachyonVpnClient.Ping(tachyonVpnClient.PingReq{
			Ip:       req.Ip,
			Count:    req.Count,
			DebugLog: true,
		})
		udwErr.PanicIfError(err)
	})
}
