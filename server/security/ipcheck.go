package security

import (
	"bytes"
	"net"
	"strings"

	netutils "github.com/najibulloShapoatov/server-core/utils/net"
)

// checkIP checks if IP is in list
func CheckIP(ip string, list []string) bool {
	// Check if ip si valid IP address version 4 or 6
	userIP := net.ParseIP(ip)
	if userIP != nil {
		// check IP in list
		for _, addr := range list {
			_, _, err := net.ParseCIDR(addr)

			switch {
			case err == nil:
				if netutils.CIDRMatch(userIP, addr) {
					return true
				}
			case strings.Contains(addr, "-"):
				if netutils.BetweenMatch(userIP, addr) {
					return true
				}
			case strings.Contains(addr, "*"):
				if netutils.WildcardMatch(userIP, addr) {
					return true
				}
			default:
				if bytes.Equal(userIP, net.ParseIP(addr)) {
					return true
				}
			}
		}
		return false
	}
	return false
}
