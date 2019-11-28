package tyTls

import (
	"crypto/x509"
	"crypto/tls"
	"time"
	"fmt"
)

func GetClientTlsConfigServerCertPem(serverCertPem string) (config *tls.Config,errMsg string){
	cert,errMsg:=DecodeOneCertFromPemContent([]byte(serverCertPem))
	if errMsg!=""{
		return nil,errMsg
	}
	return GetClientTlsConfigFromServerX509Cert(cert)
}

func GetClientTlsConfigFromServerX509Cert(cert *x509.Certificate) (config *tls.Config,errMsg string){
	certPool:=x509.NewCertPool()
	certPool.AddCert(cert)
	getSnFn:=func()string{
		if len(cert.DNSNames)>0{
			return cert.DNSNames[0]
		}
		if len(cert.IPAddresses)>0{
			return cert.IPAddresses[0].String()
		}
		if cert.Subject.CommonName!=""{
			return cert.Subject.CommonName
		}
		return ""
	}
	sn:=getSnFn()
	if sn==""{
		return nil,"certificate is not valid for any names"
	}
	// avoid time problem. if it can use current time, it will use.
	timeFn:=func()time.Time{
		t:=time.Now()
		fmt.Println(t)
		if t.Before(cert.NotBefore) || t.After(cert.NotAfter){
			t = cert.NotBefore.Add(cert.NotAfter.Sub(cert.NotBefore)/2)
			fmt.Println(t)
			return t
		}
		return t
	}
	// can use Time field to avoid certificate time invalid.
	tlsConfig:=&tls.Config{
		Time: timeFn,
		ServerName: sn,
		RootCAs: certPool,
		MinVersion: tls.VersionTLS12,
	}
	return tlsConfig,""
}

func GetX509CertFromTlsCert(tlsCert *tls.Certificate) (cert *x509.Certificate,errMsg string){
	if len(tlsCert.Certificate)==0{
		return nil,"3eavg6xrec"
	}
	x509Cert,err:=x509.ParseCertificate(tlsCert.Certificate[0])
	if err!=nil{
		return nil,err.Error()
	}
	return x509Cert,""
}