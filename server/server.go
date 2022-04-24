package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/najibulloShapoatov/server-core/monitoring/log"
	"github.com/najibulloShapoatov/server-core/server/security"
	"github.com/najibulloShapoatov/server-core/settings"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Server struct {
	// config for server
	Config *Config
	// stop channel that notifies when server is shutting down
	stop chan bool
	// flag set when the server was started
	started bool
	// active connections
	active sync.WaitGroup
	// Certificates Manager
	certManager Manager
	// Server
	httpServer *http.Server
	//
	staticFiles map[string]struct{}
}

const (
	healthCheckPath = "/healthcheck"
	honeyPotPath    = "/honeypot"
	versionList     = "/versions"
)

func New(config *Config) (*Server, error) {
	if config == nil {
		s := settings.GetSettings()
		err := s.Unmarshal(&config)
		if err != nil {
			return nil, errors.New("invalid_server_config")
		}
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}

	svc := &Server{
		Config: config,
		stop:   make(chan bool),
	}
	http.DefaultServeMux = http.NewServeMux()
	http.HandleFunc("/", svc.handler)

	return svc, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler(w, r)
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	s.active.Add(1)
	defer s.active.Done()

	ctx := newContext(w, r)
	ctx.Server = s
	var h HandlerFunc

	if r.URL.Path == honeyPotPath {
		security.SetBannedIP(ctx.RemoteAddr())
		ctx.Response.WriteHeader(http.StatusNoContent)
		return
	}

	if r.URL.Path == healthCheckPath {
		ctx.Response.WriteHeader(http.StatusOK)
		return
	}

	if r.URL.Path == versionList {
		_ = s.listVersions(ctx)
		return
	}

	if _, ok := s.staticFiles[r.URL.Path]; ok {
		h = s.staticFileHandler
	}

	if h == nil {
		h = s.matchRoute(ctx)
		if h == nil {
			return
		}
	}

	for _, m := range middlewares {
		h = m(h)
	}

	err := h(ctx)
	if err != nil {
		if !ctx.Response.Committed {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			_, _ = fmt.Fprint(w, err.Error())
		}
	}
}

func (s *Server) matchRoute(ctx *Context) HandlerFunc {
	r := ctx.Request
	w := ctx.Response

	parts := strings.Split(ctx.Request.URL.Path, "/")
	if len(parts) < 4 {
		// @TODO: check some other static paths
		http.NotFound(w, r)
		return nil
	}

	serviceKey := parts[1] + "-" + parts[2]
	service, ok := routes[serviceKey]
	if !ok {
		http.NotFound(w, r)
		return nil
	}

	if r.Method == http.MethodOptions {
		s.optionsHandler(ctx, service, parts[3])
		return nil
	}

	methodKey := r.Method + parts[3]
	handler, ok := service[methodKey]
	if !ok {
		methodKey := r.Method + strings.ToLower(parts[3])
		handler, ok = service[methodKey]
		if !ok {
			http.NotFound(w, r)
			return nil
		}
	}
	return handler.Handler
}

func (s *Server) optionsHandler(ctx *Context, service map[string]handler, name string) {
	if _, ok := service[ctx.Request.Header.Get("Access-Control-Request-Method")+strings.ToLower(name)]; ok {
		_ = postSecurityMiddleware(func(context2 *Context) error { return nil })(ctx)
		return
	}

	http.NotFound(ctx.Response, ctx.Request)
}

func (s *Server) staticFileHandler(ctx *Context) error {
	f, _ := os.Open(filepath.Join(s.Config.StaticPath, ctx.Request.URL.Path))
	if f != nil {
		ext := filepath.Ext(ctx.Request.URL.Path)
		ctx.Response.Header().Set("Content-Type", mime.TypeByExtension(ext))

		_, _ = io.Copy(ctx.Response, f)
		_ = f.Close()
	}

	if s.Config.Security.URLScanner && strings.HasSuffix(ctx.Request.URL.Path, "robots.txt") {
		_, _ = fmt.Fprintf(ctx.Response, "\n\nUser-agent: *\nDisallow: %s\n", honeyPotPath)
	}

	if ctx.Response.Size == 0 {
		ctx.Response.WriteHeader(http.StatusNotFound)
	}

	return nil
}

func (s *Server) Start() error {
	var tlsConfig *tls.Config
	var addr string

	UseMiddleware(
		accessLogMiddleware,
		recoverMiddleware,
		monitoringMiddleware,
		traceMiddleware,
		preSecurityMiddleware,
		cacheMiddleware,
		postSecurityMiddleware,
		compressMiddleware,
	)

	s.readStaticFiles()

	if s.Config.Security.BruteForce.Enabled {
		_ = security.NewCollector(s.Config.Security.BruteForce.Rate, s.Config.Security.BruteForce.Capacity)
		UseMiddleware(bruteForceMiddleware)
	}

	if s.Config.HTTPS.Enabled {
		addr = fmt.Sprintf("%s:%d", s.Config.Address, s.Config.HTTPS.Port)

		// There are no certificates defined and the server cannot auto fetch any
		if s.Config.HTTPS.Cert != "" && s.Config.HTTPS.Key != "" && !s.Config.HTTPS.Auto {
			fmt.Printf("You must provide TLS certificate, enable it to auto generate or start the server in HTTP\n")
			os.Exit(1)
		}

		// test that key is valid
		ok, err := testKey(s.Config.HTTPS.Cert, s.Config.HTTPS.Key, s.Config.HTTPS.Auto)
		if err != nil && !s.Config.HTTPS.Auto {
			return err
		}

		// pick the certificate manager
		switch t := s.Config.HTTPS.CertType; {
		case ok:
			s.certManager = newExternalCertificate(s.Config.HTTPS.Cert, s.Config.HTTPS.Key)
		case t == "self-signed":
			s.certManager = newSelfSignManager()
			tlsConfig = s.certManager.TLSConfig()
		case t == "lets-encrypt" || t == "auto":
			s.certManager = newLetsEncryptManager([]string{s.Config.Host})
			tlsConfig = s.certManager.TLSConfig()
		default:
			return fmt.Errorf("invalid certificate provider: %s", s.Config.HTTPS.CertType)
		}

		if s.certManager == nil {
			return fmt.Errorf("no valid certificate provider")
		}

		if tlsConfig == nil {
			tlsConfig = &tls.Config{
				GetCertificate: s.certManager.GetCertificate,
			}
		}
	} else {
		addr = fmt.Sprintf("%s:%d", s.Config.Address, s.Config.Port)
	}

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s,
		TLSConfig:    tlsConfig,
		ReadTimeout:  s.Config.ReadTimeout,
		WriteTimeout: s.Config.WriteTimeout,
		IdleTimeout:  s.Config.IdleTimeout,
	}

	go func() {
		s.started = true
		var err error
		if s.Config.HTTPS.Enabled {
			crt, key := s.certManager.GetCertificateFiles()
			err = s.httpServer.ListenAndServeTLS(crt, key)
		} else {
			err = s.httpServer.ListenAndServe()
		}
		if err != nil {
			if err != http.ErrServerClosed {
				fmt.Printf("Error starting server: %s\n", err)
				os.Exit(1)
			}
		}
		s.started = false
	}()

	return nil
}

func (s *Server) Stop() error {
	err := s.httpServer.Shutdown(context.Background())
	if err != nil {
		log.Debugf("Shutting down server failed: %s", err)
	} else {
		log.Debug("Shutting down server successful")
	}

	stopped := make(chan bool, 1)
	go func() {
		done := make(chan bool)
		go func() {
			s.active.Wait()
			close(done)
		}()

		// wait for time out or notification that all active connections are done
		select {
		case <-time.After(time.Second * 10):
			log.Warn("Server killed (timed out)")
		case <-done:
			log.Info("Server stopped gracefully")
		}
		close(stopped)
	}()
	s.started = false
	<-stopped
	return err
}

func (s *Server) Restart() error {
	if s.started {
		if e := s.Stop(); e != nil {
			return e
		}
	}
	return s.Start()
}

func (s *Server) readStaticFiles() {
	var assets = make(map[string]struct{})
	files := getAllFiles(s.Config.StaticPath, true)
	for _, f := range files {
		assets[f] = struct{}{}
	}
	if _, ok := assets["/robots.txt"]; !ok {
		assets["/robots.txt"] = struct{}{}
	}
	s.staticFiles = assets
}

func getAllFiles(path string, removeBasePath bool) (res []string) {
	path = strings.ReplaceAll(path, "\\", "/")
	dir, err := os.Open(path)
	if err != nil {
		return
	}
	paths, e := dir.Readdir(-1)
	if e != nil {
		return
	}
	for _, f := range paths {
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}
		if f.IsDir() {
			for _, child := range getAllFiles(filepath.Join(path, f.Name()), false) {
				if removeBasePath {
					child = strings.Replace(child, path, "", 1)
				}
				res = append(res, child)
			}

		} else {
			fname := strings.ReplaceAll(filepath.Clean(filepath.Join(path, f.Name())), "\\", "/")
			if removeBasePath {
				fname = strings.Replace(fname, path, "", 1)
			}
			res = append(res, fname)
		}
	}
	return res
}

func (s *Server) listVersions(ctx *Context) error {
	var res = make(map[string]string)
	for name := range routes {
		parts := strings.Split(name, "-")
		res[parts[0]] = parts[1]
	}

	data, _ := json.MarshalIndent(res, "", "    ")

	ctx.Response.WriteHeader(http.StatusOK)
	_, _ = ctx.Response.Write(data)
	return nil
}
