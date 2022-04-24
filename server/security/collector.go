package security

import (
	"container/heap"
	"sync"
	"time"
)

type bucketMap map[string]*LeakyBucket
type priorityQueue []*LeakyBucket

var collector *Collector

// A collector can keep track of multiple leaky buckets
type Collector struct {
	buckets  bucketMap
	heap     priorityQueue
	rate     float64
	capacity int64
	lock     sync.Mutex
	quit     chan bool
}

// Creates a new collector and check for empty buckets
func NewCollector(rate float64, capacity int64) *Collector {
	collector = &Collector{
		buckets:  make(bucketMap),
		heap:     make(priorityQueue, 0, 4096),
		rate:     rate,
		capacity: capacity,
		quit:     make(chan bool),
	}
	collector.periodicRemoveEmptyBuckets(time.Second)

	return collector
}

// Return the collector
func GetCollector() *Collector {
	if collector == nil {
		collector = NewCollector(100, 10000)
	}
	return collector
}

// Remove internal bucket associated with key
func (c *Collector) Remove(key string) {
	c.lock.Lock()
	if b, ok := c.buckets[key]; ok {
		delete(c.buckets, b.key)
		heap.Remove(&c.heap, b.index)
	}
	c.lock.Unlock()
}

func (c *Collector) Add(key string, amount int64) int64 {
	c.lock.Lock()
	b, ok := c.buckets[key]
	if !ok {
		// Create a new bucket.
		b = &LeakyBucket{
			key:      key,
			capacity: c.capacity,
			rate:     c.rate,
			p:        time.Now(),
		}
		c.heap.Push(b)
		c.buckets[key] = b
	}
	c.lock.Unlock()

	n := b.Add(amount)
	if n > 0 {
		heap.Fix(&c.heap, b.index)
	}
	return n
}

// Remove all empty buckets in the collector.
func (c *Collector) removeEmptyBuckets() {
	c.lock.Lock()
	for c.heap.Peak() != nil {
		b := c.heap.Peak()
		if time.Now().Before(b.p) {
			break
		}
		// The bucket should be empty.
		delete(c.buckets, b.key)
		heap.Remove(&c.heap, b.index)
	}
	c.lock.Unlock()
}

// Periodic checks and remove empty buckets
func (c *Collector) periodicRemoveEmptyBuckets(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C:
				c.removeEmptyBuckets()
			case <-c.quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (p priorityQueue) Len() int {
	return len(p)
}

func (p priorityQueue) Less(i, j int) bool {
	return p[i].p.Before(p[j].p)
}

func (p priorityQueue) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
	p[i].index = i
	p[j].index = j
}

func (p *priorityQueue) Push(x interface{}) {
	n := len(*p)
	b := x.(*LeakyBucket)
	b.index = n
	*p = append(*p, b)
}

func (p *priorityQueue) Pop() interface{} {
	old := *p
	n := len(old)
	*p = old[0 : n-1]
	return old[n-1]
}

func (p priorityQueue) Peak() *LeakyBucket {
	if len(p) <= 0 {
		return nil
	}
	return p[0]
}
