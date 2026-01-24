package executors

import "sync"

type Closable struct {
	closables map[uint16]func() error
	lock      sync.Mutex
	counter   uint16
}

func (c *Closable) AddCloser(closer func() error) uint16 {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.counter++
	c.closables[c.counter] = closer
	return c.counter
}

func (c *Closable) RemoveCloser(id uint16) {
	c.lock.Lock()
	defer c.lock.Unlock()
}

func (c *Closable) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, closer := range c.closables {
		err := closer()
		if err != nil {
			return err
		}
	}
	c.closables = make(map[uint16]func() error)
	c.counter = 0
	return nil
}
