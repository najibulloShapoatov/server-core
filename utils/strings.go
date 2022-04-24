package utils

import (
	"crypto/rand"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Matches check if string matches the pattern (pattern is regular expression)
// In case of error return false
func Matches(str, pattern string) bool {
	match, _ := regexp.MatchString(pattern, str)
	return match
}

// map of string representations of boolean values
var boolString = map[string]bool{
	"true":    true,
	"yes":     true,
	"on":      true,
	"1":       true,
	"toggled": true,
	"active":  true,
	"present": true,

	"nan":       false,
	"undefined": false,
	"null":      false,
	"":          false,
	"false":     false,
	"no":        false,
	"off":       false,
	"0":         false,
	"untoggled": false,
	"inactive":  false,
	"absent":    false,
	"missing":   false,
}

// Truthy converts string values to their boolean representations.
// If input string is not found in the list of defined truthy values, false is returned.
//
// FALSE values: NaN, undefined, null, "" (empty string), false, no, off,  0, untoggled, inactive
// TRUE values: true, yes, on, 1, toggled, active
func Truthy(input string) bool {
	input = strings.TrimSpace(strings.ToLower(input))
	if val, ok := boolString[input]; ok {
		return val
	}
	return false
}

const (
	// StdLen is the minimum string length required to achieve ~95 bits of entropy.
	StdLen = 16
	// MaxLen is the maximum length possible for generated strings
	MaxLen = 1048576
)

// RandomString returns a new random string of the standard length, consisting of
// standard characters.
func RandomString() string {
	return newLenChars(StdLen, StdChars)
}

// RandomStringLen returns a new random string of the provided length (0 < len < MaxLen), consisting of
// standard characters.
func RandomStringLen(length int) string {
	return newLenChars(length, StdChars)
}

// NewLenChars returns a new random string of the provided length, consisting
// of the provided byte slice of allowed characters (maximum 256).
func newLenChars(length int, chars []byte) string {
	length = int(math.Abs(float64(length)))

	if length == 0 {
		return ""
	}

	if length > MaxLen {
		length = MaxLen
	}

	b := make([]byte, length)
	r := make([]byte, length+(length/4))

	clen := byte(len(chars))
	maxrb := byte(256 - (256 % len(chars)))

	i := 0
	for {
		io.ReadFull(rand.Reader, r)
		for _, c := range r {
			if c >= maxrb {
				// Skip this number to avoid modulo bias.
				continue
			}
			b[i] = chars[c%clen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}

// TrimAndLower returns a lowercase representation of the input
func TrimAndLower(str string) string {
	return strings.ToLower(strings.TrimSpace(str))
}

// StdChars is a set of standard characters allowed in Base64 string.
var StdChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

// ISO3166Entry stores country codes
type ISO3166Entry struct {
	EnglishShortName string
	FrenchShortName  string
	Alpha2Code       string
	Alpha3Code       string
	Numeric          string
}

// ToStringSlice converts the int slice to a string slice
func ToStringSlice(list []int) []string {
	res := make([]string, len(list))
	for i, id := range list {
		res[i] = strconv.Itoa(id)
	}
	return res
}
