package server

import (
	"io"
	"net/http"
)

type Response struct {
	Writer    http.ResponseWriter
	wr        io.Writer
	Status    int
	Size      int64
	Committed bool
}

func newResponse(w http.ResponseWriter) *Response {
	return &Response{
		Writer: w,
		wr:     w,
	}
}

// Header returns the response writer header
func (r *Response) Header() http.Header {
	return r.Writer.Header()
}

// WriteHeader writers the response status code the the client
func (r *Response) WriteHeader(code int) {
	if r.Committed {
		return
	}
	if code == 0 {
		code = http.StatusOK
	}
	r.Status = code
	r.Writer.WriteHeader(code)
	r.Committed = true
}

func (r *Response) Compressor(wr io.Writer) {
	if wr != nil {
		r.wr = wr
	} else {
		r.wr = r.Writer
	}
}

// Write sends the given byte array to the user through the ResponseWriter
func (r *Response) Write(b []byte) (n int, err error) {
	if !r.Committed {
		r.WriteHeader(http.StatusOK)
	}
	n, err = r.wr.Write(b)
	r.Size += int64(n)
	return
}

// Flush buffered data to the client.
func (r *Response) Flush() {
	r.wr.(http.Flusher).Flush()
}

// Custom writer used for compression
type compressedResponseWriter struct {
	io.Writer
	*Response
}

// Use the Writer part of compressedResponseWriter to write the http response
func (w compressedResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
