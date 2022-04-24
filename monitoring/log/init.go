// Package log provides a logging package.
package log

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

var (
	// default log level
	logLevel Level
	// default log writer
	logWriter io.WriteCloser
	// default log formatter
	logFormatter Formatter
)

// Formatter is responsible to generate a byte array representing a log event in a specific format (json/txt/xml/etc)
type Formatter interface {
	Format(Entry *Entry) []byte
}

// Level of severity for a log entry
type Level byte

const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota + 1
	// FatalLevel level. Logs and then calls `os.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
)

func (l Level) String() string {
	switch l {
	case PanicLevel:
		return "PANIC"
	case FatalLevel:
		return "FATAL"
	case ErrorLevel:
		return "ERROR"
	case WarnLevel:
		return "WARN"
	case InfoLevel:
		return "INFO"
	case DebugLevel:
		return "DEBUG"
	}
	return "UNKNOWN"
}

type Config struct {
	Writer    string `config:"log.writer" default:"stdout"`
	Formatter string `config:"log.format" default:"text"`
	Level     string `config:"log.level" default:"warning"`
	MaxSize   int64  `config:"log.maxFileSize" default:"10000000"` // 10MB
}

func Setup(cfg Config) error {
	// parse debug level
	low := strings.ToLower(cfg.Level)
	switch low {
	default:
	case "off", "disabled", "none":
		SetLevel(0)
	case "panic":
		SetLevel(PanicLevel)
	case "fatal":
		SetLevel(FatalLevel)
	case "error":
		SetLevel(ErrorLevel)
	case "warning", "warn":
		SetLevel(WarnLevel)
	case "info":
		SetLevel(InfoLevel)
	case "debug":
		SetLevel(DebugLevel)
	}

	// parse writer
	low = strings.ToLower(cfg.Writer)
	switch {
	case low == "none", low == "disabled":
		SetWriter(NewNilWriter())
	case low == "stdout":
		SetWriter(NewDefaultWriter())
	case isFilePath(low):
		f, err := NewFileWriter(cfg.Writer, cfg.MaxSize)
		if err != nil {
			return err
		}
		SetWriter(f)
	default:
		return fmt.Errorf("invalid log writer: %s", cfg.Writer)
	}

	// parse formatter
	low = strings.ToLower(cfg.Formatter)
	switch {
	case low == "none", low == "disabled":
		SetFormatter(NewNilFormatter())
	case low == "text":
		SetFormatter(NewTextFormatter(nil))
	case low == "json":
		SetFormatter(NewJSONFormatter(nil))
	default:
		return fmt.Errorf("invalid log formatter: %s", cfg.Formatter)
	}

	return nil
}

func SetWriter(writer io.WriteCloser) {
	logWriter = writer
}

func SetFormatter(formatter Formatter) {
	logFormatter = formatter
}

func SetLevel(lvl Level) {
	logLevel = lvl
}

func init() {
	go processLogs()
}

// isFilePath check is a string is Win or Unix file path
func isFilePath(str string) bool {
	if match, _ := regexp.MatchString(`^[a-zA-Z]:\\(?:[^\\/:*?"<>|\r\n]+\\)*[^\\/:*?"<>|\r\n]*$`, str); match {
		// check windows path limit see:
		//  http://msdn.microsoft.com/en-us/library/aa365247(VS.85).aspx#maxpath
		return len(str[3:]) <= 32767
	} else if match, _ := regexp.MatchString(
		`^((?:/[a-zA-Z0-9.:]+(?:_[a-zA-Z0-9:.]+)*(?:-[:a-zA-Z0-9.]+)*)+/?)$`, str); match {
		return true
	}
	return false
}
