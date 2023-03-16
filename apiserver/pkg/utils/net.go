package utils

import (
	"net"
	"net/netip"
	"regexp"
)

var domainRegexp = regexp.MustCompile(`^(?i)[a-z0-9-]+(\.[a-z0-9-]+)+\.?$`)

type IPNet struct {
	IP  *netip.Addr
	Net *netip.Prefix
	V4  bool
	V6  bool
}

func ParseIPorCIDR(s string) IPNet {
	var ipNet IPNet
	if addr, err := netip.ParseAddr(s); err == nil {
		ipNet.IP = &addr
		if addr.Is4() {
			ipNet.V4 = true
		}
		if addr.Is6() {
			ipNet.V6 = true
		}
	}
	if prefix, err := netip.ParsePrefix(s); err == nil {
		ipNet.Net = &prefix
		if prefix.Addr().Is4() {
			ipNet.V4 = true
		}
		if prefix.Addr().Is6() {
			ipNet.V6 = true
		}
	}
	return ipNet
}

func IsValidDomain(domain string) bool {
	return domainRegexp.MatchString(domain)
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return net.IP{}
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}
