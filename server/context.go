package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/najibulloShapoatov/server-core/monitoring/log"
	"github.com/najibulloShapoatov/server-core/platform"
	"github.com/najibulloShapoatov/server-core/server/session"
	"github.com/najibulloShapoatov/server-core/utils/net"
	"mime/multipart"
	"net/http"
	"reflect"
)

// Context provided by the server to handle a request
type Context struct {
	// Request is a reference to the original HTTP request
	Request *http.Request
	// Response is a reference to the original HTTP response writer
	Response *Response
	// Server is reference to the server instance through which you can access the server
	// configuration
	Server *Server
	// Session
	Session *session.Session
	// Data is a map of values that can be stored for the duration of the request
	Data map[string]interface{}
	// DoNotTrack flag
	DNT bool
	// Consent given to track and use cookies
	Consent bool
	// private
	parsed bool
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request:  r,
		Response: newResponse(w),
		Data:     make(map[string]interface{}),
	}
}

// RemoteAddress returns the network address that sent the request
func (c *Context) RemoteAddr() string {
	return net.GetClientIP(c.Request)
}

// UserAgent returns the client's User-Agent, if sent in the request.
func (c *Context) UserAgent() string {
	return c.Request.UserAgent()
}

// Generic bad request from user (missing parameters, bad encoding, etc)
func (c *Context) BadRequest(err interface{}) {
	c.Response.WriteHeader(http.StatusBadRequest)
	log.Error(err)
	_, _ = c.Response.Write([]byte(c.error(err).Error()))
}

// User is not authenticated
func (c *Context) Unauthorized(err interface{}) {
	c.Response.WriteHeader(http.StatusUnauthorized)
	log.Error(err)
	_, _ = c.Response.Write([]byte(c.error(err).Error()))
}

// User is authenticated but doesn't have permission to do what it wants
func (c *Context) Forbidden(err interface{}) {
	c.Response.WriteHeader(http.StatusForbidden)
	log.Error(err)
	_, _ = c.Response.Write([]byte(c.error(err).Error()))
}

// User is not authenticated
func (c *Context) ServerError(err error) {
	c.Response.WriteHeader(http.StatusInternalServerError)
	log.Error(err)
	_, _ = c.Response.Write([]byte(c.error(err).Error()))
}

// Generic bad request from user (missing parameters, bad encoding, etc)
func (c *Context) ErrorBadRequest(e interface{}) (status int, err error) {
	log.Error(err)
	return http.StatusBadRequest, c.error(e)
}

// User is not authenticated
func (c *Context) ErrorUnauthorized(err error) (int, error) {
	log.Error(err)
	return http.StatusUnauthorized, c.error(err)
}

// User is not authenticated
func (c *Context) ErrorServerError(err error) (int, error) {
	log.Error(err)
	return http.StatusInternalServerError, c.error(err)
}

// User is authenticated but doesn't have permission to do what it wants
func (c *Context) ErrorForbidden(err error) (int, error) {
	log.Error(err)
	return http.StatusForbidden, c.error(err)
}

func (c *Context) error(e interface{}) (err error) {
	if er, ok := e.(error); ok {
		return er
	}
	if er, ok := e.(string); ok {
		return errors.New(er)
	}
	return errors.New("unknown error")
}

// Authenticated verifies the user if it is authenticated
func (c *Context) Authenticated() bool {
	return c.Session != nil
}

// AccountID returns the current logged in user id, 0 otherwise
func (c *Context) AccountID() db.ID {
	if !c.Authenticated() {
		return db.ID("")
	}
	return *c.Session.AccountID
}

// Can verifies if the user can perform the given operations
func (c *Context) Can(permission platform.Permission) bool {
	if !c.Authenticated() {
		return false
	}
	return c.Session.Permissions.Can(permission)
}

// Can verifies if the user can perform the given operations
func (c *Context) CanAny(permissions ...platform.Permission) bool {
	if !c.Authenticated() {
		return false
	}
	return c.Session.Permissions.CanAny(permissions...)
}

// Can verifies if the user can perform the given operations
func (c *Context) CanAll(permissions ...platform.Permission) bool {
	if !c.Authenticated() {
		return false
	}
	return c.Session.Permissions.CanAll(permissions...)
}

func (c *Context) OK() (int, error) {
	return http.StatusOK, nil
}

func (c *Context) Set(key string, val interface{}) {
	c.Data[key] = val
}

func (c *Context) FormFile(name string) (file multipart.File, header *multipart.FileHeader, err error) {
	if !c.parsed {
		err = c.Request.ParseMultipartForm(int64(c.Server.Config.PostMaxSize))
		if err != nil {
			return
		}
		c.parsed = true
	}
	file, header, err = c.Request.FormFile(name)
	return
}

func (c *Context) FormValue(name string) (string, error) {
	if !c.parsed {
		err := c.Request.ParseMultipartForm(int64(c.Server.Config.PostMaxSize))
		if err != nil {
			return "", err
		}
		c.parsed = true
	}
	return c.Request.FormValue(name), nil
}

func (c *Context) Get(key string, val interface{}) error {
	v, ok := c.Data[key]
	if !ok {
		return fmt.Errorf("not found")
	}
	if reflect.TypeOf(val) != reflect.TypeOf(v) {
		d, _ := json.Marshal(v)
		_ = json.Unmarshal(d, val)
		return nil
	}

	valPtr := reflect.ValueOf(val).Elem()
	if !valPtr.CanSet() {
		return fmt.Errorf("pointer required")
	}
	valPtr.Set(reflect.ValueOf(v).Elem())
	return nil
}
