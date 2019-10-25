package tachyonVpnClient

import (
	"crypto/tls"
	"fmt"
	"github.com/tachyon-protocol/udw/udwBinary"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwCmd"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwFile"
	"github.com/tachyon-protocol/udw/udwIpPacket"
	"github.com/tachyon-protocol/udw/udwLog"
	"github.com/tachyon-protocol/udw/udwNet"
	"github.com/tachyon-protocol/udw/udwNet/udwTapTun"
	"github.com/tachyon-protocol/udw/udwSys"
	"github.com/tachyon-protocol/udw/udwTlsSelfSignCertV2"
	"net"
	"strconv"
	"strings"
	"sync"
	"tachyonSimpleVpnProtocol"
)

type vpnClient struct {
	id          uint64
	vpnIpOffset int
	vpnIp       net.IP

	locker sync.Mutex
	conn   net.Conn
}

var (
	gLocker         sync.Mutex
	gClientMap      map[uint64]*vpnClient
	gVpnIpList      [maxCountVpnIp]*vpnClient
	gNextVpnIpIndex int
)

var (
	READONLY_vpnIpStart  = net.IP{172, 21, 0, 0}
	READONLY_vpnIpClient = net.IP{172, 21, 0, 1}
)

const maxCountVpnIp = 1 << 16

func getClient(clientId uint64, conn net.Conn) *vpnClient {
	gLocker.Lock()
	if gClientMap == nil {
		gClientMap = map[uint64]*vpnClient{}
	}
	client := gClientMap[clientId]
	if client != nil {
		gLocker.Unlock()
		return client
	}
	client = &vpnClient{
		id:          clientId,
		conn:        conn,
		locker:      sync.Mutex{},
		vpnIpOffset: 0,
	}
	lastIpOffset := gNextVpnIpIndex
	for {
		gNextVpnIpIndex = (gNextVpnIpIndex + 1) % maxCountVpnIp
		if lastIpOffset == gNextVpnIpIndex {
			gLocker.Unlock()
			panic("ip pool is full")
		}
		if gNextVpnIpIndex == 0 || gNextVpnIpIndex == 1 || gNextVpnIpIndex == 2 {
			// 172.21.0.0 ,172.21.0.1, 172.21.0.2 will not allocate to client
			continue
		}
		if gVpnIpList[gNextVpnIpIndex] == nil {
			client.vpnIpOffset = gNextVpnIpIndex
			gVpnIpList[gNextVpnIpIndex] = client
			break
		}
	}
	gClientMap[client.id] = client
	gLocker.Unlock()
	return client
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

func getClientByVpnIp(vpnIp net.IP) *vpnClient {
	offset := getVpnIpOffset(vpnIp, READONLY_vpnIpStart)
	if offset < 0 || offset >= maxCountVpnIp {
		return nil
	}
	offset = offset % 65536
	gLocker.Lock()
	client := gVpnIpList[offset]
	gLocker.Unlock()
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

func mustIptablesRestoreExist(){
	const cmd = "iptables-restore"
	if udwCmd.Exist(cmd)==false{
		udwCmd.MustRun("apt install -y iptables")
	}
	if udwCmd.Exist(cmd)==false {
		panic("7fgwy8n93j")
	}
}

func ServerRun() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(tachyonSimpleVpnProtocol.VpnPort))
	udwErr.PanicIfError(err)
	tun, err := udwTapTun.NewTun("")
	udwErr.PanicIfError(err)
	err = udwTapTun.SetP2PIpAndUp(udwTapTun.SetP2PIpRequest{
		IfaceName: tun.Name(),
		SrcIp:     udwNet.Ipv4AddAndCopyWithBuffer(READONLY_vpnIpStart, 2, nil),
		DstIp:     udwNet.Ipv4AddAndCopyWithBuffer(READONLY_vpnIpStart, 1, nil),
		Mtu:       tachyonSimpleVpnProtocol.Mtu,
		Mask:      net.CIDRMask(16, 32),
	})
	udwErr.PanicIfError(err)
	networkConfig()
	clientId := tachyonSimpleVpnProtocol.GetClientId()
	go func() {
		bufR := make([]byte, 3<<20)
		bufW := udwBytes.NewBufWriter(nil)
		for {
			n, err := tun.Read(bufR)
			udwErr.PanicIfError(err)
			packetBuf := bufR[:n]
			ipPacket, errMsg := udwIpPacket.NewIpv4PacketFromBuf(packetBuf)
			if errMsg != "" {
				//noinspection SpellCheckingInspection
				udwLog.Log("[psmddnegwg] TUN Read parse IPv4 failed", errMsg)
				return
			}
			ip := ipPacket.GetDstIp()
			client := getClientByVpnIp(ip)
			if client == nil {
				//noinspection SpellCheckingInspection
				udwLog.Log("[rdtp9rk84m] TUN Read no such client")
				continue
			}
			responseVpnPacket := &tachyonSimpleVpnProtocol.VpnPacket{
				ClientIdFrom: clientId,
				Cmd:          tachyonSimpleVpnProtocol.CmdData,
			}
			ipPacket.SetDstIp__NoRecomputeCheckSum(READONLY_vpnIpClient)
			ipPacket.TcpFixMss__NoRecomputeCheckSum(tachyonSimpleVpnProtocol.Mss)
			ipPacket.RecomputeCheckSum()
			responseVpnPacket.Data = ipPacket.SerializeToBuf()
			bufW.Reset()
			responseVpnPacket.Encode(bufW)
			_ = udwBinary.WriteByteSliceWithUint32LenNoAllocV2(client.conn, bufW.GetBytes()) //TODO
		}
	}()
	fmt.Println("Server started ✔")
	certs := []tls.Certificate{
		*udwTlsSelfSignCertV2.GetTlsCertificate(),
	}
	for {
		conn, err := ln.Accept()
		udwErr.PanicIfError(err)
		if tachyonSimpleVpnProtocol.Debug {
			udwLog.Log("New Conn", conn.RemoteAddr())
		}
		conn = tls.Server(conn, &tls.Config{
			Certificates: certs,
			NextProtos:   []string{"http/1.1"},
		})
		go func() {
			bufR := make([]byte, 3<<20)
			vpnPacket := &tachyonSimpleVpnProtocol.VpnPacket{}
			vpnIpBufW := udwBytes.NewBufWriter(nil)
			for {
				out, err := udwBinary.ReadByteSliceWithUint32LenNoAllocLimitMaxSize(conn, bufR, uint32(len(bufR)))
				if err != nil {
					_ = conn.Close()
					return
				}
				err = vpnPacket.Decode(out)
				if err != nil {
					_ = conn.Close()
					return
				}
				client := getClient(vpnPacket.ClientIdFrom, conn)
				ipPacket, errMsg := udwIpPacket.NewIpv4PacketFromBuf(vpnPacket.Data)
				if errMsg != "" {
					_ = conn.Close()
					return
				}
				vpnIp := udwNet.Ipv4AddAndCopyWithBuffer(READONLY_vpnIpStart, uint32(client.vpnIpOffset), vpnIpBufW)
				ipPacket.SetSrcIp__NoRecomputeCheckSum(vpnIp)
				ipPacket.TcpFixMss__NoRecomputeCheckSum(tachyonSimpleVpnProtocol.Mss)
				ipPacket.RecomputeCheckSum()
				_, err = tun.Write(ipPacket.SerializeToBuf())
				if err != nil {
					_ = conn.Close()
					return
				}
			}
		}()
	}
}
