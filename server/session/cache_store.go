package session

import (
	"strings"

	"github.com/najibulloShapoatov/server-core/cache"
)

type cacheStore struct {
	store cache.Cache
}

func (c *cacheStore) Set(session *Session) error {
	ttl := config.TTL
	if session.Persistent {
		ttl = 0
	}
	return c.store.Set(string(session.ID), session, ttl)
}

func (c *cacheStore) Get(token Token) (session *Session) {
	_ = c.store.Get(string(token), &session)
	return
}

func (c *cacheStore) Del(token Token) error {
	return c.store.Del(string(token))
}

func (c *cacheStore) List(accountID *string) (res []*Session) {
	keys := c.store.Keys(sessionPrefix + "*")
	for _, k := range keys {
		if tmp := c.Get(Token(strings.TrimPrefix(k, sessionPrefix))); tmp != nil {
			if len(strings.TrimSpace(*accountID)) > 0 {
				if tmp.AccountID == accountID {
					res = append(res, tmp)
				}
			} else {
				res = append(res, tmp)
			}
		}
	}
	return
}

func (r *cacheStore) GC() {
}

func (c *cacheStore) Close() {
	c.store = nil
}
