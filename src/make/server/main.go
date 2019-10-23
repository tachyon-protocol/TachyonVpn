package main

import (
	"fmt"
	"github.com/tachyon-protocol/udw/udwBinary"
	"github.com/tachyon-protocol/udw/udwBytes"
	"github.com/tachyon-protocol/udw/udwErr"
	"github.com/tachyon-protocol/udw/udwIpPacket"
	"github.com/tachyon-protocol/udw/udwLog"
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


func getClientByVpnIp(ip net.IP) (client *vpnClient){
	offset :=
	if offset < 0 || offset >= maxNatIpNumber {
		if debugError {
			kmgLog.Log("error", "[kmgVpnServer.tunReadThread] dst ip not in ip list range", ip.String())
		}
		return nil
	}
	offset = offset % 65536

	client = server.vpnIPList[offset]
	if client == nil {
		if debugError{
			server.lastNotFoundClientIpOffsetLocker.Lock()
			if server.lastNotFoundClientIpOffset != offset {
				// 这个是为了避免重复消息数量过多?
				if debugError {
					kmgLog.Log("error", "[kmgVpnServer.tunReadThread] client not exist", ip.String())
				}
			}
			server.lastNotFoundClientIpOffset = offset
			server.lastNotFoundClientIpOffsetLocker.Unlock()
		}
		return nil
	}
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
			packetBuf := bufR[:n]
			ipPacket, errMsg := udwIpPacket.NewIpv4PacketFromBuf(packetBuf)
			if errMsg != "" {
				udwLog.Log("[psmddnegwg] TUN Read parse IPv4 failed", errMsg)
				return
			}
			ip := ipPacket.GetDstIp()
			gLocker.RLock()
			client :=getClientByNatIp__NOLOCK(ip)
			if client==nil{
				server.locker.RUnlock()
				return
			}
			responsePacket := &kmgVpnV2.VpnPacket{
				ClientId:       client.clientId,
				Cmd:            kmgVpnV2.CmdData,
				DataSequenceId: client.lastPacketId,
			}
			server.locker.RUnlock()
			ipPacket.SetDstIp__NoRecomputeCheckSum(kmgVpnV2.ClientIp)
			ipPacket.TcpFixMss__NoRecomputeCheckSum(kmgVpnV2.TcpMss)
			ipPacket.RecomputeCheckSum() // 修改了客户端ip，此处必须重新计算checksum（如果能bc，可以丢给客户端做）
			//if server.req.PacketRecordCallback != nil {
			//server.req.PacketRecordCallback(PacketRecordMessage{
			//	UserId:       client.userId,
			//	IpPacket:     ipPacket,
			//	IsSendToUser: true,
			//})
			//}
			responsePacket.Data = ipPacket.SerializeToBuf()
			if client.clientType == clientInServerTypeExit {
				selfIpByte := []byte(net.ParseIP(server.req.SelfIp).To4())
				if len(selfIpByte) == 0 {
					return
				}
				responsePacket.EntranceIp = selfIpByte
				responsePacket.ExitIp = client.entranceIp
				responsePacket.RemoteVersion = kmgVpnV2.Version
				buf, err := kmgVpnV2.WriteVpnPacketToBytes(responsePacket)
				if err != nil {
					if debugError {
						kmgLog.Log("error", "[error 70]", err)
					}
					return
				}
				server.forwardToNextHop(net.IP(responsePacket.ExitIp).To4().String(), buf)
				return
			}
			//if accountDebug {
			//	kmgLog.Log("debug", "[accountDebug3] client.clientType==", client.clientType, len(responsePacket.Data), server.req.SelfIp, client.entranceIp)
			//}
			hasDropPacket := server.addDataAccountMessage(client, int64(len(responsePacket.Data)+kmgVpnV2.DataPacketOverHead))
			if hasDropPacket {
				return
			}
			errSlice := server.writeVpnPacketToClient(client, responsePacket)
			if debugError {
				for _, err := range errSlice {
					kmgLog.Log("error", "tunReadThread", err.Error())
				}
			}
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
