package main

import (
	"sync"
)

// Counter is a threadsafe construct meant to be incremented across multiple goroutines
type Counter struct {
	count int
	*sync.RWMutex
}

// NewCounter creates and initializes a new counter
func NewCounter() *Counter {
	return &Counter{
		0,
		new(sync.RWMutex),
	}
}

// Incr increments the counter by 1
func (c *Counter) Incr() {
	c.Lock()
	defer c.Unlock()
	c.count++
}

// Count gets the current count
func (c *Counter) Count() int {
	c.RLock()
	defer c.RUnlock()
	ret := c.count
	return ret
}
