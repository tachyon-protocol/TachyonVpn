package main

import "github.com/tachyon-protocol/udw/udwGoSource/udwGoBuild"

func main(){
	udwGoBuild.MustBuild(udwGoBuild.BuildRequest{
		PkgPath: "make/serverCheckHelper",
	})
	udwGoBuild.MustBuild(udwGoBuild.BuildRequest{
		PkgPath: "make/serverCheckHelper",
		TargetOsCpuArch: udwGoBuild.TargetLinuxAmd64,
	})
}
