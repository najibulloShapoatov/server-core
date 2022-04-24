package cluster

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	redisDriver "github.com/go-redis/redis"
	"github.com/najibulloShapoatov/server-core/cache"
	"github.com/najibulloShapoatov/server-core/cache/redis"
	"github.com/najibulloShapoatov/server-core/utils/net"
)

type MessageHandler func(*Message)

type nodeInfo struct {
	IP       string    `json:"ip"`
	LastSeen time.Time `json:"lastSeen"`
}

type sharedLock struct {
	Name   string
	Time   time.Time
	NodeId int
}

var mutex sync.Mutex

type Cluster struct {
	name   string
	nodeID int

	handler MessageHandler

	cache  *redis.Cache
	pubSub *redisDriver.PubSub

	key         string
	channelName string
	ip          string

	stop chan bool

	activeLocks []sharedLock
}

// Join a cluster by name
func Join(name string) (cluster *Cluster, err error) {
	// check that we can get a hold of a redis instance
	r := cache.GetCache(cache.Redis)
	if r == nil {
		return nil, errors.New("no redis connection")
	}
	red := r.(*redis.Cache)

	cluster = &Cluster{
		key:         fmt.Sprintf(redisClusterKey, name),
		channelName: fmt.Sprintf(redisChannelKey, name),
		name:        name,
		cache:       red,
		stop:        make(chan bool),
		ip:          net.GetLocalAddr(),
	}

	// obtain node id
	cluster.nodeID = red.HInc(cluster.key, redisIncrementProp)

	cluster.writeNodeInfo()

	cluster.pubSub = red.Subscribe(cluster.channelName, cluster.listener)
	msg, _ := cluster.wrapMessage(nodeJoined, cluster.nodeID)
	if err = cluster.cache.Publish(cluster.channelName, msg).Err(); err != nil {
		return nil, err
	}
	go cluster.ping()
	return
}

// Leave a cluster
func (c *Cluster) Leave() (err error) {
	// Announce that node is leaving
	msg, _ := c.wrapMessage(nodeLeave, c.nodeID)
	if err = c.cache.Publish(c.channelName, msg).Err(); err != nil {
		return err
	}
	// Remove itself from cluster table
	_ = c.cache.HDel(c.key, fmt.Sprintf("%d", c.nodeID))

	if c.pubSub != nil {
		_ = c.pubSub.Close()
	}
	close(c.stop)
	return
}

// Send a message to all other nodes
func (c *Cluster) Broadcast(payload interface{}) (err error) {
	msg, err := c.wrapMessage(nodeBroadcast, payload)
	if err != nil {
		return err
	}
	return c.cache.Publish(c.channelName, msg).Err()
}

func (c *Cluster) ID() int {
	return c.nodeID
}

func (c *Cluster) wrapMessage(typ messageType, payload interface{}) (*Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	msg := &Message{
		Type:   typ,
		NodeID: c.nodeID,
		Data:   data,
	}
	return msg, nil
}

func (c *Cluster) Lock(name string) error {
	mutex.Lock()
	defer mutex.Unlock()

	// check that a lock with the same name doesn't already exists
	for _, n := range c.activeLocks {
		if n.Name == name {
			return fmt.Errorf("a lock with this name already exists")
		}
	}

	var lock sharedLock
	if _ = c.cache.Get(fmt.Sprintf(redisLocksKey, c.name, name), &lock); !lock.Time.IsZero() {
		return fmt.Errorf("lock already acquired by %d", lock.NodeId)
	}

	lock = sharedLock{
		Name:   name,
		Time:   time.Now(),
		NodeId: c.nodeID,
	}
	c.activeLocks = append(c.activeLocks, lock)
	return c.cache.Set(fmt.Sprintf(redisLocksKey, c.name, name), lock, lockTTL)
}

func (c *Cluster) Unlock(name string) error {
	mutex.Lock()
	defer mutex.Unlock()

	var found bool
	// remove from active locks
	for idx, n := range c.activeLocks {
		if n.Name == name {
			found = true
			if idx == len(c.activeLocks)-1 {
				c.activeLocks = c.activeLocks[:idx]
			} else {
				c.activeLocks = append(c.activeLocks[:idx], c.activeLocks[idx+1:]...)
			}
			break
		}
	}
	if !found {
		return errors.New("no such lock")
	}
	return c.cache.Del(fmt.Sprintf(redisLocksKey, c.name, name))
}

// Register a callback to handle cluster messages
func (c *Cluster) OnMessage(callback MessageHandler) {
	c.handler = callback
	return
}

func (c *Cluster) listener(data *redisDriver.Message) {
	if data == nil {
		return
	}
	var msg Message
	err := json.Unmarshal([]byte(data.Payload), &msg)
	if err != nil {
		return
	}
	switch msg.Type {
	case nodeJoined:
		// id, _ := msg.Int()
	case nodeLeave:
		// id, _ := msg.Int()
	case nodeBroadcast:
		if c.handler != nil {
			c.handler(&msg)
		}
	}
}

func (c *Cluster) writeNodeInfo() {
	// write node info
	var n = nodeInfo{
		IP:       c.ip,
		LastSeen: time.Now(),
	}
	data, _ := json.Marshal(n)
	_ = c.cache.HSet(c.key, fmt.Sprintf("%d", c.nodeID), string(data))
}

func (c *Cluster) ping() {
	timer := time.NewTicker(pingTime)
	gcTimer := time.NewTicker(pingTime * 2)
	lockTimer := time.NewTicker(lockTTL)
	for {
		select {

		// maintain acquired locks by this node
		case <-lockTimer.C:
			for _, lock := range c.activeLocks {
				_ = c.cache.Set(fmt.Sprintf(redisLocksKey, c.name, lock.Name), lock, lockTTL)
			}

		// update cluster nodes list to set current time
		case <-timer.C:
			c.writeNodeInfo()

		// stop everything
		case <-c.stop:
			return

		// clear nodes that may be dead
		case <-gcTimer.C:
			if err := c.Lock("cluster-gc"); err != nil {
				continue
			}
			records, err := c.cache.HGetAll(c.key)
			if err != nil {
				continue
			}
			now := time.Now()
			for k, v := range records {
				if k == "nodeId" {
					continue
				}
				var node nodeInfo
				if err := json.Unmarshal([]byte(v), &node); err != nil {
					continue
				}
				if node.LastSeen.Add(pingTime).Before(now) {
					_ = c.cache.HDel(c.key, k)
				}
			}
			_ = c.Unlock("cluster-gc")
		}
	}
}
