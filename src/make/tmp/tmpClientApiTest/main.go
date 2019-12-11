package main

import (
	"github.com/tachyon-protocol/udw/tyVpnClient"
	"github.com/tachyon-protocol/udw/udwLog"
	"time"
	"fmt"
)

func main(){
	tyVpnClient.SetOnChangeCallbackFilterSame("cmd",func(vpnStatus string,lastError string){
		udwLog.Log("SetOnChangeCallback","["+vpnStatus+"]","["+lastError+"]")
	})
	for i:=0;i<10;i++{
		startTime:=time.Now()
		tyVpnClient.Reconnect()
		for {
			status:=tyVpnClient.GetVpnStatus()
			if status!=tyVpnClient.Connecting{
				break
			}
			time.Sleep(time.Millisecond*10)
		}
		tyVpnClient.Disconnect()
		fmt.Println("finish waiting",time.Since(startTime))
	}
	time.Sleep(time.Second*10)
}
