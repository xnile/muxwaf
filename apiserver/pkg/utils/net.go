package utils

import (
	"net"
	"net/netip"
	"regexp"
)

const (
	domainRegexString      = `^(?i)[a-z0-9-]+(\.[a-z0-9-]+)+\.?$`
	fqdnRegexStringRFC1123 = `^([a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62})(\.[a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62})*?(\.[a-zA-Z]{1}[a-zA-Z0-9]{0,62})\.?$`
)

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

func IsIPv4(s string) bool {
	ip := net.ParseIP(s)

	return ip != nil && ip.To4() != nil
}

func IsValidDomain(domain string) bool {
	domainRegexp := regexp.MustCompile(domainRegexString)
	return domainRegexp.MatchString(domain)
}

func IsFQDN(s string) bool {

	if s == "" {
		return false
	}
	fqdnRegexRFC1123 := regexp.MustCompile(fqdnRegexStringRFC1123)

	return fqdnRegexRFC1123.MatchString(s)
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
