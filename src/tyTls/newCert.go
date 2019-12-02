package tyTls

import (
	"crypto/tls"
	"crypto/x509"
	"time"
	"math/big"
	"net"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/pem"
	"crypto/rand"
)

func NewTlsCert(isClient bool) (cert tls.Certificate){
	var ExtKeyUsage x509.ExtKeyUsage
	if isClient{
		ExtKeyUsage = x509.ExtKeyUsageClientAuth
	}else{
		ExtKeyUsage = x509.ExtKeyUsageServerAuth
	}
	const dur = 100*365*24*time.Hour
	startTime:=time.Now()
	notBefore:=startTime.Add(-dur)
	notAfter:=startTime.Add(dur)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore: notBefore,
		NotAfter:  notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{ExtKeyUsage},
		BasicConstraintsValid: true,
	}
	if isClient==false{
		template.IPAddresses = []net.IP{net.IPv4(127,0,0,1)}
	}
	//template.DNSNames = []string{"google.com"}
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err!=nil{
		panic(err)
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		panic(err)
	}
	certPem:=pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE",
		Bytes: derBytes,
	})
	b,err := x509.MarshalECPrivateKey(priv)
	if err!=nil{
		panic(err)
	}
	privPem:=pem.EncodeToMemory(&pem.Block{
		Type: "EC PRIVATE KEY",
		Bytes: b,
	})
	cert, err = tls.X509KeyPair(certPem, privPem)
	if err != nil {
		panic(err)
	}
	return cert
}