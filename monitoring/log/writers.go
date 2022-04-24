package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type nilWriter struct{}

// NewNilWriter will return a writer that does nothing
func NewNilWriter() io.WriteCloser {
	return &nilWriter{}
}

func (w nilWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (w nilWriter) Close() error {
	return nil
}

type defaultWriter struct{}

// NewDefaultWriter will return a writer that writes to Stdout
func NewDefaultWriter() io.WriteCloser {
	return &defaultWriter{}
}

func (w defaultWriter) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

func (w defaultWriter) Close() error {
	return nil
}

type fileWriter struct {
	filename string
	file     *os.File
	size     int64
	maxSize  int64
}

// NewFileWriter will return a writer that writes to the given file up to the maxSize after which it will
// it will rotate the file. If maxSize is set to 0, then it will not rotate the file.
func NewFileWriter(filename string, maxSize int64) (io.WriteCloser, error) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return nil, err
	}
	w := &fileWriter{
		filename: filename,
		file:     f,
		maxSize:  maxSize,
	}
	// get file size since we need to know when to rotate the files
	if stat, er := f.Stat(); er == nil {
		w.size = stat.Size()
	}
	return w, nil
}

func (w *fileWriter) checkSize() error {
	if w.maxSize == 0 {
		return nil
	}
	if w.size >= w.maxSize {
		if er := w.file.Close(); er != nil {
			return er
		}
		ext := filepath.Ext(w.filename)
		name := strings.TrimSuffix(w.filename, ext)
		name = fmt.Sprintf("%s_%s_%d%s", name, time.Now().Format("02-Jan-2006"), time.Now().UnixNano(), ext)
		if err := os.Rename(w.filename, name); err != nil {
			return fmt.Errorf("error rotating log file: %s", err)
		}
		f, er := os.OpenFile(w.filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if er != nil {
			return er
		}
		w.file = f
		w.size = 0
	}
	return nil
}

func (w *fileWriter) Write(p []byte) (n int, err error) {
	if e := w.checkSize(); e != nil {
		return 0, e
	}
	n, err = w.file.Write(p)
	w.size += int64(n)
	return
}

func (w *fileWriter) Close() error {
	return w.file.Close()
}
