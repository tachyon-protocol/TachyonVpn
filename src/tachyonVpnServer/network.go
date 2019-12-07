package tachyonVpnServer

import (
	"errors"
	"github.com/tachyon-protocol/udw/udwCmd"
	"github.com/tachyon-protocol/udw/udwFile"
	"github.com/tachyon-protocol/udw/udwSys"
	"net"
	"strings"
	"sync"
)

var (
	READONLY_vpnIpStart  = net.IP{172, 21, 0, 0}
	READONLY_vpnIpClient = net.IP{172, 21, 0, 1}
)

const maxCountVpnIp = 1 << 16

func (s *Server) clientAllocateVpnIp_NoLock(client *vpnClient) error {
	lastIpOffset := s.nextVpnIpIndex
	for {
		s.nextVpnIpIndex = (s.nextVpnIpIndex + 1) % maxCountVpnIp
		if lastIpOffset == s.nextVpnIpIndex {
			return errors.New("[993yzr1tbz] ip pool is full")
		}
		if s.nextVpnIpIndex == 0 || s.nextVpnIpIndex == 1 || s.nextVpnIpIndex == 2 {
			// 172.21.0.0 ,172.21.0.1, 172.21.0.2 will not allocate to client
			continue
		}
		if s.vpnIpList[s.nextVpnIpIndex] == nil {
			client.vpnIpOffset = s.nextVpnIpIndex
			s.vpnIpList[s.nextVpnIpIndex] = client
			return nil
		}
	}
}

func getVpnIpOffset(ip1 net.IP, ip2 net.IP) int {
	ipv41 := ip1.To4()
	ipv42 := ip2.To4()
	if ipv41 == nil {
		panic("[ipSub] ip1 is not ipv4 addr")
	}
	if ipv42 == nil {
		panic("[ipSub] ip2 is not ipv4 addr")
	}
	out := 0
	base := 1
	for i := 3; i >= 0; i-- {
		out = out + int(ipv41[i]-ipv42[i])*base
		base = base * 256
	}
	return out
}

func (s *Server) getClientByVpnIp(vpnIp net.IP) *vpnClient {
	offset := getVpnIpOffset(vpnIp, READONLY_vpnIpStart)
	if offset < 0 || offset >= maxCountVpnIp {
		return nil
	}
	offset = offset % 65536
	s.lock.Lock()
	client := s.vpnIpList[offset]
	s.lock.Unlock()
	if client == nil {
		return nil
	}
	return client
}

var (
	networkConfigOnce                  = &sync.Once{}
	networkConfigIptablesConfigContent = []byte(`*filter
COMMIT
*mangle
-A PREROUTING -s 172.20.0.0/16 -p tcp -j TPROXY --on-port 23498 --on-ip 127.0.0.1 --tproxy-mark 0x1/0x1
COMMIT
*nat
-A POSTROUTING -s 172.20.0.0/16 -p udp -j MASQUERADE
-A POSTROUTING -s 172.21.0.0/16 -j MASQUERADE
COMMIT
`)
)

func networkConfig() {
	networkConfigOnce.Do(func() {
		mustIptablesRestoreExist()
		udwSys.SetIpForwardOn()
		const iptablesConfigFile = `/tmp/iptables.config`
		udwFile.MustWriteFile(iptablesConfigFile, networkConfigIptablesConfigContent)
		udwCmd.MustRun("iptables-restore " + iptablesConfigFile)
		b := udwCmd.MustRunAndReturnOutput("ip rule")
		if !strings.Contains(string(b), "fwmark 0x1 lookup 100") {
			udwCmd.MustRun("ip rule add fwmark 1 lookup 100")
		}
		_ = udwCmd.Run("ip route add local 0.0.0.0/0 dev lo table 100")
	})
}

func mustIptablesRestoreExist() {
	const cmd = "iptables-restore"
	if udwCmd.Exist(cmd) == false {
		udwCmd.MustRun("apt install -y iptables")
	}
	if udwCmd.Exist(cmd) == false {
		panic("7fgwy8n93j")
	}
}
