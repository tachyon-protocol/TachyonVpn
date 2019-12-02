package tyTls

import (
	"encoding/base64"
	"github.com/tachyon-protocol/udw/udwCryptoSha3"
	"crypto/tls"
)

func HashChk(b []byte) string{
	hcF:=udwCryptoSha3.Sum512(b)
	hc:=hcF[:32]
	checkSumByte:=udwCryptoSha3.Sum512(hc)[0]
	return base64.RawURLEncoding.EncodeToString(append(hc,checkSumByte))
}

func MustHashChkFromTlsCert(tlsCert *tls.Certificate) string{
	if tlsCert==nil || len(tlsCert.Certificate)==0{
		panic("r3af7nn6fp")
	}
	return HashChk(tlsCert.Certificate[0])
}

func IsChkValid(chk string) bool{
	if len(chk)!=44{
		return false
	}
	chkB,err:=base64.RawURLEncoding.DecodeString(chk)
	if err!=nil{
		return false
	}
	if len(chkB)!=33{
		return false
	}
	checkSumByte:=udwCryptoSha3.Sum512(chkB[:32])[0]
	if checkSumByte!=chkB[32]{
		return false
	}
	return true
}