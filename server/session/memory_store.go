package session

import (
	"errors"
	"github.com/najibulloShapoatov/server-core/cache"
)

type memoryStore struct {
	cacheStore
}

func (m *memoryStore) New() error {
	if m != nil {
		return nil
	}
	mem := cache.GetCache(cache.BigCache)
	if mem == nil {
		return errors.New("session store error - local cache is not initialized")
	}
	store := &memoryStore{
		cacheStore{store: mem},
	}
	*m = *store
	return nil
}

func (m *memoryStore) Type() string {
	return "mem"
}

func init() {
	var store *memoryStore
	stores["mem"] = store
}
