//+build linux

package tachyonVpnClient

import (
	"bytes"
	"github.com/tachyon-protocol/udw/udwCmd"
	"github.com/tachyon-protocol/udw/udwFile"
	"github.com/tachyon-protocol/udw/udwNet"
	"github.com/tachyon-protocol/udw/udwTime"
)

func configLocalNetwork() {
	route:= udwNet.MustGetDefaultRouteRule()
	if !bytes.Contains(udwFile.MustReadFile("/etc/iproute2/rt_tables"), []byte("\n200 isp2")) {
		udwFile.MustAppendFile("/etc/iproute2/rt_tables", []byte("\n200 isp2"))
	}
	gatewayInterfaceIpString := route.GetOutInterface().GetFirstIpv4IP().String()
	if !bytes.Contains(udwCmd.MustCombinedOutput("ip rule"), []byte("from "+gatewayInterfaceIpString+" lookup isp2")) {
		udwCmd.MustRun("ip rule add from " + gatewayInterfaceIpString + " table isp2")
	}
	b,err:=udwCmd.CmdString("ip route add default via " + route.GetGatewayIp().String() + " dev " + route.GetOutInterface().GetName() + " table isp2").
		CombinedOutputAndNotExitStatusCheck()
	if err!=nil{
		panic("qfqr388vtr "+err.Error()+" "+string(b))
	}
	udwFile.MustCopyFile("/etc/resolv.conf", "/etc/backup/resolv.conf")
	udwFile.MustCopyFile("/etc/resolv.conf", "/etc/backup/resolv.conf."+udwTime.NowWithFileNameFormatV2())

	udwFile.MustWriteFile("/etc/resolv.conf", []byte(`# replace by vpn
nameserver 8.8.8.8
nameserver 8.8.4.4
`))
}

func recoverLocalNetwork() {
	udwFile.MustCopyFile("/etc/resolv.conf", "/etc/backup/resolv.conf."+udwTime.NowWithFileNameFormatV2())
	if udwFile.MustFileExist("/etc/backup/resolv.conf") {
		udwFile.MustCopyFile("/etc/backup/resolv.conf", "/etc/resolv.conf")
	}
}
