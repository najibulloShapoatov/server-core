package cache

import "time"

// Built in driver name
const (
	FreeCache = "freeCache"
	BigCache  = "bigCache"
	Redis     = "redis"
	MemCache  = "memCache"
)

// Manager definition with default driver name and drivers map
type Manager struct {
	defName string
	drivers map[string]Cache
}

// New creates a cache manager instance
func New() *Manager {
	return &Manager{
		drivers: make(map[string]Cache),
	}
}

// DefaultUse sets default driver name
func (m *Manager) DefaultUse(driverName string) {
	m.defName = driverName
}

// Register new driver object
func (m *Manager) Register(name string, driver Cache) *Manager {
	m.drivers[name] = driver
	return m
}

// Default returns the default driver instance
func (m *Manager) Default() Cache {
	return m.drivers[m.defName]
}

// Use driver object by name and set it as default driver.
func (m *Manager) Use(driverName string) Cache {
	m.DefaultUse(driverName)
	return m.Driver(driverName)
}

// Driver get a driver instance by name
func (m *Manager) Driver(driverName string) Cache {
	return m.drivers[driverName]
}

// DefName get default driver name
func (m *Manager) DefName() string {
	return m.defName
}

// Has checks if key is available in cache
func (m *Manager) Has(key string) (ok bool) {
	return m.Default().Has(key)
}

// Get retrieves value at key from cache
func (m *Manager) Get(key string, value interface{}) (err error) {
	return m.Default().Get(key, value)
}

// Set stores a key with a given life time. 0 for permanent
func (m *Manager) Set(key string, value interface{}, ttl time.Duration) (err error) {
	return m.Default().Set(key, value, ttl)
}

// Del remove a key by name
func (m *Manager) Del(key string) (err error) {
	return m.Default().Del(key)
}

// Keys lists all available cache keys
func (m *Manager) Keys(pattern string) (available []string) {
	return m.Default().Keys(pattern)
}
