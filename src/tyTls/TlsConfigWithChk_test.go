package tyTls

import (
	"testing"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwNet/udwNetTestV2"
	"crypto/tls"
)

func TestTlsConfigWithChk(ot *testing.T){
	EnableTlsVersion13()
	certS := NewTlsCert(false)
	ServerChk :=HashChk(certS.Certificate[0])
	//fmt.Println("ServerChk",ServerChk)
	certC:= NewTlsCert(true)
	ClientChk :=HashChk(certC.Certificate[0])
	//fmt.Println("ClientChk",ClientChk)
	{
		cc,errMsg:=NewClientTlsConfigWithChk(NewClientTlsConfigWithChkReq{
			ServerChk: ServerChk,
			ClientCert: certC,
		})
		udwErr.PanicIfErrorMsg(errMsg)
		sc,errMsg:=NewServerTlsConfigWithChk(NewServerTlsConfigWithChkReq{
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

