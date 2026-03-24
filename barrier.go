package kubecrtutils

import (
	"sync"
)

type Barrier struct {
	flag chan struct{}
	once sync.Once
}

func NewBarrier() *Barrier {
	return &Barrier{flag: make(chan struct{})}
}

func (c *Barrier) Done() {
	c.once.Do(func() { close(c.flag) })
}

func (c *Barrier) Wait() {
	<-c.flag
}
