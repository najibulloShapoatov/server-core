package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const timeFormat = "2006-01-02 15:04:05.9999"

type nilFormatter struct {
	tmp []byte
}

// NewNilFormatter will return a Formatter that does nothing and returns an empty byte slice every time
func NewNilFormatter() Formatter {
	return &nilFormatter{make([]byte, 0)}
}

func (f *nilFormatter) Format(Entry *Entry) []byte {
	return f.tmp
}

type textFormatter struct {
	keys *bytes.Buffer
}

// NewTextFormatter will return encode the Entry along with the default provided keys as key=value pairs
func NewTextFormatter(defaultKeys map[string]interface{}) Formatter {
	buf := &bytes.Buffer{}
	for k, v := range defaultKeys {
		encode(buf, k, v)
	}
	return &textFormatter{keys: buf}
}

func (f *textFormatter) Format(entry *Entry) []byte {
	var buf = &bytes.Buffer{}
	// write level and date
	_, _ = fmt.Fprintf(buf, "[%s] %s ", time.Now().Format(timeFormat), entry.level)

	// write default tags
	if f.keys.Len() != 0 {
		buf.Write(bytes.ReplaceAll(f.keys.Bytes(), []byte{0}, []byte{32}))
	}

	// write entry tags
	if entry.tags.Len() != 0 {
		buf.Write(bytes.ReplaceAll(entry.tags.Bytes(), []byte("\x00"), []byte(" ")))
	}
	// write entry message
	buf.Write(entry.message.Bytes())
	buf.WriteString("\n")

	return buf.Bytes()
}

type jsonFormatter struct {
	keys map[string]interface{}
}

// NewJSONFormatter will encode the entry along with the default provided keys as a JSON string
func NewJSONFormatter(defaultKeys map[string]interface{}) Formatter {
	return &jsonFormatter{keys: defaultKeys}
}

func (f *jsonFormatter) Format(entry *Entry) []byte {
	var msg = map[string]interface{}{
		"level": entry.level.String(),
		"date":  time.Now(),
	}

	// write default tags
	for k, v := range f.keys {
		msg[k] = v
	}
	// write entry tags
	if entry.tags.Len() != 0 {
		tags := strings.Split(entry.tags.String(), "\x00")
		for _, tag := range tags {
			parts := strings.Split(tag, "=")
			if len(parts) == 2 {
				msg[parts[0]] = strings.Trim(parts[1], `"`)
			}
		}
	}
	msg["message"] = entry.message.String()
	res, err := json.Marshal(msg)
	if err != nil {
		return nil
	}
	return res
}

// Sanitize log entry to prevent log forging
func sanitize(d []byte) []byte {
	d = bytes.Replace(d, []byte{13}, []byte{}, -1)
	// replace LF \n (mac) with space (unix)
	d = bytes.Replace(d, []byte{10}, []byte("\\n"), -1)
	return d
}
