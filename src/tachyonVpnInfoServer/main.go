package tachyonVpnInfoServer

import (
	"github.com/tachyon-protocol/udw/udwSqlite3"
	"sync"
	"net/http"
	"github.com/tachyon-protocol/udw/udwTlsSelfSignCertV2"
	"fmt"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwJson"
	"net"
	"time"
	"github.com/tachyon-protocol/udw/udwTime"
)

func ServerAsyncRun() func(){
	initDb()
	s:=http.Server{
		Addr: ":443",
		Handler: http.HandlerFunc(serverHandler),
		TLSConfig: udwTlsSelfSignCertV2.GetTlsConfig(),
	}
	wg:=sync.WaitGroup{}
	wg.Add(1)
	go func(){
		err := s.ListenAndServeTLS("","")
		wg.Done()
		if err!=nil && err!=http.ErrServerClosed{
			fmt.Println("hguuwustns",err)
			return
		}
	}()
	return func(){
		s.Close()
		wg.Wait()
	}
}

func serverHandler(w http.ResponseWriter,req *http.Request){
	errMsg:=udwErr.PanicToErrorMsg(func(){
		if req.Method!=http.MethodPost{
			http.NotFound(w,req)
			return
		}
		values:=req.URL.Query()
		n:=values.Get("n")
		switch n {
		case "RegisterFromIpAsVpnNode":
			fromIp:=getClientIpStringIgnoreError(req)
			serverRpcObj{}.RegisterFromIpAsVpnNode(fromIp)
		case "UnregisterFromIpAsVpnNode":
			fromIp:=getClientIpStringIgnoreError(req)
			serverRpcObj{}.UnregisterFromIpAsVpnNode(fromIp)
		case "GetVpnNodeIpList":
			ipList:=serverRpcObj{}.GetVpnNodeIpList()
			w.Write(udwJson.MustMarshal(ipList))
		default:
			http.NotFound(w,req)
			return
		}
	})
	if errMsg!=""{
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
	}
	return
}

type serverRpcObj struct{}
func (serverRpcObj) RegisterFromIpAsVpnNode(fromIp string){
	startTime:=time.Now()
	getDb().MustSet(k1VpnNodeIp,fromIp,udwTime.MustDbTimeGetStringFromObj(startTime))
}
func (serverRpcObj) UnregisterFromIpAsVpnNode(fromIp string){
	getDb().MustDelete(k1VpnNodeIp,fromIp)
}
func (serverRpcObj) GetVpnNodeIpList() []string{
	outputList:=[]string{}
	startTime:=time.Now()
	getDb().MustGetRangeCallback(udwSqlite3.GetRangeReq{
		K1: k1VpnNodeIp,
		Limit: 1000,
	},func(k string,v string){
		t:=udwTime.MustDbTimeGetObjFromString(v)
		if t.Before(startTime.Add(-time.Second*30)){
			getDb().MustDelete(k1VpnNodeIp,k)
		}else{
			outputList = append(outputList,k)
		}
	})
	return outputList
}

var gSqlite3Db *udwSqlite3.Db
var gSqlite3DbOnce sync.Once

func initDb(){
	gSqlite3DbOnce.Do(func(){
		gSqlite3Db = udwSqlite3.MustNewDb(udwSqlite3.NewDbRequest{
			FilePath: "/usr/local/var/tachyonVpnInfoServer.sqlite3",
			EmptyDatabaseIfDatabaseCorrupt: true,
		})
	})
}

func getDb() *udwSqlite3.Db{
	return gSqlite3Db
}

func getClientIpStringIgnoreError(req *http.Request) string{
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil {
		return host
	}
	return ""
}

const k1VpnNodeIp = "k1VpnNodeIp"