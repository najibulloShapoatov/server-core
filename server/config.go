package server

import (
	"github.com/najibulloShapoatov/server-core/server/session"
	"time"
)

// config contains all the configurations of the server and is suited with default values that
// require minimum to no intervention to start a secure web server
type Config struct {
	// HTTPS configuration
	HTTPS HTTPSConfig `config:"."`
	// Name of the server that will be used in response headers
	// Default value is ServerCore
	Name string `config:"platform.server.name" default:"ServerCore"`
	// Host is the domain name the web server will reply to.
	// Default value is localhost
	Host string `config:"platform.server.host" default:"localhost"`
	// Address on which the server will bind to.
	// Default value is 0.0.0.0 which will bind to all network interfaces
	Address string `config:"platform.server.address" default:"0.0.0.0"`
	// StaticPath where static assets are loaded from
	StaticPath string `config:"platform.server.staticPath" default:"/var/www"`
	// TraceHeader represents the name of the HTTP header used to add trace ids
	// Default value is X-Trace-Id
	TraceHeader string `config:"platform.server.security.tracing.header" default:"X-Trace-Id"`
	// Port to use to bind the HTTP server to.
	// Default value is 80
	Port int `config:"platform.server.port" default:"80"`
	// ReadTimeout for client requests. A timeout of 0 means no timeout.
	// Default value is 0
	ReadTimeout time.Duration `config:"platform.sever.readTimeout" default:"0"`
	// WriteTimeout for client responses. A timeout of 0 means no timeout.
	// Default value is 0
	WriteTimeout time.Duration `config:"platform.server.writeTimeout" default:"0"`
	// IdleTimeout for keep-alive connections. A timeout of 0 means no timeout.
	// Default value is 0
	IdleTimeout time.Duration `config:"platform.server.idleTimeout" default:"0"`
	// PostMaxSize is the maximum amount of payload a client can send.
	// Default value is 100MB
	PostMaxSize int `config:"platform.server.maxPostSize" default:"100MB"`
	// Session settings
	Session *session.Config `config:"."`
	// Cache settings
	Cache *CacheConfig `config:"."`
	// Security settings
	Security *SecurityConfig `config:"."`
	// UseCompression will enable a middleware to compress server responses
	// using one of the supported compression methods (GZip, Deflate, Br).
	// Default value is enabled
	UseCompression bool `config:"platform.server.gzip" default:"true"`
	// EnableTracing will enable the TraceHeader on all responses.
	// Default value is enabled
	EnableTracing bool `config:"platform.server.security.tracing.enabled" default:"yes"`
	// TraceRequired indicates that this service will be a middleware or backend service and
	// all requests coming to it should come from a service that emits the tracing header, thus making
	// it a requirement on all incoming requests.
	// Default value is disabled
	TraceRequired bool `config:"platform.server.security.tracing.required" default:"no"`
}

type HTTPSConfig struct {
	// Enable HTTPS
	Enabled bool `config:"platform.server.https.enabled" default:"false"`
	// Try to automatically generate or acquire SSL certificate from LetsEncrypt
	Auto bool `config:"platform.server.https.auto" default:"true"`
	// Method of auto generated certificate to use. Available options are lets-encrypt, self-signed or auto
	CertType string `config:"platform.server.https.autoType" default:"lets-encrypt"`
	// Default HTTPS port
	Port int `config:"platform.server.https.port" default:"433"`
	// Path to server certificate
	Cert string `config:"platform.server.https.cert"`
	// Path to server private key
	Key string `config:"platform.server.https.key"`
}

type SecurityConfig struct {
	// BruteForce protection configuration
	BruteForce *BruteForceConfig `config:"."`
	// CSRFTokenRequired indicates that POST, PUT, PATCH methods should have a CSRF token header
	// or they will be discarded.
	// Default value is disabled.
	CSRFTokenRequired bool `config:"platform.server.security.csrfRequired" default:"no"`
	// DNT flag indicates if the server should respect Do Not Track requests
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/DNT
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Tk
	// Default value is enabled
	DNT bool `config:"platform.server.security.dnt" default:"yes"`
	// PreventIFraming is set if you want to prevent web page to be embedded in a iframe by adding
	// the X-Frame-Options header which tells the browser the page should not be rendered in an iframe
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Frame-Options
	// Default value is enabled
	PreventIFraming bool `config:"platform.server.security.preventIFraming" default:"no"`
	// XSSProtection provides protection against cross-site scripting attack (XSS)
	// by setting the `X-XSS-Protection` header.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-XSS-Protection
	// Default value is enabled.
	XSSProtection bool `config:"platform.server.security.XSSProtection" default:"yes"`
	// HSTS (http strict transport security header) enables the Strict-Transport-Security header
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security
	// Default value is no.
	HSTS bool `config:"platform.server.security.hsts" default:"no"`
	// ContentTypeOptions enables the X-Content-Type-Options response HTTP header which is a marker used by
	// the server to indicate that the MIME types advertised in the Content-Type headers should not be
	// changed and be followed.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options
	// Default value is enabled
	ContentTypeOptions bool `config:"platform.server.security.contentTypeOptions" default:"true"`
	// HSTSDirectives enables the HSTS Strict-Transport-Security header directives
	// Default value is "max-age=63072000; includeSubDomains"
	HSTSDirectives string `config:"platform.server.security.HSTSDirectives" default:"max-age=63072000; includeSubDomains"`
	// CSP enables Content-Security-Policy header
	// Header can also be used for reporting by adding 'report-uri http://reportcollector.example.com/collector.cgi'
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy
	// Default value is "default-src 'self'"
	CSP string `config:"platform.server.security.csp" default:"default-src 'self'"`
	// Whitelist contains a comma separated list of IP's that can access the application.
	// The IP's can be defined as simple IP addresses, IP ranges, CIDR ranges and *
	Whitelist string `config:"platform.server.security.whitelist"`
	// Blacklist contains a comma separated list of IP's that cannot access the application.
	// The IP's can be defined as simple IP addresses, IP ranges, CIDR ranges and *
	Blacklist string `config:"platform.server.security.blacklist"`
	// Server security with url scan
	URLScanner bool `config:"platform.server.security.urlScanner" default:"false"`
	// IP ban time for url scan detection
	BanDuration time.Duration `config:"platform.server.security.banDuration" default:"5h"`
	// The Access-Control-Allow-Origin response header indicates whether the response can be
	// shared with requesting code from the given origin.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
	CORSOrigin string `config:"platform.server.security.cors.origin"`
	// The Access-Control-Allow-Headers response header is used in response to a preflight request
	// which includes the Access-Control-Request-Headers to indicate which HTTP headers can be used
	// during the actual request.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Headers
	CORSHeaders string `config:"platform.server.security.cors.headers"`
	// The Access-Control-Expose-Headers response header indicates which headers can be exposed
	// as part of the response by listing their names.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Expose-Headers
	CORSExposeHeaders string `config:"platform.server.security.cors.expose"`
	// The Access-Control-Allow-Methods response header specifies the method or methods allowed
	// when accessing the resource in response to a preflight request.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Methods
	CORSMethods string `config:"platform.server.security.cors.methods"`
	// The Access-Control-Request-Headers request header is used by browsers when issuing a preflight
	// request, to let the server know which HTTP headers the client might send when the actual request
	// is made (such as with setRequestHeader()). This browser side header will be answered by the complementary
	// server side header of Access-Control-Allow-Headers.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Request-Headers
	CORSRequest string `config:"platform.server.security.cors.request"`
}

type BruteForceConfig struct {
	// Enabled the brute force protection using a leaky bucket rate limiter
	Enabled bool `config:"platform.server.security.bruteForce.enabled" default:"false"`
	// Rate parameter for the leaky bucket
	Rate float64 `config:"platform.server.security.bruteForce.rate" default:"1"`
	// Capacity parameter for the leaky bucket
	Capacity int64 `config:"platform.server.security.bruteForce.capacity" default:"10"`
}

type CacheConfig struct {
	// Enable cache headers for static resources
	Enabled bool `config:"platform.server.cache.enable" default:"yes"`
	// TTL for static resources (js, css, images etc)
	TTL time.Duration `json:"platform.server.cache.ttl" default:"3d"`
}

func (cfg *Config) Validate() error {
	return nil
}
