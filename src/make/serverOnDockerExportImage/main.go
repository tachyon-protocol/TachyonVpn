package main

import (
	"buildServerDockerImage"
	"fmt"
	"github.com/tachyon-protocol/udw/udwCmd"
	"github.com/tachyon-protocol/udw/udwProjectPath"
)

func main() {
	imageName := buildServerDockerImage.Build()
	imageOutputPath := udwProjectPath.MustPathInProject("bin/" + imageName + ".image")
	udwCmd.MustRun(`docker image save -o ` + imageOutputPath + ` ` + imageName)
	fmt.Println("- - - \nexport image:")
	fmt.Println(imageOutputPath)
}
