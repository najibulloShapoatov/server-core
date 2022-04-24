# Server
Server library provides a complete HTTP/HTTPS server that can be easily embedded in your application and
requires minimum to no configuration, being able to run securely with its default settings.

#### Features
 - Automatic routing
 - Auto acquire HTTPS certificates through Let's Encrypt
 - Session management with support for different stores (DB, Redis, In-Memory built in stores)
 - Support for various input/output encodings (JSON, XML, Binary, gRPC built in)
 - Support for various compression algorithms (GZip, Deflate, Brotli built in)
 - Access logs (Apache format compatible)
 - Security enhancements (XSS, CSRF, DNT, HSTS, CSP, BruteForce, Rate Limiter, Whitelist/Blacklist built in)
 - Request tracing
 - Server statistics
 - Internationalization and GeoIP detection support 
  
### Install

```bash
$ go get github.com/najibulloShapoatov/server-core/server
```

### Usage example(s)


##### Start a server that listens on HTTP(80)
```go
import (
   "github.com/najibulloShapoatov/server-core/server"
)

func main() {
    svc := server.New(nil)
    svc.Start()
}
```
 
##### Register custom stores and engines 
```go
// register 2 new middleware functions
server.UseMiddleware(m1, m2) 

// register decoder/encoder for some custom content type
server.RegisterDecoder("text/yaml", myYAMLDecoder)
server.RegisterEncoder("text/yaml", myYAMLEncoder)

// register http endpoints for all your module handlers
server.RegisterRoute(myService)


```
