package tachyonVpnRouteClient

import "time"

//type RegisterAsVpnNodeReq struct{
//	Ip string `json:",omitempty"`
//	ServerChk string `json:",omitempty"`
//}

type VpnNode struct{
	Ip string `json:",omitempty"`
	ServerChk string `json:",omitempty"`
	UpdateTime time.Time
}