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
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println(`CheckHelper usage:
CheckHelper [ServerId]
`)
		os.Exit(-1)
		return
	}
	fmt.Println("CheckHelper start server, try to listen to 443 port.")
	serverId := strings.TrimSpace(os.Args[1])
	l, err := net.Listen("tcp", ":443")
	if err != nil {
		fmt.Println("net.Listen fail " + err.Error())
		os.Exit(-1)
		return
	}
	go func() {
		err := http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte(serverId))
		}))
		if err != nil {
			fmt.Println("http.Serve fail " + err.Error())
			os.Exit(-1)
			return
		}
	}()
	time.Sleep(time.Millisecond)
	for i := 0; ; i++ {
		if i == 9 {
			fmt.Println("CheckHelper start server fail time out")
			l.Close()
			os.Exit(-1)
			return
		}
		resp, err := http.Get("http://127.0.0.1:443")
		if err != nil || resp.StatusCode != 200 {
			time.Sleep(time.Second)
			continue
		}
		body := make([]byte, len(serverId)+4096)
		nr, err := io.ReadAtLeast(resp.Body, body, len(serverId))
		if err != nil || nr != len(serverId) || bytes.Equal(body[:nr], []byte(serverId)) == false {
			time.Sleep(time.Second)
			continue
		}
		break
	}
	fmt.Println("CheckHelper ✔")
	fmt.Println("use Ctrl+C or kill " + strconv.Itoa(os.Getpid()) + " to close it.")
	waitForExit()
	l.Close()
}

func waitForExit() {
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-ch
}
