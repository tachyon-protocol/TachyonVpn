package tachyonSimpleVpnProtocol

import (
	"fmt"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwChan"
	"github.com/tachyon-protocol/udw/udwErr"
	"testing"
	"time"
)

func TestInternalTlsConnection(t *testing.T) {
	conn := &internalConnection{
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
