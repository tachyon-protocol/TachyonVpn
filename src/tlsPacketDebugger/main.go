package tlsPacketDebugger

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

type record struct {
	ContentType  string
	Length       int
	ProtocolList []handshakeProtocol
}

type handshakeProtocol struct {
	HandshakeType string
	Length        int
}

func Dump(logPrefix string, packet []byte) {
	records := GetRecordList(packet)
	fmt.Println(logPrefix, "write", len(packet))
	for _, r := range records {
		fmt.Println("	", r.ContentType, r.Length)
		for _, p := range r.ProtocolList {
			fmt.Println("		", p.HandshakeType, p.Length)
		}
	}
}

func GetRecordList(packet []byte) (list []record) {
	i := 0
	for {
		r := record{
			Length: int(binary.BigEndian.Uint16(packet[i+3 : i+5])),
		}
		switch packet[i] {
		case 0x16:
			r.ContentType = "Handshake"
			j := i + 5
			if packet[j] == 0 && packet[j+1] == 0 && packet[j+2] == 0 && packet[j+3] == 0 {
				r.ProtocolList = append(r.ProtocolList, handshakeProtocol{
					HandshakeType: "EncryptedHandshakeMessage",
				})
				break
			}
			lenBuf := make([]byte, 4)
			for {
				lenBuf[0] = 0
				lenBuf[1] = 0
				lenBuf[2] = 0
				lenBuf[3] = 0
				copy(lenBuf[1:], packet[j+1:j+1+3])
				p := handshakeProtocol{
					Length: int(binary.BigEndian.Uint32(lenBuf)),
				}
				switch packet[j] {
				case 0x01:
					p.HandshakeType = "ClientHello"
				case 0x10:
					p.HandshakeType = "ClientKeyExchange"
				case 0x02:
					p.HandshakeType = "ServerHello"
				case 0x0b:
					p.HandshakeType = "Certificate"
				case 0x0c:
					p.HandshakeType = "ServerKeyExchange"
				case 0x0e:
					p.HandshakeType = "ServerHelloDone"
				default:
					//panic("unknown protocol type:" + hex.EncodeToString([]byte{packet[j]}))
					p.HandshakeType = "Unknown " + hex.EncodeToString([]byte{packet[j]})
				}
				r.ProtocolList = append(r.ProtocolList, p)
				j = j + 4 + p.Length
				if j == i+r.Length+5 {
					break
				}
			}
		case 0x14:
			r.ContentType = "ChangeCipherSpec"
		case 0x17:
			r.ContentType = "ApplicationData"
		default:
			//panic("unknown content type:" + hex.EncodeToString([]byte{packet[i]}))
			r.ContentType = "unknown content type:" + hex.EncodeToString([]byte{packet[i]})
		}
		list = append(list, r)
		i = i + 5 + r.Length
		if i >= len(packet) {
			break
		}
	}
	return list
}
