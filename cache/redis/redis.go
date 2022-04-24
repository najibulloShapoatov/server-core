package redis

import (
	"encoding/json"
	"fmt"
	"github.com/najibulloShapoatov/server-core/cache"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

type Config struct {
	Addr     string `config:"platform.cache.redis.addr" default:"localhost:6379"`
	Password string `config:"platform.cache.redis.password" default:""`
}

type Cache struct {
	config       *Config
	redis        *redis.Client
	closeChannel chan struct{}
	subscription map[string]*SubscriptionInfo
}

var (
	instance *Cache
	once     sync.Once
)

// New represents a new redis client
func New(config *Config) *Cache {
	if instance != nil {
		return instance
	}
	options, err := redis.ParseURL(config.Addr)
	if err != nil {
		options = &redis.Options{
			Addr:     config.Addr,
			Password: config.Password,
		}
	}

	if options.TLSConfig != nil {
		options.TLSConfig.InsecureSkipVerify = true
	}

	client := redis.NewClient(options)
	_, err = client.Ping().Result()
	if err != nil {
		_ = fmt.Errorf("redis connection error: %s", err)
	}
	once.Do(func() {
		instance = &Cache{
			config:       config,
			redis:        client,
			closeChannel: make(chan struct{}),
			subscription: make(map[string]*SubscriptionInfo),
		}
	})
	return instance
}

// Get retrieves value at key from cache
func (c *Cache) Get(key string, value interface{}) (err error) {
	var data []byte
	if err := c.redis.Get(key).Scan(&data); err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

// Has checks if key is available in cache
func (c *Cache) Has(key string) (ok bool) {
	item, err := c.redis.Keys(key).Result()
	if err != nil {
		return false
	}
	if len(item) != 0 {
		return true
	}
	return false
}

// Set stores a key with a given life time. 0 for permanent
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) (err error) {
	raw, _ := json.Marshal(value)
	_, err = c.redis.Set(key, raw, ttl).Result()
	if err != nil {
		_ = fmt.Errorf("write error: %s", err)
	}
	return err
}

// Del removes a value from redis
func (c *Cache) Del(key string) (err error) {
	_, err = c.redis.Del(key).Result()
	return err
}

// Keys list all available cache keys
func (c *Cache) Keys(pattern string) (available []string) {
	result, err := c.redis.Keys(pattern).Result()
	if err != nil {
		panic(err)
	}
	return result

}

// Type returns the type of the cache
func (c *Cache) Type() string {
	return cache.Redis
}

// Clear removes all keys and closes the client
func (c *Cache) Clear() {
	defer func() {
		_ = recover()
	}()
	c.redis.FlushAll()
	close(c.closeChannel)
}
