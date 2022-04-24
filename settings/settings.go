package settings

import (
	"errors"
	"github.com/najibulloShapoatov/server-core/utils/reflection"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Settings struct {
	lock sync.RWMutex
	data map[string]string
}

// Loader is used to load and parse configuration values from various formats and location
type Loader interface {
	// Parse method is called
	Parse() (map[string]string, error)
}

// single instance of the Settings structure
var instance = &Settings{data: make(map[string]string)}

// Instantiate new settings reader
func GetSettings() *Settings {
	return instance
}

// Has returns true if the given key exists
func (s *Settings) Has(key string) bool {
	s.lock.Lock()
	_, ok := s.data[key]
	s.lock.Unlock()
	return ok
}

// Load runs the given loaders in order to load and parse the configuration values. The first loader
// that returns an error stops the load process
func (s *Settings) Load(loaders ...Loader) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data = make(map[string]string)

	// load content and parse and store in data
	for _, loader := range loaders {
		values, err := loader.Parse()
		if err != nil {
			return err
		}
		for k, v := range values {
			s.data[k] = v
		}
	}
	return nil
}

// GetString returns the value at the given key as a string and true if the key exists
// or empty string and false if the key doesn't exist
func (s *Settings) GetString(key string) (string, bool) {
	s.lock.RLock()
	val, ok := s.data[key]
	s.lock.RUnlock()
	return s.resolveVar(val), ok
}

// GetFloat returns the value at the given key parsed as a float and true if the key exists
// or 0.0 and false if the key doesn't exist or failed to parse as a float64
func (s *Settings) GetFloat(key string) (float64, bool) {
	s.lock.RLock()
	val, ok := s.data[key]
	s.lock.RUnlock()

	if !ok {
		return 0, false
	}
	intVal, err := strconv.ParseFloat(s.resolveVar(val), 64)
	if err != nil {
		return 0, false
	}
	return intVal, true
}

// GetInt returns the value at the given key parsed as a int and true if the key exists
// or 0 and false if the key doesn't exist or failed to parse as a int
func (s *Settings) GetInt(key string) (int, bool) {
	v, ok := s.GetFloat(key)
	return int(v), ok
}

// GetDuration returns the value at the given key parsed as Duration and true if the key exists
// or Duration(0) and false if the key doesn't exist or failed to parse as Duration
func (s *Settings) GetDuration(key string) (time.Duration, bool) {
	s.lock.RLock()
	val, ok := s.data[key]
	s.lock.RUnlock()
	if !ok {
		return 0, false
	}
	duration, err := ParseDuration(s.resolveVar(val))
	if err != nil {
		return 0, false
	}
	return duration, true
}

func (s *Settings) resolveVar(val string) string {
	reg := regexp.MustCompile(`(\${.[^}]*})`)
	matches := reg.FindAllStringSubmatch(val, -1)
	for _, match := range matches {
		env := strings.TrimSuffix(strings.TrimPrefix(match[0], "${"), "}")
		val = strings.ReplaceAll(val, match[0], os.Getenv(env))
	}

	if strings.HasPrefix(val, "${") && strings.HasSuffix(val, "}") {
		val = strings.TrimSuffix(strings.TrimPrefix(val, "${"), "}")
		return os.Getenv(val)
	}
	return val
}

func (s *Settings) GetKeys() (res []string) {
	s.lock.RLock()
	for k := range s.data {
		res = append(res, k)
	}
	s.lock.RUnlock()
	return
}

var truthTable = map[string]bool{
	"yes":     true,
	"on":      true,
	"true":    true,
	"1":       true,
	"active":  true,
	"set":     true,
	"enabled": true,

	"no":       false,
	"off":      false,
	"false":    false,
	"0":        false,
	"inactive": false,
	"unset":    false,
	"disabled": false,
}

// GetBool returns the value at the given key parsed as bool and true if the key exists
// or false and false if the key doesn't exist or failed to parse as a truthy value
func (s *Settings) GetBool(key string) (bool, bool) {
	s.lock.RLock()
	val, ok := s.data[key]
	s.lock.RUnlock()
	if !ok {
		return false, false
	}
	boolVal, ok := truthTable[strings.ToLower(s.resolveVar(val))]
	if !ok {
		return false, false
	}
	return boolVal, true
}

// Unmarshal decodes the configuration in a structure based on the `config` and `default` tags
func (s *Settings) Unmarshal(destinationPtr interface{}) error {
	rv := reflect.ValueOf(destinationPtr)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("pointer_required")
	}
	pv := reflection.Indirect(rv, false)
	var v = pv
	t := v.Type()

	var defKey = "\x00\x01"

	defer delete(s.data, defKey)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		defValue := field.Tag.Get("default")
		cfgKey := field.Tag.Get("config")

		if defValue == "" && cfgKey == "" {
			continue
		}
		s.data[defKey] = defValue

		fv := reflection.Indirect(pv.Field(i), false)

		if fv.CanSet() {
			var v reflect.Value
			if fv.Type().AssignableTo(reflect.TypeOf(time.Duration(0))) {
				fv.Set(decode(reflect.ValueOf(s.GetDuration), cfgKey, defValue))
			} else {
				switch fv.Kind() {
				case reflect.Bool:
					v = decode(reflect.ValueOf(s.GetBool), cfgKey, defValue)
				case reflect.String:
					v = decode(reflect.ValueOf(s.GetString), cfgKey, defValue)
				case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
					reflect.Uint, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
					v = decode(reflect.ValueOf(s.GetInt), cfgKey, defValue).Convert(fv.Type())
				case reflect.Float32, reflect.Float64:
					v = decode(reflect.ValueOf(s.GetFloat), cfgKey, defValue).Convert(fv.Type())
				case reflect.Struct:
					_ = s.Unmarshal(fv.Addr().Interface())
				}

				if v.IsValid() {
					fv.Set(v)
				}
			}
		}
	}
	return nil
}

func decode(fn reflect.Value, cfgKey, defValue string) reflect.Value {
	if fn.Kind() == reflect.Func {
		out := fn.Call([]reflect.Value{reflect.ValueOf(cfgKey)})
		if out[1].Bool() {
			return out[0]
		} else if defValue != "" {
			out := fn.Call([]reflect.Value{reflect.ValueOf("\x00\x01")})
			return out[0]
		}
	}
	return reflect.Value{}
}

var unitMap = map[string]int64{
	"ns": int64(time.Nanosecond),
	"us": int64(time.Microsecond),
	"µs": int64(time.Microsecond), // U+00B5 = micro symbol
	"μs": int64(time.Microsecond), // U+03BC = Greek letter mu
	"ms": int64(time.Millisecond),
	"s":  int64(time.Second),
	"m":  int64(time.Minute),
	"h":  int64(time.Hour),
	"d":  int64(time.Hour * 24),
	"w":  int64(time.Hour * 24 * 7),
	"M":  int64(time.Hour * 24 * 7 * 30),
	"y":  int64(time.Hour * 24 * 7 * 30 * 12),
}

var errLeadingInt = errors.New("time: bad [0-9]*") // never printed

// leadingInt consumes the leading [0-9]* from s.
func leadingInt(s string) (x int64, rem string, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > (1<<63-1)/10 {
			// overflow
			return 0, "", errLeadingInt
		}
		x = x*10 + int64(c) - '0'
		if x < 0 {
			// overflow
			return 0, "", errLeadingInt
		}
	}
	return x, s[i:], nil
}

// leadingFraction consumes the leading [0-9]* from s.
// It is used only for fractions, so does not return an error on overflow,
// it just stops accumulating precision.
func leadingFraction(s string) (x int64, scale float64, rem string) {
	i := 0
	scale = 1
	overflow := false
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if overflow {
			continue
		}
		if x > (1<<63-1)/10 {
			// It's possible for overflow to give a positive number, so take care.
			overflow = true
			continue
		}
		y := x*10 + int64(c) - '0'
		if y < 0 {
			overflow = true
			continue
		}
		x = y
		scale *= 10
	}
	return x, scale, s[i:]
}

// ParseDuration parses a duration string.
// A duration string is a possibly signed sequence of
// decimal numbers, each with optional fraction and a unit suffix,
// such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
func ParseDuration(s string) (time.Duration, error) {
	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
	orig := s
	var d int64
	neg := false

	// Consume [-+]?
	if s != "" {
		c := s[0]
		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}
	// Special case: if all that is left is "0", this is zero.
	if s == "0" {
		return 0, nil
	}
	if s == "" {
		return 0, errors.New("time: invalid duration " + orig)
	}
	for s != "" {
		var (
			v, f  int64       // integers before, after decimal point
			scale float64 = 1 // value = v + f/scale
		)

		var err error

		// The next character must be [0-9.]
		if !(s[0] == '.' || '0' <= s[0] && s[0] <= '9') {
			return 0, errors.New("time: invalid duration " + orig)
		}
		// Consume [0-9]*
		pl := len(s)
		v, s, err = leadingInt(s)
		if err != nil {
			return 0, errors.New("time: invalid duration " + orig)
		}
		pre := pl != len(s) // whether we consumed anything before a period

		// Consume (\.[0-9]*)?
		post := false
		if s != "" && s[0] == '.' {
			s = s[1:]
			pl := len(s)
			f, scale, s = leadingFraction(s)
			post = pl != len(s)
		}
		if !pre && !post {
			// no digits (e.g. ".s" or "-.s")
			return 0, errors.New("time: invalid duration " + orig)
		}

		// Consume unit.
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			if c == '.' || '0' <= c && c <= '9' {
				break
			}
		}
		if i == 0 {
			return 0, errors.New("time: missing unit in duration " + orig)
		}
		u := s[:i]
		s = s[i:]
		unit, ok := unitMap[u]
		if !ok {
			return 0, errors.New("time: unknown unit " + u + " in duration " + orig)
		}
		if v > (1<<63-1)/unit {
			// overflow
			return 0, errors.New("time: invalid duration " + orig)
		}
		v *= unit
		if f > 0 {
			// float64 is needed to be nanosecond accurate for fractions of hours.
			// v >= 0 && (f*unit/scale) <= 3.6e+12 (ns/h, h is the largest unit)
			v += int64(float64(f) * (float64(unit) / scale))
			if v < 0 {
				// overflow
				return 0, errors.New("time: invalid duration " + orig)
			}
		}
		d += v
		if d < 0 {
			// overflow
			return 0, errors.New("time: invalid duration " + orig)
		}
	}

	if neg {
		d = -d
	}
	return time.Duration(d), nil
}
