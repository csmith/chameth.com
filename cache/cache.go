package cache

import (
	"sync"
	"time"
)

// Cache lazily maintains a value of type T, refreshing it via update when it
// is older than maxAge. A failed refresh is not cached: the previous value is
// served and the refresh retried on the next call to Get. Until the first
// successful refresh, Get returns the zero value of T.
type Cache[T any] struct {
	mutex      sync.Mutex
	value      T
	loaded     bool
	update     func() (T, error)
	lastUpdate time.Time
	maxAge     time.Duration
}

func New[T any](maxAge time.Duration, update func() (T, error)) *Cache[T] {
	return &Cache[T]{
		maxAge: maxAge,
		update: update,
	}
}

func (c *Cache[T]) Get() T {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.loaded && time.Since(c.lastUpdate) < c.maxAge {
		return c.value
	}
	if value, err := c.update(); err == nil {
		c.value = value
		c.loaded = true
		c.lastUpdate = time.Now()
	}
	return c.value
}

// KeyedCache maintains one Cache per key, created on first use.
type KeyedCache[T any] struct {
	mutex  sync.Mutex
	values map[string]*Cache[T]

	update func(key string) (T, error)
	maxAge time.Duration
}

func NewKeyed[T any](maxAge time.Duration, update func(key string) (T, error)) *KeyedCache[T] {
	return &KeyedCache[T]{
		values: make(map[string]*Cache[T]),
		update: update,
		maxAge: maxAge,
	}
}

func (c *KeyedCache[T]) Get(key string) T {
	c.mutex.Lock()
	entry, ok := c.values[key]
	if !ok {
		entry = New(c.maxAge, func() (T, error) {
			return c.update(key)
		})
		c.values[key] = entry
	}
	c.mutex.Unlock()
	return entry.Get()
}
