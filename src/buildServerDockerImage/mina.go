package buildServerDockerImage

import (
	"buildRelease"
	"github.com/tachyon-protocol/udw/udwCmd"
	"github.com/tachyon-protocol/udw/udwFile"
	"github.com/tachyon-protocol/udw/udwProjectPath"
	"github.com/tachyon-protocol/udw/udwRand"
	"path/filepath"
)

/**
docker swarm init
docker container list -a

docker image ls
docker image prune
docker image build -t tachyon-server-on-docker:1 .
docker container rm tachyon-server;docker container run --publish 29443:29443 --privileged --cap-add=NET_ADMIN --device=/dev/net/tun --name tachyon-server tachyon-server-on-docker:1
docker exec -it tachyon-server /bin/bash
docker container logs tachyon-server

docker image load -i path/to/tachyon-server-on-docker.image
docker container run --publish 29443:29443 --privileged --cap-add=NET_ADMIN --device=/dev/net/tun --name tachyon-server tachyon-server-on-docker:1

Relay Mode
docker container run --publish 29443:29443 --privileged --cap-add=NET_ADMIN --device=/dev/net/tun --name tachyon-server tachyon-server-on-docker:1 server -UseRelay -RelayServerIp
 */

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
