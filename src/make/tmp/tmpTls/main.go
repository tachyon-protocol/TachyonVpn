package main

import (
	"crypto/tls"
	"tyTls"
	"fmt"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwNet/udwNetTestV2"
	"crypto/x509"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwCryptoSha3"
	"encoding/base64"
	"encoding/pem"
	"bytes"
)

type ClientToken struct{
	ServerChk string
	ClientCert string
}

func main(){
	tyTls.EnableTlsVersion13()
	{
		cert_1:=NewTlsCert(false,"27cz7epj6m")
		s_1,errMsg:=CertMarshal(&cert_1)
		udwErr.PanicIfErrorMsg(errMsg)
		cert_2:=NewTlsCert(false,"27cz7epj6m")
		s_2,errMsg:=CertMarshal(&cert_2)
		udwErr.PanicIfErrorMsg(errMsg)
		fmt.Println(s_1==s_2)
	}
	certS :=tyTls.NewTlsCert(false)
	ServerChk :=tyTls.HashChk(certS.Certificate[0])
	s,errMsg:=CertMarshal(&certS)
	udwErr.PanicIfErrorMsg(errMsg)
	fmt.Println("ServerCert",s,len(s))
	certS_1,errMsg:=CertUnmarshal(s)
	udwErr.PanicIfErrorMsg(errMsg)
	certS=*certS_1

	fmt.Println("ServerChk",ServerChk,len(ServerChk))
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

func CertMarshal(cert *tls.Certificate) (s string,errMsg string){
	buf:=udwBytes.BufWriter{}
	pkB,err:=x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	if err!=nil{
		return "",err.Error()
	}
	buf.WriteUvarint(uint64(len(pkB)))
	buf.Write_(pkB)
	buf.WriteUvarint(uint64(len(cert.Certificate)))
	for _, certRaw:=range cert.Certificate{
		buf.WriteUvarint(uint64(len(certRaw)))
		buf.Write_(certRaw)
	}
	sum:=udwCryptoSha3.Sum512Slice(buf.GetBytes())
	buf.Write_(sum[:4])
	return base64.RawURLEncoding.EncodeToString(buf.GetBytes()),""
}

func CertUnmarshal(s string) (tlsCert *tls.Certificate,errMsg string){
	b,err:=base64.RawURLEncoding.DecodeString(s)
	if err!=nil{
		return nil,err.Error()
	}
	reader:=udwBytes.NewBufReader(b)
	pkSize,ok:=reader.ReadUvarint()
	if !ok|| pkSize<=1{
		return nil,"gmfp28u374"
	}
	pkB,ok:=reader.ReadByteSlice(int(pkSize))
	if !ok{
		return nil,"nkt4xh9mfe"
	}
	pkPem:=pem.EncodeToMemory(&pem.Block{
		Type: "PRIVATE KEY",
		Bytes: pkB,
	})
	certListLen,ok:=reader.ReadUvarint()
	if !ok{
		return nil,"yetwm28kyj"
	}
	certBuf:=udwBytes.BufWriter{}
	for i:=0;i<int(certListLen);i++{
		thisCertLen,ok:=reader.ReadUvarint()
		if !ok{
			return nil,"bbemaktzw4"
		}
		certB,ok:=reader.ReadByteSlice(int(thisCertLen))
		if !ok{
			return nil,"m9y5us7quv"
		}
		certBuf.Write_(pem.EncodeToMemory(&pem.Block{
			Type: "CERTIFICATE",
			Bytes: certB,
		}))
	}
	pos:=reader.GetPos()
	shouldSum,ok:=reader.ReadByteSlice(4)
	if !ok{
		return nil,"h4psrrv3r7"
	}
	sum:=udwCryptoSha3.Sum512Slice(b[:pos])
	if bytes.Equal(sum[:4],shouldSum)==false{
		return nil,"kr4xx74xub"
	}
	tlsCert_1,err:=tls.X509KeyPair(certBuf.GetBytes(),pkPem)
	if err!=nil{
		return nil,err.Error()
	}
	return &tlsCert_1,""
}

