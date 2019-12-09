package tachyonVpnRoute_Build

import (
	"github.com/tachyon-protocol/udw/udwRpc2/udwRpc2Builder"
)

func UdwBuild(){
	udwRpc2Builder.Generate(udwRpc2Builder.GenerateReq{
		RpcDefine:      getRpcService(),
		FromObjName:    "serverRpcObj",
		FromPkgPath:    "tachyonVpnRouteServer",
		TargetPkgPath:  "tachyonVpnRouteServer",
		Prefix:         "Rpc",
		TargetFilePath: "src/tachyonVpnRouteServer/rpc.go",
		GoFmt:          true,
		DisableGenClient: true,
	})
	udwRpc2Builder.Generate(udwRpc2Builder.GenerateReq{
		RpcDefine:      getRpcService(),
		FromObjName:    "serverRpcObj",
		FromPkgPath:    "tachyonVpnRouteServer",
		TargetPkgPath:  "tachyonVpnRouteServer/tachyonVpnRouteClient",
		Prefix:         "Rpc",
		TargetFilePath: "src/tachyonVpnRouteServer/tachyonVpnRouteClient/rpc.go",
		GoFmt:          true,
		DisableGenServer: true,
	})
}

func getRpcService() udwRpc2Builder.RpcService {
	return udwRpc2Builder.RpcService{
		List: []udwRpc2Builder.RpcApi{
			{
				Name: "VpnNodeRegister",
				InputParameterList: []udwRpc2Builder.RpcParameter{
					{
						Type: udwRpc2Builder.RpcType{
							Kind:       udwRpc2Builder.RpcTypeKindNamedStruct,
							StructName: "PeerIp",
							GoPkg:      "github.com/tachyon-protocol/udw/udwRpc2",
						},
					},
					{
						Type: udwRpc2Builder.RpcType{
							Kind:       udwRpc2Builder.RpcTypeKindNamedStruct,
							StructName: "VpnNode",
							GoPkg:      "tachyonVpnRouteServer/tachyonVpnRouteClient",
						},
					},
				},
				OutputParameterList: []udwRpc2Builder.RpcParameter{
					{
						Type: udwRpc2Builder.RpcType{
							Kind: udwRpc2Builder.RpcTypeKindString,
						},
					},
				},
			},
			{
				Name: "VpnNodeList",
				OutputParameterList: []udwRpc2Builder.RpcParameter{
					{
						Type: udwRpc2Builder.RpcType{
							Kind: udwRpc2Builder.RpcTypeKindSlice,
							Elem: &udwRpc2Builder.RpcType{
								Kind:       udwRpc2Builder.RpcTypeKindNamedStruct,
								StructName: "VpnNode",
								GoPkg:      "tachyonVpnRouteServer/tachyonVpnRouteClient",
							},
						},
					},
				},
			},
		},
	}
}