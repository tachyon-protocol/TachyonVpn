package tyTls

import (
	"crypto/x509"
	"crypto/tls"
	"errors"
)

type NewClientTlsConfigWithChkReq struct{
	ServerChk string // must have
	ClientCert tls.Certificate // can be nil
}
func NewClientTlsConfigWithChk(req NewClientTlsConfigWithChkReq) (config *tls.Config,errMsg string){
	if IsChkValid(req.ServerChk)==false{
		return nil,"sxw489xmmp chk is not valid"
	}
	tlsConfig:=&tls.Config{
		InsecureSkipVerify: true,
		VerifyPeerCertificate:func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error{
			if len(rawCerts)==0{
				return errors.New("7wk8yam67e")
			}
			thisCertB:=rawCerts[0]
			x509Cert,err:=x509.ParseCertificate(thisCertB)
			if err!=nil{
				return err
			}
			if checkExtKeyUsageContain(x509.ExtKeyUsageServerAuth,x509Cert)==false{
				return errors.New("ftsh78pkc9")
			}
			chk1:=HashChk(thisCertB)
			if chk1!=req.ServerChk{
				return errors.New("ntexabpdw4")
			}
			return nil
		},
		MinVersion: tls.VersionTLS12,
	}
	if len(req.ClientCert.Certificate)>0{
		x509Cert,errMsg:=GetX509CertFromTlsCert(&req.ClientCert)
		if errMsg!=""{
			return nil,errMsg
		}
		if checkExtKeyUsageContain(x509.ExtKeyUsageClientAuth,x509Cert)==false{
			return nil,"up5ja99hfz"
		}
		tlsConfig.Certificates = []tls.Certificate{req.ClientCert}
	}
	return tlsConfig,""
}

type NewServerTlsConfigWithChkReq struct{
	ClientChk string // can be nil
	ServerCert tls.Certificate // must have
}
func NewServerTlsConfigWithChk(req NewServerTlsConfigWithChkReq) (config *tls.Config,errMsg string){
	if req.ClientChk!="" && IsChkValid(req.ClientChk)==false{
		return nil,"hp6pngedds chk is not valid"
	}
	x509Cert,errMsg:=GetX509CertFromTlsCert(&req.ServerCert)
	if errMsg!=""{
		return nil,errMsg
	}
	if checkExtKeyUsageContain(x509.ExtKeyUsageServerAuth,x509Cert)==false{
		return nil,"sfs6tvydkm"
	}
	tlsConfig:=&tls.Config{
		Certificates: []tls.Certificate{req.ServerCert},
		MinVersion: tls.VersionTLS12,
	}
	if req.ClientChk!=""{
		if IsChkValid(req.ClientChk)==false{
			return nil,"sxw489xmmp chk is not valid"
		}
		tlsConfig.ClientAuth = tls.RequestClientCert
		tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error{
			if len(rawCerts)==0{
				return errors.New("7wk8yam67e")
			}
			thisCertB:=rawCerts[0]
			x509Cert,err:=x509.ParseCertificate(thisCertB)
			if err!=nil{
				return err
			}
			if checkExtKeyUsageContain(x509.ExtKeyUsageClientAuth,x509Cert)==false{
				return errors.New("jve5m9gusv")
			}
			chk1:=HashChk(thisCertB)
			if chk1!=req.ClientChk{
				return errors.New("ntexabpdw4")
			}
			return nil
		}
	}
	return tlsConfig,""
}

func checkExtKeyUsageContain(keyUsage x509.ExtKeyUsage,x509Cert *x509.Certificate) bool {
	for _,ku:=range x509Cert.ExtKeyUsage{
		if ku==keyUsage{
			return true
		}
	}
	return false
}