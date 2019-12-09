package main

import (
	"github.com/tachyon-protocol/udw/udwFile"
	"strings"
	"github.com/tachyon-protocol/udw/udwCmd"
	"path/filepath"
)

func main(){
	tryGoInstall("make/client")
	tryGoInstall("make/server")
	tryGoInstall("tachyonVpnRouteServer")
	thisPath:=udwFile.MustGetFullPath("src/github.com/tachyon-protocol/udw")
	dirSet:=map[string]struct{}{}
	for _, fullPath :=range udwFile.MustGetAllFiles(thisPath){
		if strings.Contains(fullPath,"/.git"){
			continue
		}
		ext:=udwFile.GetExt(fullPath)
		if ext!=".go"{
			continue
		}
		dirSet[filepath.Dir(fullPath)] = struct{}{}
	}
	for fullpath:=range dirSet{
		rel:=udwFile.MustGetRelativePath(thisPath,fullpath)
		thisPkg:="github.com/tachyon-protocol/udw/"+rel
		tryGoInstall(thisPkg)
		udwCmd.CmdString("go test -v -race "+thisPkg).MustSetEnv("GOPATH",udwFile.MustGetWd()).MustRun()
	}
}

func tryGoInstall(pkgPath string) {
	udwCmd.CmdString("go install "+pkgPath).MustSetEnv("GOPATH",udwFile.MustGetWd()).MustRun()
}