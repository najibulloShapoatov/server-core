package security

import (
	"math"
	"time"
)

type LeakyBucket struct {
	// Identifying key
	key string
	// Bucket capacity
	capacity int64
	// Amount of bucket leaks per second
	rate float64
	// Priority of the bucket
	p time.Time
	// Needed for heap methods
	index int
}

// NewLeakyBucket creates a new LeakyBucket with the give rate and capacity.
func NewLeakyBucket(rate float64, capacity int64) *LeakyBucket {
	return &LeakyBucket{
		rate:     rate,
		capacity: capacity,
		p:        time.Now(),
	}
}

func (b *LeakyBucket) Add(amount int64) int64 {
	count := b.count()
	if count >= b.capacity {
		// The bucket is full.
		return 0
	}

	if !time.Now().Before(b.p) {
		// The bucket needs to be reset.
		b.p = time.Now()
	}
	remaining := b.capacity - count
	if amount > remaining {
		amount = remaining
	}
	t := time.Duration(float64(time.Second) * (float64(amount) / b.rate))
	b.p = b.p.Add(t)

	return amount
}

func (b *LeakyBucket) count() int64 {
	if !time.Now().Before(b.p) {
		return 0
	}

	nsRemaining := float64(b.p.Sub(time.Now()))
	nsPerDrip := float64(time.Second) / b.rate
	count := int64(math.Ceil(nsRemaining / nsPerDrip))

	return count
}
