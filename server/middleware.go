package server

import (
	"compress/flate"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/najibulloShapoatov/server-core/monitoring/log"
	"github.com/najibulloShapoatov/server-core/server/security"
	"github.com/najibulloShapoatov/server-core/server/session"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
)

const (
	headerAcceptEncoding        = "Accept-Encoding"
	headerContentEncoding       = "Content-Encoding"
	headerXFrameOptions         = "X-Frame-Options"
	headerXXSSProtection        = "X-Xss-Protection"
	headerVary                  = "Vary"
	headerCORSAllowCreadentials = "Access-Control-Allow-Credentials"
	headerCORSOrigin            = "Access-Control-Allow-Origin"
	headerCORSHeaders           = "Access-Control-Allow-Headers"
	headerCORSExpose            = "Access-Control-Expose-Headers"
	headerCORSMethods           = "Access-Control-Allow-Methods"
	headerCORSReqHeaders        = "Access-Control-Request-Headers"
	headerCORSReqMethod         = "Access-Control-Request-Method"
	headerHSTS                  = "Strict-Transport-Security"
	headerCSP                   = "Content-Security-Policy"
	headerXContentType          = "X-Content-Type-Options"
	headerXCSRF                 = "X-Csrf-Token"
	headerDNT                   = "DNT"
	headerXTrace                = "X-Trace-Id"
	headerTK                    = "Tk"
)

var middlewares = make([]Middleware, 0)

// Register a handler that will be called before the request handler is called
func UseMiddleware(middleware ...Middleware) {
	middlewares = append(middlewares, middleware...)
}

// Handler function used by middleware to chain call all of them
type HandlerFunc func(*Context) error

// Middleware function signature
type Middleware func(HandlerFunc) HandlerFunc

func recoverMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx *Context) error {
		defer func() {
			r := recover()
			if r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", err)
				}
				if err != nil {
					log.Debugf("[RECOVERED] %s", err)
				}
			}
		}()
		return next(ctx)
	}
}
func authMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx *Context) error {
		// add session to the context
		if ctx.Session == nil {
			var sessionID session.Token
			// search for a session id on the session cookie
			cookie, _ := ctx.Request.Cookie(ctx.Server.Config.Session.CookieName)
			// if session cookie is not present try on the session header
			if cookie == nil {
				sessionID = session.Token(ctx.Request.Header.Get("X-Session-Id"))
			} else {
				sessionID = session.Token(cookie.Value)
			}
			// if a valid session id is found restore the session
			if sessionID.Valid() {
				ctx.Session = session.Restore(sessionID)
			}
		}
		return next(ctx)
	}
}

// Validate all pre handler security checks (CSRF token, etc)
func preSecurityMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx *Context) error {
		addr := ctx.RemoteAddr()
		ua := ctx.UserAgent()
		urlPath := ctx.Request.URL.Path

		cfg := ctx.Server.Config.Security
		req := ctx.Request
		res := ctx.Response

		m := req.Method

		// check if IP is whitelisted/blacklisted
		wh := cfg.Whitelist
		if len(wh) != 0 && !security.CheckIP(addr, strings.Split(wh, ",")) {
			res.WriteHeader(http.StatusForbidden)
			return fmt.Errorf("%q is not whitelisted", addr)
		} else if bl := cfg.Blacklist; len(bl) != 0 && security.CheckIP(addr, strings.Split(bl, ",")) {
			res.WriteHeader(http.StatusForbidden)
			return fmt.Errorf("%q is blacklisted", addr)
		}

		// Check if request is url scanner
		if cfg.URLScanner && security.IsCrawler(urlPath, addr, ua, cfg.BanDuration) {
			ctx.Response.WriteHeader(http.StatusForbidden)
			return errors.New("your IP address was banned")
		}

		// Check and enforce CSRF token usage
		if cfg.CSRFTokenRequired &&
			ctx.Session != nil && (m == http.MethodPost || m == http.MethodPut || m == http.MethodPatch) {
			if token := req.Header.Get(headerXCSRF); token != "" && token != ctx.Session.CSRFToken {
				res.WriteHeader(http.StatusNotAcceptable)
				return errors.New("invalid CSRF token")
			} else if token == "" {
				res.WriteHeader(http.StatusNotAcceptable)
				return errors.New("missing CSRF token")
			}
		}

		if cfg.DNT && req.Header.Get(headerDNT) == "1" {
			ctx.DNT = true
			if ctx.Consent {
				res.Header().Set(headerTK, "C")
			} else {
				res.Header().Set(headerTK, "N")
			}
		}

		if cfg.PreventIFraming {
			res.Header().Set(headerXFrameOptions, "SAMEORIGIN")
		}

		if cfg.XSSProtection {
			res.Header().Set(headerXXSSProtection, "1; mode=block")
		}

		if cfg.HSTS && cfg.HSTSDirectives != "" {
			res.Header().Set(headerHSTS, cfg.HSTSDirectives)
		}

		if cspHeader := cfg.CSP; cspHeader != "" {
			res.Header().Set(headerCSP, cspHeader)
		}

		// add Content type marker header
		if cfg.ContentTypeOptions {
			res.Header().Set(headerXContentType, "nosniff")
		}

		return next(ctx)
	}
}

// Add required security headers
func postSecurityMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx *Context) error {
		cfg := ctx.Server.Config.Security
		res := ctx.Response

		if cfg.CORSOrigin != "" {
			res.Header().Set(headerCORSAllowCreadentials, "true")
			if cfg.CORSOrigin == "*" {
				res.Header().Set(headerCORSOrigin, cfg.CORSOrigin)
			} else if strings.Contains(cfg.CORSOrigin, ",") && strings.Contains(cfg.CORSOrigin, ctx.Request.Header.Get("Origin")) {
				res.Header().Set(headerCORSOrigin, ctx.Request.Header.Get("Origin"))
				res.Header().Set(headerVary, "Origin")
			} else {
				res.Header().Set(headerCORSOrigin, ctx.Request.Header.Get("Origin"))
			}
		}
		if ctx.Request.Method == http.MethodOptions && cfg.CORSExposeHeaders != "" {
			res.Header().Set(headerCORSExpose, cfg.CORSExposeHeaders)
		}
		if ctx.Request.Method == http.MethodOptions && cfg.CORSMethods != "" {
			if cfg.CORSMethods == "*" {
				res.Header().Set(headerCORSMethods, cfg.CORSMethods)
			} else if strings.Contains(strings.ToLower(cfg.CORSMethods), strings.ToLower(ctx.Request.Header.Get(headerCORSReqMethod))) {
				res.Header().Set(headerCORSMethods, ctx.Request.Header.Get(headerCORSReqMethod))
			}
		}
		if ctx.Request.Method == http.MethodOptions && cfg.CORSHeaders != "" {
			res.Header().Set(headerCORSHeaders, cfg.CORSHeaders)
		}
		return next(ctx)
	}
}

func cacheMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx *Context) error {
		return next(ctx)
	}
}

// log http call in apache access log format
func accessLogMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx *Context) error {
		req := ctx.Request
		h := ctx.RemoteAddr() // the IP address of the client (remote host)
		// u - the userID that requested the information
		u := "-"
		if ctx.Session != nil {
			if ctx.Session.AccountID != nil {
				u = fmt.Sprintf("%s", *ctx.Session.AccountID)
			}
		}
		t := time.Now().String()                               // the time that the request was received
		r := req.Method + " " + req.URL.Path + " " + req.Proto // the client request line, ex: "GET /image.png HTTP/1.0"

		err := next(ctx)

		s := ctx.Response.Status                                    // the response status code
		b := ctx.Response.Size                                      // the size of the object returned to the client
		ti := ctx.Request.Header.Get(ctx.Server.Config.TraceHeader) // ti - the request trace id

		log.Infof("%s %s %s %s %d %d %s", h, u, t, r, s, b, ti)
		return err
	}
}

// compressMiddleware compresses the http response if the compression is enabled
// and the client can handle it any of the supported compression methods (br, gzip, deflate)
func compressMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx *Context) error {
		var (
			wr  io.WriteCloser
			req = ctx.Request
			res = ctx.Response
		)
		disableCompression := req.Header.Get("X-No-Compression")
		if ctx.Server.Config.UseCompression && disableCompression == "" {
			switch {
			case strings.Contains(req.Header.Get(headerAcceptEncoding), "br"):
				wr = brotli.NewWriter(res.Writer)
				res.Header().Set(headerContentEncoding, "br")

			case strings.Contains(req.Header.Get(headerAcceptEncoding), "gzip"):
				wr = gzip.NewWriter(res.Writer)
				res.Header().Set(headerContentEncoding, "gzip")

			case strings.Contains(req.Header.Get(headerAcceptEncoding), "deflate"):
				wr, _ = flate.NewWriter(res.Writer, flate.DefaultCompression)
				res.Header().Set(headerContentEncoding, "deflate")
			}
		}
		if wr != nil {
			res.Compressor(wr)
		}

		// else use the default writer to send data uncompressed
		err := next(ctx)

		if wr != nil {
			if e := wr.Close(); e != nil {
				log.Errorf("Error closing compressed stream: %s", err)
			}
		}
		return err
	}
}

// traceMiddleware will append a tracing token for all
// requests and forward existing ones so the user of the platform
// can trace requests across micro-services.
func traceMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx *Context) error {
		if ctx.Server.Config.EnableTracing {
			headerName := ctx.Server.Config.TraceHeader
			if headerName == "" {
				headerName = headerXTrace
			}
			traceID := ctx.Request.Header.Get(headerName) // search for a trace id on the trace header

			// if trace header is required, requests without a trace header
			// will get a 400 response
			if ctx.Server.Config.TraceRequired && traceID == "" {
				ctx.Response.WriteHeader(http.StatusBadRequest)
				return errors.New("trace token required")
			}

			// if trace header is not required but it doesn't exit
			// create one and append it to the request and response
			if traceID == "" {
				b := make([]byte, 12)
				_, _ = rand.Read(b)
				traceID = hex.EncodeToString(b)
			}
			ctx.Request.Header.Set(headerName, traceID)
			ctx.Response.Header().Set(headerName, traceID)
		}
		return next(ctx)
	}
}

func bruteForceMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx *Context) error {
		collector := security.GetCollector()
		var res int64
		if ctx.Session == nil {
			res = collector.Add(ctx.RemoteAddr(), 1)
		} else {
			res = collector.Add(string(ctx.Session.ID), 1)
		}
		if res == 0 {
			ctx.Response.WriteHeader(http.StatusTooManyRequests)
			return errors.New("to many requests")
		}
		return next(ctx)
	}
}

// gauge: active requests beeing processed
// counter: error counts
// counter: status responses counts (400, 404, 500, 200, etc)
// histogram: response time
// histogram: response size
func monitoringMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx *Context) error {
		// increment request count
		res := next(ctx)
		// decrement request count
		return res
	}
}
