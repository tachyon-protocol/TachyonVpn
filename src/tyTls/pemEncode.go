package tyTls

import (
	"encoding/pem"
	"crypto/rsa"
	"crypto/x509"
	"crypto/ecdsa"
	"fmt"
)

func EncodeObjToPemBlock(objI interface{}) (block *pem.Block,errMsg string){
	var b []byte
	switch obj:= objI.(type) {
	case *rsa.PrivateKey:
		b = x509.MarshalPKCS1PrivateKey(obj)
		return &pem.Block{
			Type: "RSA PRIVATE KEY",
			Bytes: b,
		},""
	case *rsa.PublicKey:
		PubASN1, err := x509.MarshalPKIXPublicKey(obj)
		if err!=nil{
			return nil,err.Error()
		}
		return &pem.Block{
			Type: "RSA PUBLIC KEY",
			Bytes: PubASN1,
		},""
	case *ecdsa.PrivateKey:
		b,err := x509.MarshalECPrivateKey(obj)
		if err!=nil{
			return nil,err.Error()
		}
		return &pem.Block{
			Type: "EC PRIVATE KEY",
			Bytes: b,
		},""
	case *x509.Certificate:
		return &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: obj.Raw,
		},""
	case *x509.CertificateRequest:
		return &pem.Block{
			Type:  "NEW CERTIFICATE REQUEST",
			Bytes: obj.Raw,
		},""
	default:
		return nil,fmt.Sprintf("[e8bpgvjnpm] unsupport type %T",objI)
	}
}

func MustEncodeObjToPemString(obj interface{})(pemContent string){
	p,errMsg:=EncodeObjToPemBlock(obj)
	if errMsg!=""{
		panic(errMsg)
	}
	b :=pem.EncodeToMemory(p)
	return string(b)
}