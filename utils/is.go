package utils

import (
	"fmt"
	"net"
	"net/smtp"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Used by IsFilePath func
const (
	// Unknown is unresolved OS type
	UnknownFileType = iota
	// Win is Windows type
	WinFileType
	// Unix is *nix OS types
	UnixFileType
)

// IsInt check if the string is an integer.
func IsInt(str string) bool {
	return Matches(str, "^(?:-?(?:0|[1-9][0-9]*))$")
}

// IsFloat check if the string is a float.
func IsFloat(str string) bool {
	return str != "" && Matches(str, "^(?:-?(?:[0-9]+))?(?:\\.[0-9]*)?(?:[eE][\\+\\-]?(?:[0-9]+))?$")
}

// IsNull check if the string is null.
func IsNull(str string) bool {
	return len(str) == 0
}

var emailRegexp = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

// IsEmailFormat checks if the string is a valid email address format
func IsEmailFormat(str string) bool {
	return emailRegexp.MatchString(str)
}

var phoneNumberRegexp = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)

// IsPhoneNumberFormat checks if the string is a valid phone number format
func IsPhoneNumberFormat(str string) bool {
	return phoneNumberRegexp.MatchString(str)
}

var netLookupMX = net.LookupMX

type dialer interface {
	Close() error
	Hello(localName string) error
	Mail(from string) error
	Rcpt(to string) error
}

// IsValidEmailHost tries to do a SMTP connection on the email host to validate if it's a valid email address or not
func IsValidEmailHost(email string) bool {
	_, host := SplitEmailToAccountAndDomain(email)
	if host == "" {
		return false
	}

	tries := 4
	okChan := make(chan error)

	for i := 1; i <= tries; i++ {
		go func(i int) {
			timeout := time.Duration(i*2) * time.Second
			okChan <- checkEmail(email, host, timeout)
		}(i)
	}

	for i := 1; i <= tries; i++ {
		if err := <-okChan; err == nil {
			return true
		}
	}

	return false
}

var ipRangeRe = "^(([(\\d+)(x+)]){1,3})(\\-+([(\\d+)(x)]{1,3}))?\\.(([(\\d+)(x+)]){1,3})(\\-+([(\\d+)(x)]{1,3}))?\\.(([(\\d+)(x+)]){1,3})(\\-+([(\\d+)(x)]{1,3}))?\\.(([(\\d+)(x+)]){1,3})(\\-+([(\\d+)(x)]{1,3}))?$"

func IsValidIP(ip string) bool {
	if ip == "*" {
		return true
	}
	if x := net.ParseIP(ip); x != nil {
		return true
	}
	if _, _, err := net.ParseCIDR(ip); err == nil {
		return true
	}
	if match, _ := regexp.MatchString(ipRangeRe, ip); match {
		ipRange := strings.Split(ip, "-")
		left, right := 0, 0
		for _, i := range strings.Split(ipRange[0], ".") {
			if left, _ = strconv.Atoi(i); left > 255 {
				return false
			}
		}
		for _, i := range strings.Split(ipRange[1], ".") {
			if right, _ = strconv.Atoi(i); right > 255 {
				return false
			}
		}

		if right <= left {
			return false
		}

		return true
	}

	return false
}

// IsIPInRange returns true if checkIP is in any of the specified ips or cidr ranges from rangeIPs
func IsIPInRange(rangeIPs []string, checkIP string) bool {
	ip := net.ParseIP(checkIP)

	for _, against := range rangeIPs {
		against = strings.TrimSpace(against)
		if against == "*" || strings.HasPrefix(checkIP, "127.") || strings.HasPrefix(checkIP, "192.") || strings.HasPrefix(checkIP, "10.") || checkIP == "::1" {
			return true
		}
		if x := net.ParseIP(against); x != nil && ip.Equal(x) {
			return true
		}
		if _, ipNet, err := net.ParseCIDR(against); err == nil && ipNet.Contains(ip) {
			return true
		}
		if strings.Contains(against, "-") {
			ipRange := strings.Split(against, "-")
			start, _ := strconv.Atoi(strings.Split(ipRange[0], ".")[3])
			end, _ := strconv.Atoi(ipRange[1])
			between, _ := strconv.Atoi(strings.Split(checkIP, ".")[3])
			if between >= start && between <= end {
				return true
			}
		}
	}
	return false
}

func checkEmail(email, host string, timeout time.Duration) error {
	mx, err := netLookupMX(host)
	if err != nil {
		return err
	}

	client, err := smtpClient(fmt.Sprintf("%s:%d", mx[0].Host, 25), timeout)
	if err != nil {
		return err
	}

	defer client.Close()

	if err = client.Hello("checkmail.me"); err != nil {
		return err
	}
	if err = client.Mail("just-testing@gmail.com"); err != nil {
		return err
	}
	return client.Rcpt(email)
}

// SplitEmailToAccountAndDomain splits an email address into account name and hostname
func SplitEmailToAccountAndDomain(email string) (account, host string) {
	i := strings.LastIndexByte(email, '@')
	if i == -1 {
		return
	}
	account = strings.TrimSpace(email[:i])
	host = strings.TrimSpace(email[i+1:])
	return
}

var smtpClient = func(addr string, timeout time.Duration) (dialer, error) {
	// Dial the tcp connection
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}

	// Connect to the SMTP server
	c, err := smtp.NewClient(conn, addr)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// IsHexadecimal check if the string is a hexadecimal number.
func IsHexadecimal(str string) bool {
	return Matches(str, "^[0-9a-fA-F]+$")
}

// IsBitstring check if the string is a bit string.
func IsBitstring(str string) bool {
	return Matches(str, "^[0-1]+$")
}

// IsPointer checks if given interface is a pointer
func IsPointer(obj interface{}) bool {
	if reflect.TypeOf(obj) == nil {
		return false
	}
	return reflect.TypeOf(obj).Kind() == reflect.Ptr
}

// IsSimpleType checks if given t ype is a simple Go type
func IsSimpleType(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool:
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	case reflect.String:
		return true
	}
	return false
}

// IsTruthy checks if the given input is a value that can mean true or false
func IsTruthy(input string) bool {
	_, ok := boolString[input]
	return ok
}

// IsFilePath check is a string is Win or Unix file path and returns it's type.
func IsFilePath(str string) (bool, int) {
	if Matches(str, `^[a-zA-Z]:\\(?:[^\\/:*?"<>|\r\n]+\\)*[^\\/:*?"<>|\r\n]*$`) {
		// check windows path limit see:
		//  http://msdn.microsoft.com/en-us/library/aa365247(VS.85).aspx#maxpath
		if len(str[3:]) > 32767 {
			return false, WinFileType
		}
		return true, WinFileType
	} else if Matches(str, `^((?:\/[a-zA-Z0-9\.\:]+(?:_[a-zA-Z0-9\:\.]+)*(?:\-[\:a-zA-Z0-9\.]+)*)+\/?)$`) {
		return true, UnixFileType
	}
	return false, UnknownFileType
}

// IsDNSName will validate the given string as a DNS name
func IsDNSName(str string) bool {
	if str == "" || len(strings.Replace(str, ".", "", -1)) > 255 {
		// constraints already violated
		return false
	}
	return Matches(str, `^([a-zA-Z0-9]{1}[a-zA-Z0-9_-]{1,62}){1}(.[a-zA-Z0-9]{1}[a-zA-Z0-9_-]{1,62})*$`)
}

// IsDialString validates the given string for usage with the various Dial() functions
func IsDialString(str string) bool {
	if h, p, err := net.SplitHostPort(str); err == nil && h != "" && p != "" && (IsDNSName(h) || IsIP(h)) && IsPort(p) {
		return true
	}
	return false
}

// IsIP checks if a string is either IP version 4 or 6 using the net package parser
func IsIP(str string) bool {
	return net.ParseIP(str) != nil
}

// IsPort checks if a string represents a valid port
func IsPort(str string) bool {
	if i, err := strconv.Atoi(str); err == nil && i > 0 && i < 65536 {
		return true
	}
	return false
}
