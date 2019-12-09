package tachyonVpnRouteClient

import (
	"github.com/tachyon-protocol/udw/udwRpc2"
)

func Rpc_NewClient(addr string) *Rpc_Client {
	c := udwRpc2.NewClientHub(udwRpc2.ClientReq{
		Addr: addr,
	})
	return &Rpc_Client{
		ch: c,
	}
}

type Rpc_Client struct {
	ch *udwRpc2.ClientHub
}

func (c *Rpc_Client) VpnNodeRegister(fi2 VpnNode) (fo1 string, RpcErr *udwRpc2.RpcError) {
	_networkErr := c.ch.RequestCb(func(ctx *udwRpc2.ReqCtx) {
		ctx.GetWriter().WriteUvarint(1)
		ctx.GetWriter().WriteValue(fi2)
		ctx.GetWriter().WriteArrayEnd()
		errMsg := ctx.GetWriter().Flush()
		if errMsg != "" {
			RpcErr = udwRpc2.NewNetworkError("dehqx82rjj " + errMsg)
			return
		}
		var s string
		errMsg = ctx.GetReader().ReadValue(&s)
		if errMsg != "" {
			RpcErr = udwRpc2.NewNetworkError("ehtjkea4re " + errMsg)
			return
		}
		if s != "" {
			RpcErr = udwRpc2.NewOtherError(s)
			ctx.GetReader().ReadArrayEnd()
			return
		}
		errMsg = ctx.GetReader().ReadValue(&fo1)
		if errMsg != "" {
			RpcErr = udwRpc2.NewNetworkError("kvkdcgtnk2 " + errMsg)
			return
		}
		errMsg = ctx.GetReader().ReadArrayEnd()
		if errMsg != "" {
			RpcErr = udwRpc2.NewNetworkError("4b7rug5mf2 " + errMsg)
			return
		}
		RpcErr = nil
		return
	})
	if _networkErr != "" {
		RpcErr = udwRpc2.NewNetworkError("494fehebw6 " + _networkErr)
	}
	return
}
func (c *Rpc_Client) VpnNodeList() (fo1 []VpnNode, RpcErr *udwRpc2.RpcError) {
	_networkErr := c.ch.RequestCb(func(ctx *udwRpc2.ReqCtx) {
		ctx.GetWriter().WriteUvarint(2)
		ctx.GetWriter().WriteArrayEnd()
		errMsg := ctx.GetWriter().Flush()
		if errMsg != "" {
			RpcErr = udwRpc2.NewNetworkError("dehqx82rjj " + errMsg)
			return
		}
		var s string
		errMsg = ctx.GetReader().ReadValue(&s)
		if errMsg != "" {
			RpcErr = udwRpc2.NewNetworkError("ehtjkea4re " + errMsg)
			return
		}
		if s != "" {
			RpcErr = udwRpc2.NewOtherError(s)
			ctx.GetReader().ReadArrayEnd()
			return
		}
		errMsg = ctx.GetReader().ReadValue(&fo1)
		if errMsg != "" {
			RpcErr = udwRpc2.NewNetworkError("kvkdcgtnk2 " + errMsg)
			return
		}
		errMsg = ctx.GetReader().ReadArrayEnd()
		if errMsg != "" {
			RpcErr = udwRpc2.NewNetworkError("4b7rug5mf2 " + errMsg)
			return
		}
		RpcErr = nil
		return
	})
	if _networkErr != "" {
		RpcErr = udwRpc2.NewNetworkError("494fehebw6 " + _networkErr)
	}
	return
}
func (c *Rpc_Client) Ping() (RpcErr *udwRpc2.RpcError) {
	_networkErr := c.ch.RequestCb(func(ctx *udwRpc2.ReqCtx) {
		ctx.GetWriter().WriteUvarint(3)
		ctx.GetWriter().WriteArrayEnd()
		errMsg := ctx.GetWriter().Flush()
		if errMsg != "" {
			RpcErr = udwRpc2.NewNetworkError("dehqx82rjj " + errMsg)
			return
		}
		var s string
		errMsg = ctx.GetReader().ReadValue(&s)
		if errMsg != "" {
			RpcErr = udwRpc2.NewNetworkError("ehtjkea4re " + errMsg)
			return
		}
		if s != "" {
			RpcErr = udwRpc2.NewOtherError(s)
			ctx.GetReader().ReadArrayEnd()
			return
		}
		errMsg = ctx.GetReader().ReadArrayEnd()
		if errMsg != "" {
			RpcErr = udwRpc2.NewNetworkError("4b7rug5mf2 " + errMsg)
			return
		}
		RpcErr = nil
		return
	})
	if _networkErr != "" {
		RpcErr = udwRpc2.NewNetworkError("494fehebw6 " + _networkErr)
	}
	return
}
