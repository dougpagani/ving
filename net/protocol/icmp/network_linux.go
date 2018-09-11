package icmp

import "net"

var networkType = map[string]string{
	"ipv4": "ip4:icmp",
	"ipv6": "ip6:ipv6-icmp",
}

func (p *IPing) buildDst(ipAddr *net.IPAddr) net.Addr {
	return ipAddr
}
