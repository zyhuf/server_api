package ipscan

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/routing"
)

type ipToHwAddr struct {
	ipAddr string
	hwAddr string
}

func IPScan(addrRange string) string {
	addr := ipv4RangeTrans(addrRange)
	if len(addr) == 0 {
		return ""
	}
	byteIP := net.ParseIP(addr[0])
	scanner, err := newScanner(byteIP)
	if err != nil {
		log.Println(err)
		return ""
	}

	var ip2Hw []*ipToHwAddr
	for i := 0; i < len(addr); i++ {
		byteIP = net.ParseIP(addr[i]).To4()
		TmpHwAddr, err := scanner.getHwAddrForARP(byteIP)
		if err != nil {
			log.Println(err)
			continue
		}
		tmpIp2Hw := new(ipToHwAddr)
		tmpIp2Hw.ipAddr = addr[i]
		tmpIp2Hw.hwAddr = TmpHwAddr.String()
		ip2Hw = append(ip2Hw, tmpIp2Hw)
		log.Println(tmpIp2Hw)
	}
	bytes_data, err := json.Marshal(ip2Hw)

	return string(bytes_data)
}

type scanner struct {
	iface   *net.Interface
	srcAddr net.IP
	handle  *pcap.Handle

	opts gopacket.SerializeOptions
	buf  gopacket.SerializeBuffer
}

func ipv4ToNum(addr string) uint32 {
	ips := strings.Split(addr, ".")
	ips1, _ := strconv.Atoi(ips[0])
	ips2, _ := strconv.Atoi(ips[1])
	ips3, _ := strconv.Atoi(ips[2])
	ips4, _ := strconv.Atoi(ips[3])

	return uint32(ips1<<24 | ips2<<16 | ips3<<8 | ips4)
}

func numToIPv4(num uint32) string {
	return fmt.Sprintf("%v.%v.%v.%v", (num>>24)&0xff, (num>>16)&0xff, (num>>8)&0xff, (num & 0xff))
}

func ipv4RangeTrans(addrRange string) []string {
	addr := strings.Split(addrRange, "-")
	addrStart := ipv4ToNum(addr[0])
	addrEnd := ipv4ToNum(addr[1])

	var ipAddr []string
	for num := addrStart; num <= addrEnd; num++ {
		ipAddr = append(ipAddr, numToIPv4(num))
	}

	return ipAddr
}

func newScanner(ip net.IP) (*scanner, error) {
	s := &scanner{
		opts: gopacket.SerializeOptions{
			FixLengths:       true,
			ComputeChecksums: true,
		},
		buf: gopacket.NewSerializeBuffer(),
	}

	router, err := routing.New()
	if err != nil {
		log.Println("routing error:", err)
		return nil, err
	}

	iface, _, src, err := router.Route(ip)
	if err != nil {
		return nil, err
	}
	s.srcAddr, s.iface = src, iface

	handle, err := pcap.OpenLive(iface.Name, 65536, true, pcap.BlockForever)
	if err != nil {
		return nil, err
	}
	s.handle = handle
	return s, nil
}

func (s *scanner) close() {
	s.handle.Close()
}

func (s *scanner) getHwAddrForARP(dstAddr net.IP) (net.HardwareAddr, error) {
	start := time.Now()

	eth := layers.Ethernet{
		SrcMAC:       s.iface.HardwareAddr,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(s.iface.HardwareAddr),
		SourceProtAddress: []byte(s.srcAddr),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
		DstProtAddress:    []byte(dstAddr),
	}

	if err := s.send(&eth, &arp); err != nil {
		return nil, err
	}

	// Wait 1 seconds for an ARP reply.
	for {
		if time.Since(start) > time.Second*1 {
			return nil, errors.New("timeout getting ARP reply")
		}
		data, _, err := s.handle.ReadPacketData()
		if err == pcap.NextErrorTimeoutExpired {
			continue
		} else if err != nil {
			return nil, err
		}
		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
		if arpLayer := packet.Layer(layers.LayerTypeARP); arpLayer != nil {
			arp := arpLayer.(*layers.ARP)
			if net.IP(arp.SourceProtAddress).Equal(net.IP(dstAddr)) {
				return net.HardwareAddr(arp.SourceHwAddress), nil
			}
		}
	}
}

func (s *scanner) send(l ...gopacket.SerializableLayer) error {
	if err := gopacket.SerializeLayers(s.buf, s.opts, l...); err != nil {
		return err
	}
	return s.handle.WritePacketData(s.buf.Bytes())
}
