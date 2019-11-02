package tachyonVpnProtocol

import (
	"crypto/tls"
	"fmt"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwChan"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwMath"
	"github.com/tachyon-protocol/udw/udwRand"
	"github.com/tachyon-protocol/udw/udwTlsSelfSignCertV2"
	"strings"
	"testing"
)

func TestInternalConnectionSingle(t *testing.T) {
	conn := &internalConnectionSingle{
		pipe: udwChan.MakeChanBytes(0),
		bufR: udwBytes.NewBufWriter(nil),
	}
	const (
		repeat     = 10
		inputLen   = 15 << 10
		readBufLen = 1 << 10
	)
	var loopLen = inputLen*repeat/udwMath.IntMin(readBufLen, inputLen) + 1
	var inputStr = udwRand.MustCryptoRandToReadableAlpha(inputLen)
	go func() {
		for i := 0; i < repeat; i++ {
			_, err := conn.Write([]byte(inputStr))
			udwErr.PanicIfError(err)
		}
	}()
	buf := make([]byte, readBufLen)
	var outputStr string
	for i := 0; i < loopLen; i++ {
		n, err := conn.Read(buf)
		udwErr.PanicIfError(err)
		outputStr += string(buf[:n])
		if strings.Repeat(inputStr, repeat) == outputStr {
			fmt.Println("test passed ✔")
			return
		}
	}
	panic("test failed")
}

func TestNewInternalConnectionDualTls(t *testing.T) {
	left, right := NewInternalConnectionDual()
	client := tls.Client(left, &tls.Config{
		InsecureSkipVerify: true,
	})
	server := tls.Server(right, &tls.Config{
		Certificates: []tls.Certificate{
			*udwTlsSelfSignCertV2.GetTlsCertificate(),
		},
		InsecureSkipVerify: true,
	})
	const (
		repeat     = 5
		inputLen   = 1500
		readBufLen = 3 << 10
	)
	var loopLen = inputLen*repeat/udwMath.IntMin(readBufLen, inputLen) + 1
	var inputStr = udwRand.MustCryptoRandToReadableAlpha(inputLen)
	go func() {
		for i := 0; i < repeat; i++ {
			_, err := client.Write([]byte(inputStr))
			udwErr.PanicIfError(err)
		}
	}()
	buf := make([]byte, readBufLen)
	var outputStr string
	for i := 0; i < loopLen; i++ {
		n, err := server.Read(buf)
		udwErr.PanicIfError(err)
		outputStr += string(buf[:n])
		if strings.Repeat(inputStr, repeat) == outputStr {
			fmt.Println("test passed ✔")
			return
		}
	}
	panic("test failed")
}

func TestNewInternalConnectionDualDoubleLayers(t *testing.T) {
	client, server := NewInternalConnectionDual()
	client = tls.Client(client, &tls.Config{
		ServerName:         udwRand.MustCryptoRandToReadableAlpha(5) + ".com",
		InsecureSkipVerify: true,
		NextProtos:         []string{"http/1.1", "h2"},
	})
	client = tls.Client(client, &tls.Config{
		ServerName:         udwRand.MustCryptoRandToReadableAlpha(5) + ".com",
		InsecureSkipVerify: true,
		NextProtos:         []string{"http/1.1", "h2"},
	})
	server = tls.Server(server, &tls.Config{
		Certificates: []tls.Certificate{
			*udwTlsSelfSignCertV2.GetTlsCertificate(),
		},
		NextProtos: []string{"http/1.1"},
	})
	server = tls.Server(server, &tls.Config{
		Certificates: []tls.Certificate{
			*udwTlsSelfSignCertV2.GetTlsCertificate(),
		},
		NextProtos: []string{"http/1.1"},
	})
	var inputStr = udwRand.MustCryptoRandToReadableAlpha(1 << 10)
	go func() {
		_, err := client.Write([]byte(inputStr))
		udwErr.PanicIfError(err)
		fmt.Println("test passed ✔")
	}()
	buf := make([]byte, 1<<8)
	var outputStr string
	testPass := false
	for i := 0; i < 100; i++ {
		n, err := server.Read(buf)
		udwErr.PanicIfError(err)
		outputStr += string(buf[:n])
		if outputStr == inputStr {
			testPass = true
			break
		}
	}
	if !testPass {
		panic("test failed")
	}
}
