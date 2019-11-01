package tachyonVpnInfoClient

import (
	"net/http"
	"crypto/tls"
	"io/ioutil"
	"encoding/json"
)

func NewClient(ip string) *Client{
	return &Client{ip+":443"}
}

type Client struct{
	remoteAddr string
}

func (c *Client) RegisterFromIpAsVpnNode() (errMsg string){
	ctx:=&httpClientCtx{
		Url: "https://"+c.remoteAddr+"/?n=RegisterFromIpAsVpnNode",
		Method: "POST",
	}
	ctx.Send()
	return ctx.ErrMsg
}
func (c *Client) UnregisterFromIpAsVpnNode()(errMsg string){
	ctx:=&httpClientCtx{
		Url: "https://"+c.remoteAddr+"/?n=UnregisterFromIpAsVpnNode",
		Method: "POST",
	}
	ctx.Send()
	return ctx.ErrMsg
}
func (c *Client) GetVpnNodeIpList() (ipList []string,errMsg string){
	ctx:=&httpClientCtx{
		Url: "https://"+c.remoteAddr+"/?n=GetVpnNodeIpList",
		Method: "POST",
	}
	ctx.Send()
	if ctx.ErrMsg!=""{
		return nil,ctx.ErrMsg
	}
	err := json.Unmarshal(ctx.RespBody,&ipList)
	if err!=nil{
		return nil,err.Error()
	}
	return ipList,""
}

type httpClientCtx struct{
	Url string
	Method string
	//IsSkipTls bool

	ErrMsg string
	RespStatus int
	RespBody []byte
}

func (ctx *httpClientCtx) Send(){
	c:=&http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	req,err:=http.NewRequest(ctx.Method,ctx.Url,nil)
	if err!=nil{
		ctx.ErrMsg = err.Error()
		return
	}
	resp,err:=c.Do(req)
	if err!=nil{
		ctx.ErrMsg = err.Error()
		return
	}
	ctx.RespStatus = resp.StatusCode
	ctx.RespBody,err = ioutil.ReadAll(resp.Body)
	if err!=nil{
		ctx.ErrMsg = err.Error()
		return
	}
	return
}