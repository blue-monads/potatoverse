package luaz

import (
	"sync"
	"time"
)

// LuaStatePool implements a pool of lua.LState objects.
type LuaStatePool struct {
	m       sync.Mutex
	saved   []*LuaH
	maxSize int
	minSize int
	ttl     time.Duration
	timeout map[*LuaH]time.Time

	// Optional initialization function for new states
	initFn func() (*LuaH, error)
}

// LuaStatePoolOptions defines the configuration for a new LuaStatePool.
type LuaStatePoolOptions struct {
	MaxSize int
	MinSize int
	Ttl     time.Duration
	InitFn  func() (*LuaH, error)
}

// NewLuaStatePool creates a new LuaStatePool with the specified options.
func NewLuaStatePool(opts LuaStatePoolOptions) *LuaStatePool {
	pool := &LuaStatePool{
		maxSize: opts.MaxSize,
		minSize: opts.MinSize,
		ttl:     opts.Ttl,
		timeout: make(map[*LuaH]time.Time),
		initFn:  opts.InitFn,
	}

	L, err := pool.initFn()
	if err != nil {
		pool.saved = append(pool.saved, L)
	}

	return pool
}

// Get returns a lua.LState from the pool or creates a new one if needed.
func (p *LuaStatePool) Get() (*LuaH, error) {
	p.m.Lock()
	defer p.m.Unlock()

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

}

func (p *LuaStatePool) CleanupExpiredStates() {
	p.m.Lock()
	defer p.m.Unlock()

	now := time.Now()
	var unexpired []*LuaH

	for _, L := range p.saved {
		expiry, exists := p.timeout[L]

		// If the state has expired and we are above the minSize, close it
		if exists && now.After(expiry) && len(unexpired) >= p.minSize {
			L.Close()
			delete(p.timeout, L)
		} else {
			unexpired = append(unexpired, L)
		}
	}

	p.saved = unexpired
}

func (p *LuaStatePool) GetDebugData() map[string]any {

	p.m.Lock()
	defer p.m.Unlock()

	sizes := make([]int, 0, len(p.saved))

	for _, L := range p.saved {
		stackSize := L.L.GetTop()
		sizes = append(sizes, stackSize)
	}

	return map[string]any{
		"sizes":      sizes,
		"max_size":   p.maxSize,
		"saved_size": len(p.saved),
		"min_size":   p.minSize,
		"ttl":        p.ttl,
	}
}
