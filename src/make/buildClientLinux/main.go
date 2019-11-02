package main

import (
	"make/buildRelease"
)

func main() {
	buildRelease.Build("make/client", "linux")
}
