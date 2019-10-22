package main

import "os"

func main() {
	if len(os.Args) != 2 {
		panic("Usage: deployServer 123.123.123.123")
	}
}
