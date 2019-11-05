package main

import "make/buildRelease"

func main() {
	buildRelease.Build("make/client", "darwin")
}
