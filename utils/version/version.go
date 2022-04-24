// Package version implements semantic version according to semver.org 2.0.0 specs
//
// More details about semver you can see at http://semver.org/
// The package can also process and compare version ranges (^1.2.3, ~1.2.3, *)
package version

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

var (
	errInvalidError = errors.New("invalid version format")
	// AppVersion represents the application version and is overwritten at compile time
	AppVersion = "1.0.0"
)

// Version contains a breakdown of a semver version number
type Version struct {
	Major      uint64
	Minor      uint64
	Patch      uint64
	PreRelease string
	Meta       string

	upTo *Version
}

// New parses a string into a Version object
func New(val string) (*Version, error) {
	v := &Version{}
	if err := v.parse(strings.TrimSpace(val)); err != nil {
		return nil, err
	}
	return v, nil
}

func (v *Version) parse(val string) error {
	// every version
	if val == "*" {
		v.upTo = &Version{Major: math.MaxUint64}
		return nil
	}
	// for version strings in the form of v1.2 v1.2.3 etc
	val = strings.TrimPrefix(val, "v")

	var hasRange = 0
	// for minor version ranges
	if strings.HasPrefix(val, "~") {
		hasRange = 2
		val = val[1:]
	}
	// for major version ranges
	if strings.HasPrefix(val, "^") {
		hasRange = 1
		val = val[1:]
	}

	// 1.2.3-alpha+sha.123
	if idx := strings.Index(val, "+"); idx != -1 {
		v.Meta = val[idx+1:]
		val = val[:idx]
	}

	// 1.2.3-alpha
	if idx := strings.Index(val, "-"); idx != -1 {
		v.PreRelease = val[idx+1:]
		val = val[:idx]
	}

	var err error
	parts := strings.Split(val, ".")
	switch len(parts) {
	case 3:
		if v.Patch, err = strconv.ParseUint(parts[2], 10, 64); err != nil {
			return errInvalidError
		}
		fallthrough
	case 2:
		if v.Major, err = strconv.ParseUint(parts[0], 10, 64); err != nil {
			return errInvalidError
		}
		if v.Minor, err = strconv.ParseUint(parts[1], 10, 64); err != nil {
			return errInvalidError
		}
	default:
		return errInvalidError
	}

	if hasRange == 1 {
		v.upTo = &Version{Major: v.Major + 1}
	} else if hasRange == 2 {
		v.upTo = &Version{Major: v.Major, Minor: v.Minor + 1}
	}
	return nil
}

func (v *Version) compare(val *Version) int {
	// quick check
	if v.String() == val.String() {
		return 0
	}
	// if major versions are not the same just stop
	if v.Major != val.Major {
		if v.Major > val.Major {
			return 1
		}
		return -1
	}
	if v.Minor != val.Minor {
		if v.Minor > val.Minor {
			return 1
		}
		return -1
	}
	if v.Patch != val.Patch {
		if v.Patch > val.Patch {
			return 1
		}
		return -1
	}
	if v.PreRelease != val.PreRelease {
		if v.PreRelease > val.PreRelease {
			return 1
		}
		return -1
	}
	if v.Meta != val.Meta {
		if v.Meta > val.Meta {
			return 1
		}
		return -1
	}

	return 0
}

func (v *Version) inRange(compare *Version) bool {
	if v.compare(compare) >= 0 && v.compare(compare.upTo) == -1 {
		return true
	}
	return false
}

// Equal accepts a version string or Version object and compares
// version equality. Equal will also return true if one of the versions
// have a range defined and contains the other one
// Eg:
//  1.2.3 eq ~1.2
//  1.9 eq ^1.2
//  1.2 eq *
func (v *Version) Equal(val interface{}) bool {
	compare := getVersion(val)
	if v.compare(compare) == 0 {
		return true
	}
	switch {
	case compare.upTo != nil && v.upTo == nil:
		return v.inRange(compare)
	case compare.upTo == nil && v.upTo != nil:
		return compare.inRange(v)
	case compare.upTo != nil && v.upTo != nil:
		if v.inRange(compare) || compare.inRange(v) {
			return true
		}
	}
	return false
}

// LessThan accepts a version string or Version object and tests if
// current version is less than given test value
func (v *Version) LessThan(val interface{}) bool {
	return v.compare(getVersion(val)) == -1
}

// LessEqThan accepts a version string or Version object and tests if
// current version is less or equal to the given test value
func (v *Version) LessEqThan(val interface{}) bool {
	return v.compare(getVersion(val)) <= 0
}

// GreaterThan accepts a version string or Version object and tests if
// current version is greater than given test value
func (v *Version) GreaterThan(val interface{}) bool {
	return v.compare(getVersion(val)) == 1
}

// GreaterEqThan accepts a version string or Version object and tests if
// current version is greater or equal to the given test value
func (v *Version) GreaterEqThan(val interface{}) bool {
	return v.compare(getVersion(val)) >= 0
}

// ReleaseMajor will increment major version and clear everything after it
// Eg. 1.3.4-alpha+sha.123 -> 2.0.0
func (v *Version) ReleaseMajor() {
	v.Major++
	v.Minor = 0
	v.Patch = 0
	v.PreRelease = ""
	v.Meta = ""
}

// ReleaseMinor will increment minor version and clear everything after it
// Eg. 1.3.4-alpha+sha.123 -> 1.4.0
func (v *Version) ReleaseMinor() {
	v.Minor++
	v.Patch = 0
	v.PreRelease = ""
	v.Meta = ""
}

// ReleasePatch will increment patch version and clear everything after it
// Eg. 1.3.4-alpha+sha.123 -> 1.3.5
func (v *Version) ReleasePatch() {
	v.Patch++
	v.PreRelease = ""
	v.Meta = ""
}

// ReleaseDev will change the PreRelease tag and clear any meta data after it
// Eg. 1.3.4-alpha+sha.123 ->beta-> 1.3.4-beta
func (v *Version) ReleaseDev(stage string) {
	v.PreRelease = stage
	v.Meta = ""
}

// Return a string representation of the version
func (v Version) String() string {
	var tmp = fmt.Sprintf("%d.%d", v.Major, v.Minor)
	if v.Patch != 0 {
		tmp = fmt.Sprintf("%s.%d", tmp, v.Patch)
	}
	if strings.TrimSpace(v.PreRelease) != "" {
		tmp = fmt.Sprintf("%s-%s", tmp, strings.TrimSpace(v.PreRelease))
	}
	if strings.TrimSpace(v.Meta) != "" {
		tmp = fmt.Sprintf("%s+%s", tmp, strings.TrimSpace(v.Meta))
	}
	return tmp
}

// MarshalJSON implements JSON Encoder interface
func (v Version) MarshalJSON() ([]byte, error) {
	return []byte(`"` + v.String() + `"`), nil
}

// UnmarshalJSON implements JSON Decoder interface
func (v *Version) UnmarshalJSON(data []byte) error {
	l := len(data)
	if l == 0 || string(data) == `""` {
		return nil
	}
	if l < 2 || data[0] != '"' || data[l-1] != '"' {
		return nil
	}
	return v.parse(string(data[1 : l-1]))
}

// UnmarshalYAML implements YAML Decoder interface
func (v *Version) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	return v.parse(s)
}

// MarshalYAML implements YAML Encoder interface
func (v *Version) MarshalYAML() (interface{}, error) {
	return v.String(), nil
}

func getVersion(val interface{}) *Version {
	if s, ok := val.(string); ok {
		if v, err := New(s); err == nil {
			return v
		}
		return &Version{}
	}
	if v, ok := val.(*Version); ok {
		return v
	}
	if v, ok := val.(Version); ok {
		return &Version{v.Major, v.Minor, v.Patch, v.PreRelease, v.Meta, v.upTo}
	}
	return &Version{}
}
