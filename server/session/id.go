package session

import (
	"crypto/rand"
	"encoding/hex"
	"hash/crc32"
	"strconv"
)

type Token string

// newToken generates a new session token
func newToken() Token {
	var buf = make([]byte, 16)
	_, _ = rand.Read(buf)
	checksum := crc32.ChecksumIEEE(buf)
	buf = append(buf, byte(checksum>>24&0xFF))
	buf = append(buf, byte(checksum>>16&0xFF))
	buf = append(buf, byte(checksum>>8&0xFF))
	buf = append(buf, byte(checksum&0xFF))
	return Token(hex.EncodeToString(buf))
}

// Valid returns true if the session token is valid
func (t Token) Valid() bool {
	if len(t) != 40 {
		return false
	}
	checksum, err := strconv.ParseUint(string(t[32:]), 16, 32)
	if err != nil {
		return false
	}
	data, err := hex.DecodeString(string(t[:32]))
	if err != nil {
		return false
	}
	sum := crc32.ChecksumIEEE(data)
	return uint32(checksum) == sum
}
