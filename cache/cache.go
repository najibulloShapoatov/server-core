package cache

import (
	"time"
)

type Config struct {
	Engine string `config:"platform.cache.engine" default:"bigCache"`
}

type Cache interface {
	// Type returns the type of the cache
	Type() string
	// Get retrieve value at key from cache
	Get(key string, value interface{}) (err error)
	// Has checks if key is available in cache
	Has(key string) (ok bool)
	// Set stores a key with a given life time. 0 for permanent
	Set(key string, value interface{}, ttl time.Duration) (err error)
	// Del removes a key by name
	Del(key string) (err error)
	// Keys list all available cache keys
	Keys(pattern string) (available []string)
	// Clear removes all keys
	Clear()
}

var defMgr = New()

// Register driver to manager instance
func Register(name string, driver Cache) *Manager {
	defMgr.DefaultUse(name)
	defMgr.Register(name, driver)
	return defMgr
}

// GetCache returns a cache instance by it's type
func GetCache(driverName string) Cache {
	return defMgr.Driver(driverName)
}

// DefaultUse set default driver name
func DefaultUse(driverName string) {
	defMgr.DefaultUse(driverName)
}

// Use driver object by name and set it as default driver.
func Use(driverName string) Cache {
	return defMgr.Use(driverName)
}

// Driver get a driver instance by name
func Driver(driverName string) Cache {
	return defMgr.Driver(driverName)
}

// DefManager get default cache manager instance
func DefManager() *Manager {
	return defMgr
}

// Default get default cache driver instance
func Default() Cache {
	return defMgr.Default()
}

// Has checks if key is available in cache
func Has(key string) (ok bool) {
	return defMgr.Default().Has(key)
}

// Get retrieve value at key from cache
func Get(key string, value interface{}) (err error) {
	return defMgr.Default().Get(key, value)
}

// Set stores a key with a given life time. 0 for permanent
func Set(key string, value interface{}, ttl time.Duration) (err error) {
	return defMgr.Default().Set(key, value, ttl)
}

// Del remove a key by name
func Del(key string) (err error) {
	return defMgr.Default().Del(key)
}

// Keys list all available cache keys
func Keys(pattern string) (available []string) {
	return defMgr.Default().Keys(pattern)
}
