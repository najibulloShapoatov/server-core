package utils

import (
	"errors"
	"fmt"
	"github.com/araddon/dateparse"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	reBlacklist   = regexp.MustCompile(`[^a-zA-Z0-9_ \-]`)
	reAccepted    = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	reReplaceable = regexp.MustCompile("[\\s\\-]+")
)

func ValidUID(uid string) bool {
	return reAccepted.MatchString(uid)
}

func GetUID(name string) (string, error) {
	val := strings.Trim(
		reReplaceable.ReplaceAllString(
			strings.TrimSpace(
				reBlacklist.ReplaceAllString(
					strings.TrimSpace(name), "",
				),
			),
			"_",
		),
		` _`)
	if val == "" {
		return "", fmt.Errorf("invalid uid")
	}
	return val, nil
}

func AsFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case *float64:
		return *v, nil
	case float32:
		return float64(v), nil
	case *float32:
		return float64(*v), nil
	default:
		ret, err := AsInt(v)
		if err == nil {
			return float64(ret), nil
		}
		str, err := AsString(value)
		if err != nil {
			return 0, errors.New("not a number")
		}
		val, err := strconv.ParseFloat(str, 10)
		if err == nil {
			return float64(val), nil
		}
		return 0, fmt.Errorf("not a number")
	}
}

func AsTime(value interface{}) (time.Time, error) {
	if v, ok := value.(time.Time); ok {
		return v, nil
	}
	if v, ok := value.(*time.Time); ok {
		return *v, nil
	}
	str, err := AsString(value)
	if err != nil {
		return time.Time{}, errors.New("invalid time value")
	}
	v, err := dateparse.ParseAny(str)
	if err != nil {
		return time.Time{}, errors.New("invalid time value")
	}
	return v, nil
}

func AsString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case *string:
		return *v, nil
	}
	return "", errors.New("not a string")
}

func AsBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case *bool:
		return *v, nil
	}
	return false, errors.New("not a bool")
}

func AsInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case *int64:
		return int(*v), nil
	case int32:
		return int(v), nil
	case *int32:
		return int(*v), nil
	case int16:
		return int(v), nil
	case *int16:
		return int(*v), nil
	case int8:
		return int(v), nil
	case *int8:
		return int(*v), nil
	case uint64:
		return int(v), nil
	case *uint64:
		return int(*v), nil
	case uint32:
		return int(v), nil
	case *uint32:
		return int(*v), nil
	case uint16:
		return int(v), nil
	case *uint16:
		return int(*v), nil
	case uint8:
		return int(v), nil
	case *uint8:
		return int(*v), nil
	default:
		str, err := AsString(value)
		if err != nil {
			return 0, errors.New("not a number")
		}
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return 0, errors.New("not a number")
		}
		return int(val), nil
	}
}
