package session

import (
	"errors"
	"time"
)

// Interface of a session store
type Store interface {
	// Store factory
	New() error
	// Session store name
	Type() string
	// Stores a session in the store
	Set(*Session) error
	// Retrieves a session from the store
	Get(Token) *Session
	// Removes a session from the store
	Del(Token) error
	// List all available sessions.
	// If argument is provided, it will return only sessions that match the account
	List(*string) []*Session
	// Removes all expired sessions
	GC()
	// Closes the store
	Close()
}

type Config struct {
	// Store indicates which data store to use to hold the sessions.
	// Available built in stores are "db", "redis", "mem"
	Store string `config:"platform.server.session.store" default:"redis"`
	// Enable the use of sessions
	Enabled bool `config:"platform.server.session.enabled" default:"yes"`
	// UseCookie will enable client sessions through cookies
	UseCookie bool `config:"platform.server.session.useCookie" default:"yes"`
	// CookieName for the session cookie
	CookieName string `config:"platform.server.session.cookieName" default:"_session"`
	// HeaderName of the header that will contain the session id
	HeaderName string `config:"platform.server.session.headerName" default:"X-Session-Id"`
	// TTL is the maximum inactivity of a session till it gets removed
	TTL time.Duration `config:"platform.server.session.ttl" default:"1h"`
}

const sessionPrefix = "session:"

var (
	config *Config
	stores = make(map[string]Store)
	store  Store
)

// Initialize a session store from the configuration
func Init(cfg *Config) error {
	config = cfg
	// look for the store by it's name
	s, ok := stores[config.Store]
	if !ok {
		return errors.New("invalid session store type " + config.Store)
	}
	// initialize store
	if err := s.New(); err != nil {
		return err
	}
	store = s
	return nil
}

// Register a new session store
func RegisterStore(store Store) {
	stores[store.Type()] = store
}
