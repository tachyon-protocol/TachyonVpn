package tyTls

import (
	"encoding/pem"
	"crypto/rsa"
	"crypto/x509"
	"crypto/ecdsa"
	"bytes"
	"strings"
	"fmt"
)

func DecodePemContentCallback(content []byte,cb func(obj interface{})) (errMsg string){
	bw:=bytes.Buffer{}
	for _,line:=range strings.Split(string(content),"\n"){
		line = strings.TrimSpace(line)
		if line==""{
			continue
		}
		bw.WriteString(line)
		bw.WriteByte('\n')
	}
	content=bw.Bytes()
	var p *pem.Block
	for {
		p, content = pem.Decode(content)
		if p == nil {
			break
		}
		key,errMsg:= DecodeObjFromPemBlock(p)
		if errMsg!=""{
			return errMsg
		}
		cb(key)
	}
	return ""
}

func DecodeOneObjFromPemContent(content []byte)(obj interface{},errMsg string){
	errMsg = ""
	DecodePemContentCallback(content,func(thisObj interface{}){
		if obj!=nil{
			errMsg="c7tngmegbm"
			return
		}
		obj = thisObj
	})
	if obj==nil{
		return nil,"v7t9tymwb6"
	}
	return obj,errMsg
}

func DecodeOneCertFromPemContent(content []byte) (cert *x509.Certificate,errMsg string) {
	obj1,errMsg:=DecodeOneObjFromPemContent(content)
	if errMsg!=""{
		return nil,errMsg
	}
	cert,ok:=obj1.(*x509.Certificate)
	if !ok{
		return nil,"3cb7nnjbth "+fmt.Sprintf("%T",obj1)
	}
	return cert,""
}

/*
possible return type:
*rsa.PrivateKey
*ecdsa.PrivateKey
*x509.Certificate
*x509.CertificateRequest
 */
func DecodeObjFromPemBlock(p *pem.Block) (obj interface{},errMsg string){
	switch p.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(p.Bytes)
		if err != nil {
			return nil,err.Error()
		}
		return key,""
	case "EC PRIVATE KEY":
		key, err := x509.ParseECPrivateKey(p.Bytes)
		if err != nil {
			return nil,err.Error()
		}
		return key,""
	case "PRIVATE KEY":
		// copy from /usr/local/go/src/crypto/tls/tls.go:279 tls.parsePrivateKey
		key1, err1 := x509.ParsePKCS1PrivateKey(p.Bytes)
		if err1==nil{
			return key1,""
		}
		key2, err2 := x509.ParseECPrivateKey(p.Bytes)
		if err2==nil {
			return key2,""
		}
		key3, err3 := x509.ParsePKCS8PrivateKey(p.Bytes)
		if err3 == nil {
			switch key3.(type) {
			case *rsa.PrivateKey, *ecdsa.PrivateKey:
				return key3,""
			default:
				return key3,"f8vahfhkgp tls: found unknown private key type in PKCS#8 wrapping"
			}
		}
		return nil,"yrng5ge2bj rsa:"+err1.Error()+"\n ec:"+err2.Error()
	case "CERTIFICATE":
		cert, err := x509.ParseCertificate(p.Bytes)
		if err != nil {
			return nil,err.Error()
		}
		return cert,""
	case "NEW CERTIFICATE REQUEST":
		csr, err := x509.ParseCertificateRequest(p.Bytes)
		if err != nil {
			return nil,err.Error()
		}
		return csr,""
	default:
		return nil,"not expect pem type ["+p.Type+"]"
	}
}

