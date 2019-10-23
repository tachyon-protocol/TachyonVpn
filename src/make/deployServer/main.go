package main

import (
	"fmt"
	"github.com/tachyon-protocol/udw/udwGoSource/udwGoBuild"
	"github.com/tachyon-protocol/udw/udwSsh"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 2 {
		panic("Usage: deployServer 123.123.123.123")
	}
	serverIp := os.Args[1]
	const (
		pkgPath = `make/server`
		_os     = `linux`
		arch    = `amd64`
	)
	resp := udwGoBuild.MustBuild(udwGoBuild.BuildRequest{
		PkgPath:       pkgPath,
		TargetOs:      _os,
		TargetCpuArch: arch,
		EnableRace:    false,
	})
	pkgName := filepath.Base(pkgPath)
	fmt.Println("build successfully", pkgName, _os, "/", arch)
	udwSsh.MustScpToRemoteDefault(serverIp, resp.GetOutputExeFilePath(), "/usr/local/bin/"+pkgName)
	//udwSsh.MustRpcSshDefault(serverIp, pkgName)
}
