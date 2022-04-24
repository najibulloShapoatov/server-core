package session

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/najibulloShapoatov/server-core/platform"
	"github.com/najibulloShapoatov/server-core/utils/net"
)

// User session
type Session struct {
	// Unique session id
	ID Token `json:"sessionid" bson:"_id"`
	// Map of values stored on the session
	Data map[string]interface{} `json:"data" bson:"data"`
	// LastActivity Last time the session was used
	LastActivity time.Time `json:"lastactivity" bson:"lastActivity"`
	// ip of user
	IP string `json:"ip" bson:"ip"`
	// CSRF token
	CSRFToken string `json:"csrf" bson:"CSRF"`
	// User agent of the session holder
	UA string `json:"ua" bson:"UA"`
	// Date when the session was created
	Created time.Time `json:"created" bson:"created"`
	// Persistent flag will keep the sessions alive for a longer period of time
	Persistent bool `json:"persistent" bson:"persistent"`
	// Name oif the user
	Name *string `json:"name" bson:"name"`
	// Account id if linked with account
	AccountID *string `json:"accountid" bson:"accountId"`
	// Account id if linked with account
	ImpersonateAccountID *int `json:"impersonateAccountid" bson:"impersonateAccountid"`
	// Unique device id
	DeviceID *string `json:"deviceid" bson:"deviceId"`
	// Session can be locked for various reasons
	Locked bool `json:"locked" bson:"locked"`
	// List of user permissions
	Permissions *platform.Permissions `json:"permissions" bson:"permissions"`
}

// Creates a new session based on the user request
func New(r *http.Request) *Session {
	s := &Session{
		ID:           newToken(),
		Data:         make(map[string]interface{}),
		Created:      time.Now(),
		LastActivity: time.Now(),
		IP:           net.GetClientIP(r),
		CSRFToken:    string(newToken()),
		Permissions:  platform.NewPermissions(),
	}
	_ = store.Set(s)
	return s
}

func (s *Session) Set() {
	_ = store.Set(s)
}

func Restore(token Token) *Session {
	return store.Get(token)
}

func (s *Session) SetData(key string, val interface{}) {
	s.Data[key] = val
	_ = store.Set(s)
}

func (s *Session) GetData(key string) interface{} {
	return s.Data[key]
}

func (s *Session) Destroy() {
	_ = store.Del(s.ID)
}

func (s *Session) getData() string {
	if s.Data == nil {
		s.Data = make(map[string]interface{})
	}
	d, e := json.Marshal(s.Data)
	if e != nil {
		return "{}"
	}
	return string(d)
}
