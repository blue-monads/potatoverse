package luaz

import (
	"errors"
	"sync"
	"time"
)

// LuaStatePool implements a pool of lua.LState objects.
type LuaStatePool struct {
	m       sync.Mutex
	saved   []*LuaH
	maxSize int
	ttl     time.Duration
	timeout map[*LuaH]time.Time
	closed  bool

	// Optional initialization function for new states
	initFn func() (*LuaH, error)
}

type LuaStatePoolOptions struct {
	MaxSize int
	Ttl     time.Duration
	InitFn  func() (*LuaH, error)
}

// NewLuaStatePool creates a new LuaStatePool with the specified maximum size.
func NewLuaStatePool(opts LuaStatePoolOptions) *LuaStatePool {
	pool := &LuaStatePool{
		maxSize: opts.MaxSize,
		ttl:     opts.Ttl,
		timeout: make(map[*LuaH]time.Time),
		initFn:  opts.InitFn,
	}

	// Start a cleanup goroutine if TTL is enabled
	if opts.Ttl > 0 {
		go pool.periodicCleanup()
	}

	return pool
}

// Get returns a lua.LState from the pool or creates a new one if needed.
func (p *LuaStatePool) Get() (*LuaH, error) {
	p.m.Lock()
	defer p.m.Unlock()

	if p.closed {
		return nil, errors.New("pool is closed")
	}

	n := len(p.saved)
	if n == 0 {
		if p.initFn != nil {
			return p.initFn()
		}
		return nil, nil
	}

	// Remove the last saved state from the pool
	L := p.saved[n-1]
	p.saved = p.saved[0 : n-1]
	delete(p.timeout, L)

	return L, nil
}

// Put returns a lua.LState to the pool.
func (p *LuaStatePool) Put(L *LuaH) error {
	p.m.Lock()
	defer p.m.Unlock()

	if p.closed {
		L.Close()
		return errors.New("pool is closed")
	}

	// Reset the Lua state to make it clean for the next use
	L.L.SetTop(0) // Clear the stack

	// Only store up to maxSize states
	if len(p.saved) >= p.maxSize {
		// Discard the Lua state by closing it
		L.Close()
		return nil
	}

	p.saved = append(p.saved, L)

	// Set timeout for this state if TTL is enabled
	if p.ttl > 0 {
		p.timeout[L] = time.Now().Add(p.ttl)
	}

	return nil
}

// Close closes all Lua states in the pool.
func (p *LuaStatePool) Close() {
	p.m.Lock()
	defer p.m.Unlock()

	for _, L := range p.saved {
		L.Close()
	}

	p.saved = nil
	p.timeout = make(map[*LuaH]time.Time)
	p.closed = true
}

// periodicCleanup checks for expired Lua states and removes them.
func (p *LuaStatePool) periodicCleanup() {
	ticker := time.NewTicker(p.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		p.cleanupExpiredStates()
	}
}

// cleanupExpiredStates removes expired Lua states from the pool.
func (p *LuaStatePool) cleanupExpiredStates() {
	p.m.Lock()
	defer p.m.Unlock()

	if p.closed {
		return
	}

	now := time.Now()
	var unexpired []*LuaH

	for _, L := range p.saved {
		expiry, exists := p.timeout[L]

		// If the state has expired, close it
		if exists && now.After(expiry) {
			L.Close()
			delete(p.timeout, L)
		} else {
			unexpired = append(unexpired, L)
		}
	}

	p.saved = unexpired
}
