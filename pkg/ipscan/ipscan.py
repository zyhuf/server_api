#!/usr/bin/env python
"""
IP and MAC Scan interface handing for IPv4, IPv6
"""

#############################################################################
#                        scanning/Interfaces stuff                          #
#############################################################################

from scapy.all import srp,Ether,ARP,conf,IPv6,ICMPv6ND_NS

import optparse
import json

def IPv6FormatTrans(addr):
    tmpAddr = addr.split(':')
    for i in range(0, len(tmpAddr)):
        if ((len(tmpAddr[i]) > 0) and (len(tmpAddr[i]) < 4)):
            addrPart1 = tmpAddr[i]
            zero = 4 - len(tmpAddr[i])
            addrPart2 = ''.join('0' for i in range(0, zero))
            tmpAddr[i] = addrPart2 + addrPart1
    count = 0
    for i in range(0, len(tmpAddr)):
        count = count + len(tmpAddr[i])
    count = 32 - count
    addrPart = []
    addrPart = ''.join('0' for i in range(0, count))
    for i in range(1, len(tmpAddr) - 1):
        if len(tmpAddr[i]) == 0:
            tmpAddr[i] = addrPart
    result = ""
    for i in range(len(tmpAddr)):
        if i > 0:
            result += ":"
        num = len(tmpAddr[i]) / 4
        if num > 1:
            for j in range(num):
                result += "0000"
                if j < num-1:
                    result += ":"
        else:
            result += tmpAddr[i]
    return result
class IPScan:
    def __init__(self):
        pass

    def scan_IPv4(self, array):
        if isinstance(array, list) == False:
            raise TypeError
        dictIPv4 = {}
        for params in array:
            iface = conf.route.route(params)[0]
            ans, unans = srp(Ether(dst="FF:FF:FF:FF:FF:FF") / ARP(pdst=params), timeout=1, verbose=False, iface=iface)
            for snd, rcv in ans:
                macAddr = rcv.sprintf("%Ether.src%")
                IPv4Addr = rcv.sprintf("%ARP.psrc%")
                dictIPv4[IPv4Addr] = macAddr
        return dictIPv4
    def scan_IPv6(self, array):
        if isinstance(array, list) == False:
            raise TypeError
        dictIPv6 = {}
        for params in array:
            params = ''.join(params)
            if len(params) != 39:
                params = IPv6FormatTrans(params)
            dstMac = "33:33:FF:" + params[32:34] + ":" + params[35:37] + ":" + params[37:39]
            dstAddr = "FF02::1:FF" + params[32:34] + ":" + params[35:39]
            iface, srcAddr, _ = conf.route6.route(params)
            ans, unans = srp(Ether(dst=dstMac) / IPv6(src=srcAddr, dst=dstAddr) / ICMPv6ND_NS(tgt=params), timeout=1, verbose=False, iface=iface)
            for snd, rcv in ans:
                macAddr = rcv.sprintf("%Ether.src%")
                IPv6Addr = rcv.sprintf("%IPv6.src%")
                dictIPv6[IPv6Addr] = macAddr
        return dictIPv6
    def scan_ip_mac(self, addrRange):
        if isinstance(addrRange, list) == True:
            for addr in addrRange:
                addrRange = ''.join(addr)
                break
        start, end = addrRange.split('-')
        ipType = ""
        if ":" in start:
            ipType = "IPv6"
            if len(start) != 39:
                start = IPv6FormatTrans(start)
            if len(end) != 39:
                end = IPv6FormatTrans(end)
        else:
            ipType = "IPv4"

        startNum = ip2Num(start, ipType)
        endNum = ip2Num(end, ipType)
        addrList = [num2IP(num, ipType) for num in range(startNum,endNum+1)]
        if ipType == "IPv4":
            return json.dumps(self.scan_IPv4(addrList))
        elif ipType == "IPv6":
            return json.dumps(self.scan_IPv6(addrList))

def ip2Num(ip, ipType):
    if ipType == "IPv4":
        ips = [int(x) for x in ip.split('.')]
        return ips[0]<< 24 | ips[1]<< 16 | ips[2] << 8 | ips[3]
    elif ipType == "IPv6":
        ips = [int(x, 16) for x in ip.split(':')]
        return ips[0]<< 112 | ips[1]<< 96 | ips[2]<< 80 | ips[3]<< 64 | \
            ips[4]<< 48 | ips[5]<< 32 | ips[6]<< 16 | ips[7]

def num2IP (num, ipType):
    if ipType == "IPv4":
        return "%s.%s.%s.%s" % ((num >> 24) & 0xff, (num >> 16) & 0xff, (num >> 8) & 0xff, (num & 0xff))
    elif ipType == "IPv6":
         return '%04x:%04x:%04x:%04x:%04x:%04x:%04x:%04x' % ((num >> 112) & 0xffff, (num >> 96) & 0xffff, (num >> 80) & 0xffff, (num >> 64) & 0xffff, \
        (num >> 48) & 0xffff, (num >> 32) & 0xffff, (num >> 16) & 0xffff, (num & 0xffff))

def main():
    usage = "python process -H addrStart-addrEnd"
    parser = optparse.OptionParser(usage)
    parser.add_option('-H', dest='AddrRange', type='string', help='target address range')
    options, args = parser.parse_args()
    addrRange = options.AddrRange
    #ifce = options.iterface
    if addrRange == None:
        print(parser.usage)
        exit(0)
    else:
        ipScan = IPScan()
        result = ipScan.scan_ip_mac(addrRange)
        print(result)
if __name__ == "__main__":
    main()