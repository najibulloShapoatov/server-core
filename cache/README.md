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
