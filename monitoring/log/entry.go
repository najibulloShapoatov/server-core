package log

import (
	"bytes"
	"fmt"
	"io"
	"time"
)

// Entry represents a user log entry with additional metadata
type Entry struct {
	level   Level
	time    time.Time
	tags    *bytes.Buffer
	message *bytes.Buffer
}

func (e *Entry) reset(lvl Level) {
	e.message.Reset()
	e.tags.Reset()
	e.time = time.Now()
	e.level = lvl
}

// Add custom tags to the log entry
func (e *Entry) Tag(key string, value interface{}) *Entry {
	encode(e.tags, key, value)
	return e
}

func encode(buf io.Writer, key string, value interface{}) {
	_, _ = fmt.Fprintf(buf, "%s=\"%v\"\x00", key, value)
}
