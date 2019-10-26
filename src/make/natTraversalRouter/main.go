package main

import (
	"fmt"
	"github.com/tachyon-protocol/udw/udwErr"
	"net"
)

//kmg make sshDeploy -PkgPath make/server -Command server -Ip [Your server's IP]
func main() {
	ln, err := net.Listen("tcp", ":7398")
	udwErr.PanicIfError(err)
	for {
		conn, err := ln.Accept()
		udwErr.PanicIfError(err)
		fmt.Println("accept new conn from", conn.RemoteAddr())
		bufR := make([]byte, 1<<20)
		go func() {
			n, err := conn.Read(bufR)
			if err != nil {
				conn.Close()
				return
			}
			fmt.Println("read ", n, "from", conn.RemoteAddr())
		}()
	}
}
