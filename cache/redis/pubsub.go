package redis

import (
	"github.com/go-redis/redis"
)

type SubscriptionInfo struct {
	subscription *redis.PubSub
	channels     []string
	closeChannel chan bool
}

func (c *Cache) Subscribe(key string, redisMessageHandler func(msg *redis.Message), channels ...string) *redis.PubSub {
	channels = append(channels, key)
	subscription := c.redis.Subscribe(channels...)
	subscriptionInfo := SubscriptionInfo{
		subscription: subscription,
		channels:     channels,
		closeChannel: make(chan bool),
	}
	c.subscription[key] = &subscriptionInfo
	go c.redisClientListener(subscriptionInfo, redisMessageHandler)
	return subscription
}

func (c *Cache) Unsubscribe(key string) {
	subInfo := c.subscription[key]
	if subInfo == nil {
		return
	}
	close(subInfo.closeChannel)
	delete(c.subscription, key)
	_ = subInfo.subscription.Unsubscribe(subInfo.channels...)
}

func (c *Cache) Publish(channel string, message interface{}) *redis.IntCmd {
	return c.redis.Publish(channel, message)
}

func (c *Cache) Inc(key string) int {
	return int(c.redis.Incr(key).Val())
}

func (c *Cache) redisClientListener(subInfo SubscriptionInfo, redisClientHandler func(*redis.Message)) {
	for {
		select {
		case <-subInfo.closeChannel:
			// Channel and unsubscribe is done in Unsubscribe()
			return
		case message := <-subInfo.subscription.Channel():
			// process redis message for current subscription
			redisClientHandler(message)
		}
	}
}
