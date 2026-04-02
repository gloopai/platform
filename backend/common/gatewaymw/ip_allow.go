package gatewaymw

import (
	"net"
	"strings"
)

// IPAllowed checks whether clientHost is allowed by a comma-separated whitelist.
// Empty whitelist allows all. IPv4/IPv6 and CIDR entries are supported.
func IPAllowed(clientHost string, whitelist string) bool {
	whitelist = strings.TrimSpace(whitelist)
	if whitelist == "" {
		return true
	}
	host := strings.TrimSpace(clientHost)
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, item := range strings.Split(whitelist, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if strings.Contains(item, "/") {
			_, cidr, err := net.ParseCIDR(item)
			if err == nil && cidr.Contains(ip) {
				return true
			}
			continue
		}
		if net.ParseIP(item) != nil && item == host {
			return true
		}
	}
	return false
}
