package main

import (
	"sync"
)

// IdGen is a thread-safe structure for generating sequential ID numbers
type IdGen struct {
	current int
	*sync.Mutex
}

// NewIdGen creates and initializes an IdGen
func NewIdGen() *IdGen {
	return &IdGen{
		1,
		new(sync.Mutex),
	}
}

// GetNext gets the next sequential id value
func (i *IdGen) GetNext() int {
	i.Lock()
	defer i.Unlock()
	id := i.current
	i.current++
	return id
}
