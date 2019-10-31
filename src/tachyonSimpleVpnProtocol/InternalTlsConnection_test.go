package tachyonSimpleVpnProtocol

import (
	"crypto/tls"
	"fmt"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwChan"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwRand"
	"github.com/tachyon-protocol/udw/udwTlsSelfSignCertV2"
	"testing"
	"time"
)

func TestInternalConnectionSingle(t *testing.T) {
	conn := &internalConnectionSingle{
		pipe: udwChan.MakeChanBytes(0),
		buf:  udwBytes.NewBufWriter(nil),
	}
	const inputStr = "123456789"
	go func() {
		time.Sleep(time.Millisecond * 300)
		_, err := conn.Write([]byte(inputStr))
		udwErr.PanicIfError(err)
	}()
	buf := make([]byte, 2)
	var outputStr string
	testPass := false
	for i := 0; i < 100; i++ {
		n, err := conn.Read(buf)
		udwErr.PanicIfError(err)
		outputStr += string(buf[:n])
		fmt.Println(outputStr, n)
		if outputStr == inputStr {
			testPass = true
			break
		}
	}
	if !testPass {
		panic("test failed")
	}
}

func TestNewInternalConnectionDual(t *testing.T) {
	left, right := NewInternalConnectionDual()
	client := tls.Client(left, &tls.Config{
		ServerName:         udwRand.MustCryptoRandToReadableAlpha(5) + ".com",
		InsecureSkipVerify: true,
		NextProtos:         []string{"http/1.1", "h2"},
	})
	server := tls.Server(right, &tls.Config{
		Certificates: []tls.Certificate{
			*udwTlsSelfSignCertV2.GetTlsCertificate(),
		},
		NextProtos: []string{"http/1.1"},
	})
	var inputStr = udwRand.MustCryptoRandToReadableAlpha(1<<10)
	go func() {
		_, err := client.Write([]byte(inputStr))
		udwErr.PanicIfError(err)
		fmt.Println("written ✔")
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
