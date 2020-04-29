// Package callcache provides a duplicate call suppression mechanism with cache.
package callcache

import (
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// Dispatcher handles each call.
type Dispatcher struct {
	mu             sync.Mutex
	expiration     int64
	updateInterval int64
	calls          map[string]*call
}

// NewDispatcher creates a new Dispatcher of function or method calls.
// expiration is the period to keep the execution result. If updateInterval is
// greater than 0, the cache of the execution result will be updated in the
// background when the elapsed time from the previous execution is exceeded.
func NewDispatcher(expiration, updateInterval time.Duration) *Dispatcher {
	return &Dispatcher{
		expiration:     expiration.Nanoseconds(),
		updateInterval: updateInterval.Nanoseconds(),
		calls:          make(map[string]*call),
	}
}

// Do returns the execution result of fn associated with the given key. If there
// is a valid execution result, it is reused instead of the return value of fn.
func (d *Dispatcher) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	d.mu.Lock()
	if d.calls[key] == nil {
		d.calls[key] = &call{expiration: d.expiration, updateInterval: d.updateInterval}
	}
	d.mu.Unlock()

	return d.calls[key].do(fn)
}

// Remove removes the execution result of the given key.
func (d *Dispatcher) Remove(key string) {
	d.mu.Lock()
	delete(d.calls, key)
	d.mu.Unlock()
}

type call struct {
	mu             sync.RWMutex
	expiration     int64
	updateInterval int64
	group          singleflight.Group
	result         interface{}
	lastUpdate     int64
}

func (c *call) do(fn func() (interface{}, error)) (interface{}, error) {
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

func (c *call) update(fn func() (interface{}, error)) (interface{}, error) {
	val, err, _ := c.group.Do("update", func() (interface{}, error) {
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
