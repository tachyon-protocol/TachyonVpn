package tachyonVpnInfoClient

type RegisterAsVpnNodeReq struct{
	Ip string `json:",omitempty"`
	ServerCertPem string `json:",omitempty"`
}