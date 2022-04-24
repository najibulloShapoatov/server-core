package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/pgtype"
	uuid "github.com/satori/go.uuid"
)

type UID string

type UIDsList []UID

func (a UIDsList) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	_ = ci
	if len(a) == 0 {
		return nil, nil
	}

	all := make([]string, 0, len(a))
	for _, id := range a {
		all = append(all, id.String())
	}
	b := "{" + strings.Join(all, ",") + "}"
	buf = append(buf, b...)
	return buf, nil
}

func (a UIDsList) Has(id UID) bool {
	for _, i := range a {
		if i == id {
			return true
		}
	}
	return false
}

func (id UID) String() string {
	if IsUIDHex(string(id)) {
		return string(id)
	}
	if !id.Valid() {
		return ""
	}
	return fmt.Sprintf(`%x`, string(id))
}

func (id UID) Hex() string {
	if IsUIDHex(string(id)) {
		return string(id)
	}
	return hex.EncodeToString([]byte(id))
}

func (id UID) MarshalJSON() ([]byte, error) {
	if IsUIDHex(string(id)) {
		return []byte(`"` + id + `"`), nil
	}
	return []byte(fmt.Sprintf(`"%x"`, string(id))), nil
}

func (id UID) Equals(u UID) bool {
	return strings.EqualFold(id.Hex(), u.Hex())
}

func (id *UID) UnmarshalJSON(data []byte) error {
	var v string
	err := json.Unmarshal(data, &v)
	if err == nil {
		data, err = hex.DecodeString(v)
		if err != nil {
			return err
		}
		*id = UID(data)
	}
	return nil
}

func (id UID) MarshalText() ([]byte, error) {
	if IsUIDHex(string(id)) {
		return []byte(id), nil
	}
	return []byte(fmt.Sprintf("%x", string(id))), nil
}

func (id UID) UUID() uuid.UUID {
	u := uuid.UUID{}
	copy(u[0:], id)
	return u
}

func (id UID) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	_ = ci
	v, _ := id.MarshalText()
	buf = append(buf, []byte(fmt.Sprintf("%s", v))...)
	return buf, nil
}

func (id *UID) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	_ = ci
	*id = UIDHex(string(src))
	return nil
}

// UnmarshalText turns *bson.ObjectId into an encoding.TextUnmarshaler.
func (id *UID) UnmarshalText(data []byte) error {
	if len(data) == 1 && data[0] == ' ' || len(data) == 0 {
		*id = ""
		return nil
	}
	if len(data) != 24 {
		return fmt.Errorf("invalid ObjectId: %s", data)
	}
	var buf [12]byte
	_, err := hex.Decode(buf[:], data[:])
	if err != nil {
		return fmt.Errorf("invalid ObjectId: %s (%s)", data, err)
	}
	*id = UID(string(buf[:]))
	return nil
}

func (id UID) Valid() bool {
	return len(id) == 12
}

func NewUID() UID {
	var b [12]byte
	// Timestamp, 4 bytes, big endian
	binary.BigEndian.PutUint32(b[:], uint32(time.Now().Unix()))
	// Machine, first 3 bytes of md5(hostname)
	b[4] = machineId[0]
	b[5] = machineId[1]
	b[6] = machineId[2]
	// Pid, 2 bytes, specs don't specify endianness, but we use big endian.
	b[7] = byte(processId >> 8)
	b[8] = byte(processId)
	// Increment, 3 bytes, big endian
	i := atomic.AddUint32(&objectIdCounter, 1)
	b[9] = byte(i >> 16)
	b[10] = byte(i >> 8)
	b[11] = byte(i)
	return UID(b[:])
}

func UIDHex(s string) UID {
	if IsHexadecimal(s) && IsUIDHex(s) {
		d, err := hex.DecodeString(s)
		if err != nil || len(d) != 12 {
			panic(fmt.Sprintf("invalid input to UID hex: %q", s))
		}
		return UID(d)
	} else {
		return UID(s)
	}
}

func IsUIDHex(s string) bool {
	if len(s) != 24 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

// objectIdCounter is atomically incremented when generating a new ObjectId
// using NewObjectId() function. It's used as a counter part of an id.
var objectIdCounter = readRandomUint32()

// readRandomUint32 returns a random objectIdCounter.
func readRandomUint32() uint32 {
	var b [4]byte
	_, err := io.ReadFull(rand.Reader, b[:])
	if err != nil {
		panic(fmt.Errorf("cannot read random object id: %v", err))
	}
	return (uint32(b[0]) << 0) | (uint32(b[1]) << 8) | (uint32(b[2]) << 16) | (uint32(b[3]) << 24)
}

// machineId stores machine id generated once and used in subsequent calls
// to NewObjectId function.
var machineId = readMachineId()
var processId = os.Getpid()

// readMachineId generates and returns a machine id.
// If this function fails to get the hostname it will cause a runtime error.
func readMachineId() []byte {
	var sum [3]byte
	id := sum[:]
	hostname, err1 := os.Hostname()
	if err1 != nil {
		_, err2 := io.ReadFull(rand.Reader, id)
		if err2 != nil {
			panic(fmt.Errorf("cannot get hostname: %v; %v", err1, err2))
		}
		return id
	}
	hw := md5.New()
	hw.Write([]byte(hostname))
	copy(id, hw.Sum(nil))
	return id
}

func init() {

}
