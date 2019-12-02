package main

import (
	"crypto/tls"
	"tyTls"
	"fmt"
	"sync"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwTest"
	"github.com/tachyon-protocol/udw/udwNet/udwNetTestV2"
	"net"
)

func main(){
	tyTls.EnableTlsVersion13()
	certS :=newCert(false)
	ServerChk :=tyTls.HashChk(certS.Certificate[0])
	fmt.Println("ServerChk",ServerChk)
	certC:=newCert(true)
	ClientChk :=tyTls.HashChk(certC.Certificate[0])
	fmt.Println("ClientChk",ClientChk)
	{
		cc,errMsg:=tyTls.NewClientTlsConfigWithChk(tyTls.NewClientTlsConfigWithChkReq{
			ServerChk: ServerChk,
			ClientCert: certC,
		})
		udwErr.PanicIfErrorMsg(errMsg)
		sc,errMsg:=tyTls.NewServerTlsConfigWithChk(tyTls.NewServerTlsConfigWithChkReq{
			ClientChk: ClientChk,
			ServerCert: certS,
		})
		udwErr.PanicIfErrorMsg(errMsg)

		c1,c2:=udwNetTestV2.MustTcpPipe()
		tlsC:=tls.Client(c1,cc)
		tlsS:=tls.Server(c2,sc)
		TestTwoNetConn(tlsC,tlsS)
	}
}

func TestTwoNetConn(tlsC net.Conn,tlsS net.Conn){
	wg:=sync.WaitGroup{}
	wg.Add(1)
	go func(){
		buf:=make([]byte,4096)
		nr,err:=tlsS.Read(buf)
		fmt.Println("5")
		udwErr.PanicIfError(err)
		udwTest.Equal(buf[:nr],[]byte{1})
		wg.Done()
	}()
	_,err:=tlsC.Write([]byte{1})
	udwErr.PanicIfError(err)
	fmt.Println("3")
	wg.Wait()
	for i:=0;i<10;i++{
		_,err:=tlsC.Write([]byte{1})
		udwErr.PanicIfError(err)
		buf:=make([]byte,4096)
		nr,err:=tlsS.Read(buf)
		udwErr.PanicIfError(err)
		udwTest.Equal(buf[:nr],[]byte{1})
	}
	tlsC.Close()
	tlsS.Close()
}