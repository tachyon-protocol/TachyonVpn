package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"sync"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println(`CheckHelper2 usage:
CheckHelper2 [domain]
`)
		os.Exit(-1)
		return
	}
	fmt.Println("start get token,please wait...")
	domain:=os.Args[1]
	thisUrl:="https://"+domain+"/?n=yr8mtzfwee.GetTakenFromNodeSelfIp"
	b,ok:=getUrlContent(thisUrl,30*time.Second)
	if !ok{
		fmt.Println("can not get token from "+domain)
		os.Exit(-1)
		return
	}
	respS:=string(b)
	const resptP = `token_"`
	const resptS = `"`+"\n"
	if strings.HasPrefix(respS,resptP) ==false || strings.HasSuffix(respS,resptS)==false{
		fmt.Println("can not get token2 from "+domain)
		os.Exit(-1)
		return
	}
	token:=strings.TrimSuffix(strings.TrimPrefix(respS,resptP),resptS)
	fmt.Println("get token finish, start server... token:",token)
	l, err := net.Listen("tcp", ":443")
	if err != nil {
		fmt.Println("net.Listen fail " + err.Error())
		os.Exit(-1)
		return
	}
	gLocker :=sync.Mutex{}
	gIsCloser:=false
	wg:=sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		err := http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte(token))
		}))
		gLocker.Lock()
		isCloser:=gIsCloser
		gLocker.Unlock()
		if isCloser{
			return
		}
		if err != nil {
			fmt.Println("http.Serve fail " + err.Error())
			os.Exit(-1)
			return
		}
	}()
	wg.Wait()
	time.Sleep(time.Millisecond)
	for i := 0; ; i++ {
		if i == 9 {
			fmt.Println("CheckHelper start server fail time out")
			l.Close()
			os.Exit(-1)
			return
		}
		resp,ok:=getUrlContent("http://127.0.0.1:443",time.Second)
		if ok==false{
			time.Sleep(time.Second)
			continue
		}
		if bytes.Equal(resp, []byte(token)) == false {
			time.Sleep(time.Second)
			continue
		}
		break
	}
	fmt.Println("start web server finish, ask remote to verify...")
	resp,ok:=getUrlContent("https://"+domain+"/?n=yr8mtzfwee.VerifyTokenFromNodeSelfIp",30*time.Second)
	if !ok || bytes.Equal(resp,[]byte("success"))==false{
		fmt.Println("verify token fail from "+domain)
		os.Exit(-1)
		return
	}

	fmt.Println("CheckHelper2 ✔")
	fmt.Println("use Ctrl+C or kill " + strconv.Itoa(os.Getpid()) + " to close it.")
	waitForExit()
	gLocker.Lock()
	gIsCloser = true
	gLocker.Unlock()
	l.Close()
}

func waitForExit() {
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-ch
}

func getUrlContent(url string,timeout time.Duration) (b []byte,ok bool){
	client:=&http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != 200 || resp.Body==nil{
		return nil,false
	}
	defer resp.Body.Close()
	body := make([]byte, 4096)
	nr, err := io.ReadFull(resp.Body, body)
	if err==nil || err==io.ErrUnexpectedEOF{
		return body[:nr],true
	}
	return nil,false
}