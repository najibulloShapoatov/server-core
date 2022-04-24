package session

import (
	"errors"
	"github.com/najibulloShapoatov/server-core/cache"
)

type redisStore struct {
	cacheStore
}

func (r *redisStore) New() error {
	if r != nil {
		return nil
	}
	redis := cache.GetCache(cache.Redis)
	if redis == nil {
		return errors.New("session store error - redis cache is not initialized")
	}

	store := &redisStore{
		cacheStore{store: redis},
	}
	*r = *store
	return nil
}

func (r *redisStore) Type() string {
	return "redis"
}


func init() {
	var store *redisStore
	stores["redis"] = store
}
