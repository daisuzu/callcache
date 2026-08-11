// Package callcache provides a duplicate call suppression mechanism with cache.
package callcache

import (
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// Dispatcher handles each call.
type Dispatcher[V any] struct {
	calls          sync.Map
	expiration     int64
	updateInterval int64
}

// NewDispatcher creates a new Dispatcher of function or method calls.
// expiration is the period to keep the execution result. If updateInterval is
// greater than 0, the cache of the execution result will be updated in the
// background when the elapsed time from the previous execution is exceeded.
func NewDispatcher[V any](expiration, updateInterval time.Duration) *Dispatcher[V] {
	return &Dispatcher[V]{
		expiration:     expiration.Nanoseconds(),
		updateInterval: updateInterval.Nanoseconds(),
	}
}

// Do returns the execution result of fn associated with the given key. If there
// is a valid execution result, it is reused instead of the return value of fn.
func (d *Dispatcher[V]) Do(key string, fn func() (V, error)) (V, error) {
	if v, loaded := d.calls.Load(key); loaded {
		return v.(*call[V]).do(fn)
	}

	// Create a new entry atomically on cache miss.
	v, _ := d.calls.LoadOrStore(key, &call[V]{
		expiration:     d.expiration,
		updateInterval: d.updateInterval,
	})
	return v.(*call[V]).do(fn)
}

// Remove removes the execution result of the given key.
func (d *Dispatcher[V]) Remove(key string) {
	d.calls.Delete(key)
}

type call[V any] struct {
	mu             sync.RWMutex
	expiration     int64
	updateInterval int64
	group          singleflight.Group
	result         V
	lastUpdate     int64
}

func (c *call[V]) do(fn func() (V, error)) (V, error) {
	now := time.Now().UnixNano()

	c.mu.RLock()
	v := c.result
	t := now - c.lastUpdate
	c.mu.RUnlock()

	if t > c.expiration {
		return c.update(fn)
	}
	if c.updateInterval > 0 && t > c.updateInterval {
		go func() {
			_, _ = c.update(fn) // Ignore err from background updates.
		}()
	}
	return v, nil
}

func (c *call[V]) update(fn func() (V, error)) (V, error) {
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
	if val == nil {
		var zero V
		return zero, err
	}
	return val.(V), err
}
