package icmp

import "net"

var networkType = map[string]string{
	"ipv4": "udp4",
	"ipv6": "udp6",
}

func (p *IPing) buildDst(ipAddr *net.IPAddr) net.Addr {
	return &net.UDPAddr{IP: ipAddr.IP, Zone: ipAddr.Zone}
}
