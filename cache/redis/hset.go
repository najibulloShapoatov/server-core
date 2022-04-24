package redis

func (c *Cache) HInc(key, prop string) int {
	return int(c.redis.HIncrBy(key, prop, 1).Val())
}

func (c *Cache) HSet(key, prop string, val interface{}) error {
	return c.redis.HSet(key, prop, val).Err()
}

func (c *Cache) HDel(key, prop string) error {
	return c.redis.HDel(key, prop).Err()
}

func (c *Cache) HGet(key, prop string, ptr interface{}) error {
	return c.redis.HGet(key, prop).Scan(ptr)
}

func (c *Cache) HGetAll(key string) (map[string]string, error) {
	return c.redis.HGetAll(key).Result()
}
