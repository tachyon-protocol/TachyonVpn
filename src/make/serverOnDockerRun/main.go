package main

import (
	"buildServerDockerImage"
	"github.com/tachyon-protocol/udw/udwCmd"
)

func main() {
	imageName := buildServerDockerImage.Build()
	const containerName = "tachyon-server"
	_ = udwCmd.Run(`docker container stop ` + containerName)
	_ = udwCmd.Run(`docker container rm ` + containerName)
	udwCmd.MustRun(`docker container run --publish 29443:29443 --privileged --cap-add=NET_ADMIN --device=/dev/net/tun --name ` + containerName + ` ` + imageName)
}
