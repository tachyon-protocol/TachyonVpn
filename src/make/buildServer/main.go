package main

import (
	"fmt"
	"github.com/tachyon-protocol/udw/udwGoSource/udwGoBuild"
)

func main() {
	resp := udwGoBuild.MustBuild(udwGoBuild.BuildRequest{
		PkgPath:       `make/server`,
		TargetOs:      `linux`,
		TargetCpuArch: `amd64`,
		EnableRace:    false,
	})
	fmt.Println(resp.GetOutputExeFilePath())
}
