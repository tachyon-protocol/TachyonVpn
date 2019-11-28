package tachyonVpnClient

import (
	"crypto/tls"
	"github.com/tachyon-protocol/udw/udwBinary"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwLog"
	"net"
	"strconv"
	"tachyonVpnProtocol"
	"time"
	"tyTls"
	"errors"
)

type PingReq struct {
	Ip string
	ServerCertPem string // if it is "", it will use InsecureSkipVerify
	Count int
	DebugLog bool
}

//TODO relay mode
func Ping (req PingReq) error {
	var tlsConfig *tls.Config
	if req.ServerCertPem==""{
		tlsConfig = newInsecureClientTlsConifg()
	}else{
		var errMsg string
		tlsConfig,errMsg = tyTls.GetClientTlsConfigServerCertPem(req.ServerCertPem)
		if errMsg!=""{
			return errors.New(errMsg)
		}
	}
	conn, err := net.Dial("tcp", req.Ip+":"+strconv.Itoa(tachyonVpnProtocol.VpnPort))
	if err != nil {
		return err
	}
	conn = tls.Client(conn, tlsConfig)
	var (
		pingPacket = tachyonVpnProtocol.VpnPacket{
			Cmd:            tachyonVpnProtocol.CmdPing,
		}
		buf = udwBytes.NewBufWriter(nil)
	)
	for i := 0; i < req.Count; i++ {
		buf.Reset()
		pingPacket.Encode(buf)
		start := time.Now()
		err = udwBinary.WriteByteSliceWithUint32LenNoAllocV2(conn, buf.GetBytes())
		if err != nil {
			return err
		}
		if req.DebugLog {
			udwLog.Log("-> ...")
		}
		buf.Reset()
		err := udwBinary.ReadByteSliceWithUint32LenToBufW(conn, buf)
		if err != nil {
			return err
		}
		err = pingPacket.Decode(buf.GetBytes())
		if err != nil {
			return err
		}
		if req.DebugLog {
			udwLog.Log("<- ✔", time.Now().Sub(start))
		}
	}
	return nil
}
