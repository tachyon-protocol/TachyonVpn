package buildServerDockerImage

import (
	"buildRelease"
	"github.com/tachyon-protocol/udw/udwCmd"
	"github.com/tachyon-protocol/udw/udwFile"
	"github.com/tachyon-protocol/udw/udwProjectPath"
	"github.com/tachyon-protocol/udw/udwRand"
	"path/filepath"
)

func Build() (_imageName string) {
	binPath := buildRelease.Build("make/server", "linux")
	buildPath := udwProjectPath.MustPathInProject("tmp/buildDockerImage_" + udwRand.MustCryptoRandToReadableAlpha(5))
	udwFile.MustMkdir(buildPath)
	defer udwFile.MustDelete(buildPath)
	udwFile.MustCopy(binPath, filepath.Join(buildPath, "server"))
	udwFile.MustCopy(udwProjectPath.MustPathInProject("src/buildServerDockerImage/Dockerfile"), filepath.Join(buildPath, "Dockerfile"))
	udwFile.MustSetWd(buildPath)
	const (
		imageVersion = "1"
		imageName    = "tachyon-server-on-docker:" + imageVersion
	)
	udwCmd.MustRun(`docker image build -t ` + imageName + ` .`)
	return imageName
}
