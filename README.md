# Server
Server library provides a complete HTTP/HTTPS server that can be easily embedded in your application and
requires minimum to no configuration, being able to run securely with its default settings.

### see example [server-core-example](https://github.com/najibulloShapoatov/server-core-example) repository

https://pkg.go.dev/github.com/najibulloShapoatov/server-core

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

# Configuration library

Allows the application to load it's configuration from `.config` files or environment variables

## Install

To install the platform

```
$ go get github.com/najibulloShapoatov/server-core/settings
```

## Usage example

```go
import "github.com/najibulloShapoatov/server-core/settings"

type MySettings struct {
    Host    string        `config:"app.host" default:"localhost"`
    Port    int           `config:"app.port" default:"80"`
    Timeout time.Duration `config:"app.timeout" default:"3s"`
    Debug   bool          `config:"debug" default:"on"`
    
}

func main() {

   s := settings.GetSettings()
   err := s.Load(
         settings.NewFileLoader("file.conf", false),  // load config from this file
         settings.NewEnvLoader(false, "app")          // and also from environment variables
   )
   if err != nil {
       fmt.Prinln("failed to parse settings", err)
   }

   // unmarshall directly into a struct
   var mySettings MySettings
   if err := settings.Unmarshal(&mySettings); err != nil {
       fmt.Println("some error", err)
   }

   // retrieve a value by name
   isDebug, exists := settings.GetBool("debug")
}

```

## Configuration files

Configuration files are mostly `key=value` files but with few additions. For example, numbers are evaluated by the parser and booleans can be all truthy values besides true or false. Other files can be included using the `include` directive

```bash
# String values
key.name = "value" # this is inline comment
key.multiline = "multi \
line \
string"

# Number values
test.int.value = 5
test.float.value = 3.14
test.negative.value = -1.2
test.hex.number = 0x1234 # will parse to 4660
test.octal.number = 0o123 # will parse to 83
test.binary.number = 0b1010101 # will parse to 85
test.exponential.number = 1e3 # will parse to 1000
test.negative.exponential.number = 2e-2 # will parse to 0.02

# Boolean values
test.bool.value1 = yes       # or no
test.bool.value2 = on        # or off
test.bool.value3 = set       # or unset
test.bool.value4 = active    # or inactive
test.bool.value5 = enabled   # or disabled
test.bool.value6 = true      # or false
test.bool.value6 = 1         # or 0

# Duration
test.duration.value1 = "1h5m"
test.duration.value2 = "3s"

# Include other file
include "sub-config-file.conf"
```


# cache
Cache library with cache manager.

> All cache driver implemented the cache.Cache interface. So, You can add any custom driver.

**Supported Drivers:**

- `redis`  by `github.com/go-redis/redis`
- `memCached` by `github.com/bradfitz/gomemcache/memcache`
- `freeCache` by `github.com/coocood/freecache`
- `bigCache`  by `github.com/allegro/bigcache`

## Install

```bash
go get github.com/najibulloShapoatov/server-core/cache
```

You can also install individual drivers
```bash
go get github.com/najibulloShapoatov/server-core/cache/bigcache
```

## Cache Interface

All cache driver implemented the cache.Cache interface. So, You can add any custom driver.

```go
// Cache interface definition
type Cache interface {
	Type() string
	// Retrieve value at key from cache
	Get(key string, value interface{}) (err error)
	// Checks if key is available in cache
	Has(key string) (ok bool)
	// Stores a key with a given life time. 0 for permanent
	Set(key string, value interface{}, ttl time.Duration) (err error)
	// Remove a key by name
	Del(key string) (err error)
	// List all available cache keys
	Keys(pattern string) (available []string)
	// Removes all keys
	Clear()
}
```

## Usage example

```go
package main

import (
	"fmt"
	"github.com/allegro/bigcache"
	bigcachecore "github.com/najibulloShapoatov/server-core/cache/bigcache"
	"github.com/najibulloShapoatov/server-core/cache/freecache"
	"github.com/najibulloShapoatov/server-core/cache/memcache"
	"github.com/najibulloShapoatov/server-core/core/cache"
	"github.com/najibulloShapoatov/server-core/core/cache/redis"
	"time"
)

// change package main
func main() {
	redisCfg := redis.Config{Addr: "0.0.0.0:6379", Password: ""}
	redisClient := redis.New(&redisCfg)
	memCacheCfg := memcache.Config{Addr: "127.0.0.1:11211"}
	memCacheClient := memcache.New(&memCacheCfg)
	freeCacheCfg := freecache.Config{}
	freeCacheClient := freecache.New(&freeCacheCfg)
	bigCacheCfg := bigcache.DefaultConfig(5 * time.Second)
	bigCacheClient := bigcachecore.New(&bigCacheCfg)

	// register one(or some) cache driver

	cache.Register(cache.Redis, redisClient)
	cache.Register(cache.MemCache, memCacheClient)
	cache.Register(cache.FreeCache, freeCacheClient)
	cache.Register(cache.BigCache, bigCacheClient)
	cache.DefaultUse(cache.Redis)
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("testm-%d", i)
		value := []byte(fmt.Sprintf("value-%d", i))

		err := cache.Set(key, value, 5)
		fmt.Println(err)
	}
	var rez []byte
	_ = cache.Get("test-77", &rez)
	fmt.Println(rez)
	cache.DefaultUse(cache.BigCache)
	key := fmt.Sprintf("testm-%d", 99)
	value := []byte(fmt.Sprintf("value-%d", 48777772788))
	_ = cache.Set(key, value, 60)
	_ = cache.Get("testm-99", &rez)
	fmt.Println("BC:", rez)
	fc := cache.Driver(cache.FreeCache)
	keyb := fmt.Sprintf("testdriver-%d", 611011010)
	valueb := []byte(fmt.Sprintf("valuedriver-%d", 611011010))
	err := fc.Set(keyb, valueb, 0)
	fmt.Println("ERR:", err)
	var fcr []byte
	_ = fc.Get("test-611011010", &fcr)
	fmt.Println("FC:", &fcr)

}
```
# Cluster

Allows communication and synchronization between cluster nodes

## Install

To install the library

```
$ go get github.com/najibulloShapoatov/server-core/cluster
```

## Usage example

```go

c, err := cluster.Join("cluster-name")

// Listen for cluster messages
c.OnMessage(func(msg *cluster.Message) {
   // handle message from other nodes
})

// send a message to all nodes in the cluster
c.Broadcast(message)

// Obtain a mutex lock in the cluster
err := c.Lock("lock-name")
if err != nil {
    // Mutex is already lock by another node
}

// ...
// Perform action
// ...

// release lock 
c.Unlock("lock-name")

// Leave the cluster
c.Leave()

```


# Scheduler library

Allows the application to run background jobs.

## Install

To install the library

```
$ go get github.com/najibulloShapoatov/server-core/scheduler
``` 

## Usage example

```go
task := Task{
	Name:     "task 1",
	Spec:     "* * * * * *",
	MaxRetry: 3,
	Job: func() (err error) {
	    // Do something
	    return err
	},
}
err := RegisterJob(&task)

if err != nil {
    UnregisterJob(&task)
}
```

###Cron Expression Format

```
Field name   | Mandatory? | Allowed values  | Allowed special characters
----------   | ---------- | --------------  | --------------------------
Seconds      | Yes        | 0-59            | * / , -
Minutes      | Yes        | 0-59            | * / , -
Hours        | Yes        | 0-23            | * / , -
Day of month | Yes        | 1-31            | * / , - ?
Month        | Yes        | 1-12 or JAN-DEC | * / , -
Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?

* : every
, : multiple values
- : ranges
? : can be used for leaving Day Of month or Day of week blank
```

###Predefined schedules
```
Entry                  | Description                                | Equivalent To
-----                  | -----------                                | -------------
@yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 0 1 1 *
@monthly               | Run once a month, midnight, first of month | 0 0 0 1 * *
@weekly                | Run once a week, midnight between Sat/Sun  | 0 0 0 * * 0
@daily (or @midnight)  | Run once a day, midnight                   | 0 0 0 * * *
@hourly                | Run once an hour, beginning of hour        | 0 0 * * * *
@every <duration>      | Run every time specified, eg: @every 1h30m |
```



