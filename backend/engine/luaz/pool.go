package luaz

import (
	"errors"
	"sync"
	"time"

	"github.com/k0kubun/pp"
)

// LuaStatePool implements a pool of lua.LState objects.
type LuaStatePool struct {
	m           sync.Mutex
	saved       []*LuaH
	maxSize     int
	minSize     int
	maxOnFlight int
	onFlight    int
	ttl         time.Duration
	timeout     map[*LuaH]time.Time

	// Optional initialization function for new states
	initFn func() (*LuaH, error)
}

// LuaStatePoolOptions defines the configuration for a new LuaStatePool.
type LuaStatePoolOptions struct {
	MaxSize     int
	MinSize     int
	MaxOnFlight int
	Ttl         time.Duration
	InitFn      func() (*LuaH, error)
}

// NewLuaStatePool creates a new LuaStatePool with the specified options.
func NewLuaStatePool(opts LuaStatePoolOptions) *LuaStatePool {
	pool := &LuaStatePool{
		maxSize:     opts.MaxSize,
		minSize:     opts.MinSize,
		maxOnFlight: opts.MaxOnFlight,
		onFlight:    0,
		ttl:         opts.Ttl,
		timeout:     make(map[*LuaH]time.Time),
		initFn:      opts.InitFn,
	}

	L, err := pool.initFn()
	if err == nil {
		pool.saved = append(pool.saved, L)
	}

	if opts.InitFn == nil {
		panic("initFn is nil")
	}

	return pool
}

func (p *LuaStatePool) Get() (*LuaH, error) {
	return p.get(0)
}

// Get returns a lua.LState from the pool or creates a new one if needed.
func (p *LuaStatePool) get(sleepCount int) (*LuaH, error) {
	p.m.Lock()

	n := len(p.saved)
	if n == 0 {
		p.m.Unlock()

		if p.onFlight >= p.maxOnFlight {
			if sleepCount > 3 {

				return nil, errors.New("max on flight reached")
			}

			pp.Println("@get/1", "max on flight reached", sleepCount)

			time.Sleep(time.Millisecond * 100)

			pp.Println("@get/2", "sleeping for 100ms")

			return p.get(sleepCount + 1)

		}

		lf, err := p.initFn()
		if err != nil {
			return nil, err
		}

		p.onFlight++
		return lf, nil
	}

	// Remove the last saved state from the pool
	L := p.saved[n-1]
	p.saved = p.saved[0 : n-1]
	delete(p.timeout, L)

	p.m.Unlock()

	pp.Println("@get/3", "got state from pool")

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
		p.onFlight--
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
		"sizes":         sizes,
		"max_size":      p.maxSize,
		"saved_size":    len(p.saved),
		"min_size":      p.minSize,
		"ttl":           p.ttl,
		"on_flight":     p.onFlight,
		"max_on_flight": p.maxOnFlight,
	}
}
