// Package callcache provides a duplicate call suppression mechanism with cache.
package callcache

import (
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// Dispatcher handles each call.
type Dispatcher struct {
	calls          sync.Map
	expiration     int64
	updateInterval int64
}

// NewDispatcher creates a new Dispatcher of function or method calls.
// expiration is the period to keep the execution result. If updateInterval is
// greater than 0, the cache of the execution result will be updated in the
// background when the elapsed time from the previous execution is exceeded.
func NewDispatcher(expiration, updateInterval time.Duration) *Dispatcher {
	return &Dispatcher{
		expiration:     expiration.Nanoseconds(),
		updateInterval: updateInterval.Nanoseconds(),
	}
}

// Do returns the execution result of fn associated with the given key. If there
// is a valid execution result, it is reused instead of the return value of fn.
func (d *Dispatcher) Do(key string, fn func() (any, error)) (any, error) {
	if v, loaded := d.calls.Load(key); loaded {
		return v.(*call).do(fn)
	}

	// Create a new entry atomically on cache miss.
	v, _ := d.calls.LoadOrStore(key, &call{
		expiration:     d.expiration,
		updateInterval: d.updateInterval,
	})
	return v.(*call).do(fn)
}

// Remove removes the execution result of the given key.
func (d *Dispatcher) Remove(key string) {
	d.calls.Delete(key)
}

type call struct {
	mu             sync.RWMutex
	expiration     int64
	updateInterval int64
	group          singleflight.Group
	result         any
	lastUpdate     int64
}

func (c *call) do(fn func() (any, error)) (any, error) {
	now := time.Now().UnixNano()

	c.mu.RLock()
	v := c.result
	t := now - c.lastUpdate
	c.mu.RUnlock()

	if t > c.expiration {
		return c.update(fn)
	}
	if c.updateInterval > 0 && t > c.updateInterval {
		go c.update(fn)
	}
	return v, nil
}

func (c *call) update(fn func() (any, error)) (any, error) {
	val, err, _ := c.group.Do("update", func() (any, error) {
		now := time.Now().UnixNano()
		if t := now - c.lastUpdate; t < c.expiration && (c.updateInterval == 0 || t < c.updateInterval) {
			// If the short term timing of c.group.Do does not match, use the previous result.
			return c.result, nil
		}
		v, err := fn()
		if err == nil {
			c.mu.Lock()
			c.result = v
			c.lastUpdate = now
			c.mu.Unlock()
		}
		return v, err
	})
	return val, err
}
