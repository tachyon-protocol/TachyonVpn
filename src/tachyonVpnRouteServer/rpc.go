package tachyonVpnRouteServer

import (
	"github.com/tachyon-protocol/udw/udwRpc2"
	"tachyonVpnRouteServer/tachyonVpnRouteClient"
)

func Rpc_RunServer(addr string) (closer func()) {
	s := serverRpcObj{}
	sh := udwRpc2.NewServerHub(udwRpc2.ServerReq{
		Addr: addr,
		Handler: func(ctx *udwRpc2.ReqCtx) {
			var fnId uint64
			var errMsg string
			fnId, errMsg = ctx.GetReader().ReadUvarint()
			if errMsg != "" {
				return
			}
			panicErrMsg := udwRpc2.PanicToErrMsg(func() {
				switch fnId {
				case 1:
					tmp_1 := udwRpc2.PeerIp{ctx.GetPeerIp()}
					var tmp_2 tachyonVpnRouteClient.VpnNode
					errMsg = ctx.GetReader().ReadValue(&tmp_2)
					if errMsg != "" {
						return
					}
					errMsg = ctx.GetReader().ReadArrayEnd()
					if errMsg != "" {
						return
					}
					tmp_3 := s.VpnNodeRegister(tmp_1, tmp_2)
					ctx.GetWriter().WriteString("")
					errMsg = ctx.GetWriter().WriteValue(tmp_3)
					if errMsg != "" {
						return
					}
					ctx.GetWriter().WriteArrayEnd()
					errMsg = ctx.GetWriter().Flush()
					if errMsg != "" {
						return
					}
				case 2:
					errMsg = ctx.GetReader().ReadArrayEnd()
					if errMsg != "" {
						return
					}
					tmp_4 := s.VpnNodeList()
					ctx.GetWriter().WriteString("")
					errMsg = ctx.GetWriter().WriteValue(tmp_4)
					if errMsg != "" {
						return
					}
					ctx.GetWriter().WriteArrayEnd()
					errMsg = ctx.GetWriter().Flush()
					if errMsg != "" {
						return
					}
				case 3:
					errMsg = ctx.GetReader().ReadArrayEnd()
					if errMsg != "" {
						return
					}
					s.Ping()
					ctx.GetWriter().WriteString("")
					ctx.GetWriter().WriteArrayEnd()
					errMsg = ctx.GetWriter().Flush()
					if errMsg != "" {
						return
					}
				default:
				}
			})
			if panicErrMsg != "" {
				ctx.GetWriter().WriteString(panicErrMsg)
				ctx.GetWriter().WriteArrayEnd()
				errMsg = ctx.GetWriter().Flush()
				if errMsg != "" {
					return
				}
			}
		},
	})
	return sh.Close
}
