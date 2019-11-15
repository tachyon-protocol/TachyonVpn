package buildRelease

import (
	"github.com/tachyon-protocol/udw/udwFile"
	"github.com/tachyon-protocol/udw/udwGoSource/udwGoBuild"
	"path/filepath"
)

func Build(pkg string, os string) string {
	resp := udwGoBuild.MustBuild(udwGoBuild.BuildRequest{
		PkgPath:       pkg,
		TargetOs:      os,
		TargetCpuArch: `amd64`,
		EnableRace:    false,
	})
	udwFile.MustCopy(resp.GetOutputExeFilePath(), filepath.Join(udwFile.MustGetHomeDirPath(), "Downloads", filepath.Base(pkg)+"_"+os))
	return resp.GetOutputExeFilePath()
}
