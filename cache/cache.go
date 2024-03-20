package cache

import (
	awesomeerror "awesomeProject/error"
	"sync"
)

type Cacher interface {
	Get(key []byte) ([]byte, error)
	Set(key, value []byte) error
	Has(key []byte) bool
	Del(key []byte) ([]byte, error)
}

type Cache struct {
	lock sync.RWMutex
	data map[string][]byte
}

func NewCache() *Cache {
	return &Cache{
		lock: sync.RWMutex{},
		data: make(map[string][]byte),
	}
}

func (c *Cache) Get(key []byte) ([]byte, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	k := string(key)

	v, ok := c.data[k]
	if !ok {
		return nil, awesomeerror.New("key not found", k)
	}

	return v, nil
}

func (c *Cache) Set(key, value []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.data[string(key)] = value

	return nil
}

func (c *Cache) Has(key []byte) bool {
	c.lock.RLock()
	defer c.lock.Unlock()

	_, ok := c.data[string(key)]
	return ok
}

func (c *Cache) Del(key []byte) ([]byte, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	k := string(key)

	v, ok := c.data[k]
	if !ok {
		return nil, awesomeerror.New("key not found", k)
	}

	return v, nil
}
