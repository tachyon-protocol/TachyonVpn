package main

import (
	"buildRelease"
)

func main() {
	buildRelease.Build("make/client", "linux")
}
