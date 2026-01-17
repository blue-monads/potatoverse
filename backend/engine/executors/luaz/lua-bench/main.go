package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/flosch/go-humanize"
	lua "github.com/yuin/gopher-lua"
)

const (
	numStates      = 1000
	intervalMillis = 200
)

func main() {
	startTime := time.Now()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	mem_before := m.Alloc
	fmt.Printf("Initial memory allocated: %d bytes\n\n", mem_before)

	states := make([]*lua.LState, numStates)
	var wg sync.WaitGroup
	done := make(chan struct{})

	// Start a goroutine to monitor memory usage
	go func() {
		ticker := time.NewTicker(intervalMillis * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				runtime.GC() // Force garbage collection to get a more accurate number
				runtime.ReadMemStats(&m)
				humanReadableMemory := humanize.Bytes(uint64(m.Alloc))
				fmt.Printf("Time: %v | Memory allocated: %s\n", time.Now().Format("15:04:05.000"), humanReadableMemory)
			case <-done:
				return
			}
		}
	}()

	// Create and run the Lua states concurrently
	wg.Add(numStates)
	for i := 0; i < numStates; i++ {
		go func(id int) {
			defer wg.Done()
			L := lua.NewState()
			defer L.Close()
			states[id] = L // Store the state reference to keep it alive

			// Run the benchmark script
			if err := L.DoFile("benchmark_script.lua"); err != nil {
				fmt.Printf("Error running script on state %d: %v\n", id, err)
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Signal the monitoring goroutine to stop
	close(done)

	// Final memory stats after all states have been used
	runtime.GC()
	runtime.ReadMemStats(&m)
	final_mem := m.Alloc
	humanReadableMemory := humanize.Bytes(uint64(final_mem))
	fmt.Printf("\nFinal memory allocated after all states are used: %s\n", humanReadableMemory)
	fmt.Printf("Time taken: %s\n", time.Since(startTime))

	// Wait for the memory monitor to exit
	time.Sleep(500 * time.Millisecond) // Give the ticker a moment to stop

}
