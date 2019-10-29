package main

import (
	"github.com/tachyon-protocol/udw/udwFile"
	"strings"
	"github.com/tachyon-protocol/udw/udwCmd"
	"path/filepath"
)

func main(){
	thisPath:=udwFile.MustGetFullPath("src/github.com/tachyon-protocol/udw")
	dirSet:=map[string]struct{}{}
	for _,fullpath:=range udwFile.MustGetAllFiles(thisPath){
		if strings.Contains(fullpath,"/.git"){
			continue
		}
		ext:=udwFile.GetExt(fullpath)
		if ext!=".go"{
			continue
		}
		dirSet[filepath.Dir(fullpath)] = struct{}{}
	}
	for fullpath:=range dirSet{
		rel:=udwFile.MustGetRelativePath(thisPath,fullpath)
		thisPkg:="github.com/tachyon-protocol/udw/"+rel
		udwCmd.MustRun("kmg goinstall "+thisPkg)
		udwCmd.MustRun("kmg go test -v -race "+thisPkg)
	}
}
