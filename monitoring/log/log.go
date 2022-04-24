package log

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
)

var (
	entriesCount int64
	// entries pool
	entries = sync.Pool{
		New: func() interface{} {
			atomic.AddInt64(&entriesCount, 1)
			return &Entry{
				tags:    &bytes.Buffer{},
				message: &bytes.Buffer{},
			}
		}}
	// a cache of 1024 log entries that can wait to be written
	queue = make(chan *Entry, 1024)
	// closing means the server is going down and we want to flush the queue
	// so no new logs should be accepted
	closing bool
)

func Panic(args ...interface{}) {
	if logLevel < PanicLevel || closing {
		return
	}
	entry := getEntry(PanicLevel)
	_, _ = fmt.Fprint(entry.message, args...)
	queue <- entry
}

func Panicf(format string, args ...interface{}) {
	if logLevel < PanicLevel || closing {
		return
	}
	entry := getEntry(PanicLevel)
	_, _ = fmt.Fprintf(entry.message, format, args...)
	queue <- entry
}

func Fatal(args ...interface{}) {
	if logLevel < FatalLevel || closing {
		return
	}
	entry := getEntry(FatalLevel)
	_, _ = fmt.Fprint(entry.message, args...)
	printLog(entry)
}

func Fatalf(format string, args ...interface{}) {
	if logLevel < FatalLevel || closing {
		return
	}
	entry := getEntry(FatalLevel)
	_, _ = fmt.Fprintf(entry.message, format, args...)
	printLog(entry)
}

func Error(args ...interface{}) {
	if logLevel < ErrorLevel || closing {
		return
	}
	entry := getEntry(ErrorLevel)
	_, _ = fmt.Fprint(entry.message, args...)
	queue <- entry
}

func Errorf(format string, args ...interface{}) {
	if logLevel < ErrorLevel || closing {
		return
	}
	entry := getEntry(ErrorLevel)
	_, _ = fmt.Fprintf(entry.message, format, args...)
	queue <- entry
}

func Warn(args ...interface{}) {
	if logLevel < WarnLevel || closing {
		return
	}
	entry := getEntry(WarnLevel)
	_, _ = fmt.Fprint(entry.message, args...)
	queue <- entry
}

func Warnf(format string, args ...interface{}) {
	if logLevel < WarnLevel || closing {
		return
	}
	entry := getEntry(WarnLevel)
	_, _ = fmt.Fprintf(entry.message, format, args...)
	queue <- entry
}

func Info(args ...interface{}) {
	if logLevel < InfoLevel || closing {
		return
	}
	entry := getEntry(InfoLevel)
	_, _ = fmt.Fprint(entry.message, args...)
	queue <- entry
}

func Infof(format string, args ...interface{}) {
	if logLevel < InfoLevel || closing {
		return
	}
	entry := getEntry(InfoLevel)
	_, _ = fmt.Fprintf(entry.message, format, args...)
	queue <- entry
}

func Debug(args ...interface{}) {
	if logLevel < DebugLevel || closing {
		return
	}
	entry := getEntry(DebugLevel)
	debugAnnotations(entry)
	_, _ = fmt.Fprint(entry.message, args...)
	queue <- entry
}

func Debugf(format string, args ...interface{}) {
	if logLevel < DebugLevel || closing {
		return
	}
	entry := getEntry(DebugLevel)
	debugAnnotations(entry)
	_, _ = fmt.Fprintf(entry.message, format, args...)
	queue <- entry
}

func debugAnnotations(entry *Entry) {
	pc, file, line, ok := runtime.Caller(2)
	caller := runtime.FuncForPC(pc)
	if ok {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		b := make([]byte, 20)
		runtime.Stack(b, false)
		var goroutineNum int
		_, _ = fmt.Sscanf(string(b), "goroutine %d ", &goroutineNum)

		entry.Tag("source", fmt.Sprintf("%s:%d:%s[%d]", short, line, caller.Name(), goroutineNum))
	}
}

// This flag is so that the unit tests won't exist with os.Exit(1) on Fatal messages
var testMode = false

func processLogs() {
	for entry := range queue {
		printLog(entry)
	}
	if logWriter != nil {
		_ = logWriter.Close()
	}
}

func printLog(entry *Entry) {
	if entry.level == DebugLevel {
		debugAnnotations(entry)
	}
	if logWriter == nil {
		logWriter = NewDefaultWriter()
	}
	if logFormatter == nil {
		logFormatter = NewTextFormatter(nil)
	}
	_, err := logWriter.Write(logFormatter.Format(entry))
	if err != nil {
		panic(err)
	}
	if entry.level == FatalLevel && !testMode {
		os.Exit(1)
	}
	// put the entry back in the pool
	entries.Put(entry)
}

func getEntry(lvl Level) *Entry {
	entry := entries.Get().(*Entry)
	entry.reset(lvl)
	return entry
}
