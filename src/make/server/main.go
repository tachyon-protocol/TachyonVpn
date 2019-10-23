package main

import (
	"github.com/tachyon-protocol/udw/udwBinary"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwIpPacket"
	"github.com/tachyon-protocol/udw/udwNet"
	"github.com/tachyon-protocol/udw/udwNet/udwTapTun"
	"net"
	"sync"
	"tachyonSimpleVpnPacket"
)

type vpnClient struct {
	id          uint64
	vpnIpOffset int

	locker sync.Mutex
	conn   net.Conn
}

var (
	gLocker         sync.Mutex
	gClientMap      map[uint64]*vpnClient
	gVpnIpList      [maxCountVpnIp]*vpnClient
	gNextVpnIpIndex int
)

const maxCountVpnIp = 1 << 16

func getClient(clientId uint64, conn net.Conn) *vpnClient {
	gLocker.Lock()
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

func main() {
	startVpnIp := net.IP{172, 21, 0, 0}
	ln, err := net.Listen("tcp", ":29443")
	udwErr.PanicIfError(err)
	tun, err := udwTapTun.NewTun("")
	udwErr.PanicIfError(err)
	err = udwTapTun.SetP2PIpAndUp(udwTapTun.SetP2PIpRequest{
		IfaceName: tun.Name(),
		SrcIp:     udwNet.Ipv4AddAndCopyWithBuffer(startVpnIp, 2, nil),
		DstIp:     udwNet.Ipv4AddAndCopyWithBuffer(startVpnIp, 1, nil),
		Mtu:       tachyonSimpleVpnPacket.Mtu,
		Mask:      net.CIDRMask(16, 32),
	})
	udwErr.PanicIfError(err)
	go func() {
		bufR := make([]byte, 3<<20)
		for {
			n, err := tun.Read(bufR)
			udwErr.PanicIfError(err)

		}
	}()
	for {
		conn, err := ln.Accept()
		udwErr.PanicIfError(err)
		go func() {
			bufR := make([]byte, 3<<20)
			vpnPacket := &tachyonSimpleVpnPacket.VpnPacket{}
			vpnIpBufW := udwBytes.NewBufWriter(nil)
			for {
				out, err := udwBinary.ReadByteSliceWithUint32LenNoAllocLimitMaxSize(conn, bufR, uint32(len(bufR)))
				udwErr.PanicIfError(err)
				err = vpnPacket.Decode(out)
				udwErr.PanicIfError(err)
				client := getClient(vpnPacket.ClientIdFrom, conn)
				ipPacket, errMsg := udwIpPacket.NewIpv4PacketFromBuf(vpnPacket.Data)
				if errMsg != "" {
					panic("parse IPv4 failed:" + errMsg)
				}
				vpnIp := udwNet.Ipv4AddAndCopyWithBuffer(startVpnIp, uint32(client.vpnIpOffset), vpnIpBufW)
				ipPacket.SetSrcIp__NoRecomputeCheckSum(vpnIp)
				ipPacket.TcpFixMss__NoRecomputeCheckSum(tachyonSimpleVpnPacket.Mss)
				ipPacket.RecomputeCheckSum()
				tun.WriteWithBuffer(ipPacket.SerializeToBuf())
			}
		}()
	}
}
