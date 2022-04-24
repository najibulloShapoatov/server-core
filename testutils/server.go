package testutils

import (
	"context"
	"io"
	"net/http"
	"time"
)

type MockServer struct {
	server   *http.Server
	addr     string
	calls    []*httpCall
	response []*mockedResponses
}

type mockedResponses struct {
	method   string
	path     string
	callback http.HandlerFunc
}

type httpCall struct {
	request *http.Request
}

func NewMockServer(addr string) (s *MockServer, err error) {
	svr := &MockServer{
		server: &http.Server{
			Addr: addr,
		},
		addr: addr,
	}
	svr.server.Handler = svr
	go func() {
		err = svr.server.ListenAndServe()
	}()
	<-time.After(time.Millisecond * 500)
	return svr, err
}

func (s *MockServer) Addr() string {
	return s.addr
}

func (s *MockServer) Stop() {
	if s.server != nil {
		_ = s.server.Shutdown(context.Background())
	}
}

func (s *MockServer) ResetCalls() {
	s.calls = make([]*httpCall, 0)
	s.response = make([]*mockedResponses, 0)
}

func (s *MockServer) LastRequest() *http.Request {
	if len(s.calls) != 0 {
		return s.calls[len(s.calls)-1].request
	}
	return nil
}

func (s *MockServer) MockResponse(method, path string, status int, body io.Reader, headers map[string]string) {
	s.response = append(s.response, &mockedResponses{
		path:   path,
		method: method,
		callback: func(writer http.ResponseWriter, request *http.Request) {
			for key, val := range headers {
				writer.Header().Set(key, val)
			}
			writer.WriteHeader(status)
			if body != nil {
				_, _ = io.Copy(writer, body)
			}
		},
	})
}

func (s *MockServer) MockCallback(path string, res http.HandlerFunc) {
	s.response = append(s.response, &mockedResponses{
		path:     path,
		callback: res,
	})
}

func (s *MockServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	path := request.URL.Path
	method := request.Method
	for idx, mock := range s.response {
		if mock.method != method || mock.path != path {
			continue
		}
		s.calls = append(s.calls, &httpCall{
			request: request,
		})
		mock.callback(writer, request)
		s.response = append(s.response[:idx], s.response[idx+1:]...)
		return
	}
}
