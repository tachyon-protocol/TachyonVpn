package tachyonVpnInfoClient

import (
	"testing"
	"tachyonVpnInfoServer"
	"github.com/tachyon-protocol/udw/udwTest"
)

func TestCs(ot *testing.T){
	sCloser:=tachyonVpnInfoServer.ServerAsyncRun()
	defer sCloser()
	c:=NewClient("127.0.0.1")
	errMsg:=c.RegisterFromIpAsVpnNode()
	udwTest.Equal(errMsg,"")
	ipList,errMsg:=c.GetVpnNodeIpList()
	udwTest.Equal(errMsg,"")
	udwTest.Equal(ipList,[]string{"127.0.0.1"})
	errMsg=c.UnregisterFromIpAsVpnNode()
	udwTest.Equal(errMsg,"")
	ipList,errMsg=c.GetVpnNodeIpList()
	udwTest.Equal(errMsg,"")
	udwTest.Equal(len(ipList),0)
}