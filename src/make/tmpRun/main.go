package main

import (
	"crypto/tls"
	"fmt"
	"github.com/tachyon-protocol/udw/udwConsole"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwRand"
	"github.com/tachyon-protocol/udw/udwTlsSelfSignCertV2"
	"io"
	"net"
	"strings"
	"tachyonVpnProtocol"
)

func main() {
	const addr = "127.0.0.1:8080"
	ln, err := net.Listen("tcp", addr)
	udwErr.PanicIfError(err)
	go func() {
		for {
			conn, err := ln.Accept()
			udwErr.PanicIfError(err)
			var _conn net.Conn
			debugInternalConnection := true
			if !debugInternalConnection {
				_conn = conn
			} else {
				rBwA, rAwB := tachyonVpnProtocol.NewInternalConnectionDual(nil, nil)
				_conn = rAwB
				go func() {
					buf := make([]byte, 10<<20)
					for {
						n, err := rBwA.Read(buf)
						udwErr.PanicIfError(err)
						_, err = conn.Write(buf[:n])
						udwErr.PanicIfError(err)
					}
				}()
				go func() {
					buf := make([]byte, 10<<20)
					for {
						n, err := conn.Read(buf)
						if err == io.EOF {
							fmt.Println("EOF")
							continue
						}
						udwErr.PanicIfError(err)
						_, err = rBwA.Write(buf[:n])
						udwErr.PanicIfError(err)
					}
				}()
			}
			_conn = tls.Server(_conn, &tls.Config{
				Certificates: []tls.Certificate{
					*udwTlsSelfSignCertV2.GetTlsCertificate(),
				},
				NextProtos:         []string{"http/1.1"},
				InsecureSkipVerify: true,
			})
			go func() {
				buf := make([]byte, 3<<10)
				for {
					n, err := _conn.Read(buf)
					udwErr.PanicIfError(err)
					fmt.Println(string(buf[:n]))
				}
			}()
		}
	}()

	conn, err := net.Dial("tcp", addr)
	udwErr.PanicIfError(err)
	conn = tls.Client(conn, &tls.Config{
		ServerName:         udwRand.MustCryptoRandToReadableAlpha(5) + ".com",
		InsecureSkipVerify: true,
		//NextProtos:         []string{"http/1.1", "h2"},
	})
	for i := 0; i < 10; i++ {
		_, err = conn.Write([]byte(strings.Repeat("1", 1<<10)))
		udwErr.PanicIfError(err)
	}
	udwConsole.WaitForExit()
}
