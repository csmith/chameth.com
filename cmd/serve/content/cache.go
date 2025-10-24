package content

import (
	"sync/atomic"
	"time"
)

type Cache[T any] struct {
	value      atomic.Pointer[T]
	update     func() T
	maxAge     time.Duration
	lastUpdate time.Time
}

func NewCache[T any](maxAge time.Duration, update func() T) *Cache[T] {
	return &Cache[T]{
		maxAge: maxAge,
		update: update,
	}
}

func (c *Cache[T]) Get() *T {
	if c.lastUpdate.Add(c.maxAge).Before(time.Now()) {
		newValue := c.update()
		c.value.Store(&newValue)
		c.lastUpdate = time.Now()
	}
	return c.value.Load()
}

type KeyedCache[T any] struct {
	values map[string]*Cache[*T]
	update func(key string) *T
	maxAge time.Duration
}

func NewKeyedCache[T any](maxAge time.Duration, update func(key string) *T) *KeyedCache[T] {
	return &KeyedCache[T]{
		values: make(map[string]*Cache[*T]),
		update: update,
		maxAge: maxAge,
	}
}

func (c *KeyedCache[T]) Get(key string) *T {
	cache, ok := c.values[key]
	if !ok {
		cache = NewCache(c.maxAge, func() *T {
			return c.update(key)
		})
		c.values[key] = cache
	}
	return *cache.Get()
}
