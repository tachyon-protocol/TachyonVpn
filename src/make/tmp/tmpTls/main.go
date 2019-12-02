package main

import (
	"crypto/tls"
	"tyTls"
	"fmt"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwNet/udwNetTestV2"
	"crypto/rsa"
	"crypto/x509"
	"crypto/ecdsa"
)

func main(){
	tyTls.EnableTlsVersion13()
	certS :=tyTls.NewTlsCert(false)
	ServerChk :=tyTls.HashChk(certS.Certificate[0])
	fmt.Println("ServerChk",ServerChk)
	certC:=tyTls.NewTlsCert(true)
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
		udwNetTestV2.RunTestTwoRwc(tlsC,tlsS)
	}
}

func CertMarshal(cert *tls.Certificate) []byte{

}

func CertUnmarshal(b []byte) (cert *tls.Certificate){

}

func marshalPrivateKey(objI interface{}) (b []byte,errMsg string){
	switch obj:= objI.(type) {
	case *rsa.PrivateKey:
		b := x509.MarshalPKCS1PrivateKey(obj)
		return b,""
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(obj)
		if err != nil {
			return nil, err.Error()
		}
		return b,""
	default:
		return nil,"unknow privateKey type"
	}
}