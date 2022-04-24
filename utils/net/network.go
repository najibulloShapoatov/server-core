package net

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	HeaderXRealIP       = "X-Real-Ip"
	HeaderXForwardedFor = "X-Forwarded-For"
)

func GetClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	ra := r.RemoteAddr
	if ip := r.Header.Get(HeaderXForwardedFor); ip != "" {
		ra = strings.TrimSpace(strings.Split(ip, ",")[0])
	} else if ip := r.Header.Get(HeaderXRealIP); ip != "" {
		ra = strings.TrimSpace(strings.Split(ip, ",")[0])
	} else if strings.Contains(ra, ":") {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}

// GetLocalHostName returns the local hostname or the local IP if the hostname cannot be resolved
func GetLocalHostName() string {
	name, err := os.Hostname()
	if err != nil {
		name = GetLocalIP()
	}
	return name
}

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {

		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func CheckConnect(host string) error {
	conn, err := net.DialTimeout("tcp", host, time.Second*3)
	if err != nil {
		return nil
	}
	_ = conn.Close()
	return nil
}

func Lookup(host string) error {
	ips, err := net.LookupHost(host)
	if err != nil {
		return err
	}
	if len(ips) == 0 {
		return fmt.Errorf("host not found")
	}
	return nil
}

// GetMacAddr returns the local mac address
func GetMacAddr() (addr string) {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp != 0 && !bytes.Equal(i.HardwareAddr, nil) {
				// Don't use random as we have a real address
				addr = i.HardwareAddr.String()
				break
			}
		}
	}
	return
}

// BytesToMAC converts a binary representation of a MAC address to string representation
func BytesToMAC(mac []byte) string {
	var res = make([]string, 6)
	for i, b := range mac {
		res[i] = fmt.Sprintf("%02x", b)
	}
	return strings.Join(res, ":")
}

func IPInRange(ip string, list []string) bool {
	userIP := net.ParseIP(ip)
	if len(list) == 0 {
		return true
	}

	for _, against := range list {
		switch {
		case against == "*":
			return true
		case strings.Contains(against, "/"):
			_, in, e := net.ParseCIDR(against)
			if e != nil {
				continue
			}
			if in.Contains(userIP) {
				return true
			}
		case strings.Contains(against, "-"):
			ipRange := strings.Split(against, "-")
			start, _ := strconv.Atoi(strings.Split(ipRange[0], ".")[3])
			end, _ := strconv.Atoi(ipRange[1])
			between, _ := strconv.Atoi(strings.Split(ip, ".")[3])
			if between >= start && between <= end {
				return true
			}
		default:
			if net.ParseIP(against).Equal(userIP) {
				return true
			}
		}
	}
	return false
}

// MACToByte converts a binary representation of a MAC address to string representation
func MACToByte(mac string) []byte {
	var res = make([]byte, 6)
	_, _ = fmt.Sscanf(mac, "%x:%x:%x:%x:%x:%x", &res[0], &res[1], &res[2], &res[3], &res[4], &res[5])
	return res
}

// IPToInt encodes a string representation of a IP to a uint32
func IPToInt(ip string) uint32 {
	i := net.ParseIP(ip)
	if len(i) == 16 {
		return binary.BigEndian.Uint32(i[12:16])
	}
	return binary.BigEndian.Uint32(i)
}

// IntToIP decodes a uint32 encoded IP to a string representation
func IntToIP(ip uint32) string {
	i := make(net.IP, 4)
	binary.BigEndian.PutUint32(i, ip)
	return i.String()
}

// CIDRMatch checks if specific ip belongs to specific cidr block (ip/mask)
func CIDRMatch(ip net.IP, cidr string) bool {
	_, cidrNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	return cidrNet.Contains(ip)
}

func GetLocalAddr() string {
	ifaces, _ := net.Interfaces()
	var ipv6 = make([]string, 0)
	var ipv4 = make([]string, 0)

	// handle err
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// process IP address
			if ip != nil {
				if ip.IsLoopback() {
					continue
				}
				b, _ := ip.MarshalText()
				ipStr := string(b)
				if strings.Contains(ipStr, ":") {
					ipv6 = append(ipv6, ipStr)
				} else {
					ipv4 = append(ipv4, ipStr)
				}
			}
		}
	}

	if len(ipv4) != 0 {
		return ipv4[0]
	}

	if len(ipv6) != 0 {
		return ipv6[0]
	}

	return "0.0.0.0"
}

// BetweenMatch checks if ip belongs to specific IP Start-End range (ip-ip)
func BetweenMatch(addr net.IP, block string) bool {
	cleanStr := strings.Join(strings.Fields(block), "")
	ips := strings.Split(cleanStr, "-")
	res := IPBetween(net.ParseIP(ips[0]), net.ParseIP(ips[1]), addr)

	return res
}

// IPBetween does determine if a given ip is between two others (inclusive)
func IPBetween(from net.IP, to net.IP, userIp net.IP) bool {
	if from == nil || to == nil || userIp == nil {
		return false
	}

	from16 := from.To16()
	to16 := to.To16()
	user16 := userIp.To16()
	if from16 == nil || to16 == nil || user16 == nil {
		// An ip did not convert to a 16 byte
		return false
	}

	if bytes.Compare(user16, from16) >= 0 && bytes.Compare(user16, to16) <= 0 {
		return true
	}

	return false
}

// WildcardMatch checks if specific ip belongs to specific wildcard IP block
func WildcardMatch(userIP net.IP, wildcardIP string) bool {
	// IPv4 wildcard A.B.*.* format
	// converts to A-B format by setting * to 0 for start and 255 for end
	// for 127.0.*.* check range will be 127.0.0.0 - 127.0.255.255
	if from := net.ParseIP(strings.ReplaceAll(wildcardIP, "*", "0")); from.To4() != nil {
		to := strings.ReplaceAll(wildcardIP, "*", "255")
		res := IPBetween(from, net.ParseIP(to), userIP)

		return res
	}
	return false
}
