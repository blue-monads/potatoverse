package main

import (
	"flag"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/flosch/go-humanize"
	lua "github.com/yuin/gopher-lua"
)

const (
	luaBindableTypeName = "potato.module"
)

var (
	numStates      = flag.Int("vms", 100, "Number of Lua VMs to create")
	intervalMillis = flag.Int("interval", 200, "Memory monitoring interval in milliseconds")
	bindingType    = flag.String("type", "metatable", "Binding type: 'metatable' (with metatable) or 'direct' (obj.xyz without metatable)")
)

// Dummy data structure to hold in userdata
type TestModule struct {
	value int
}

// Create 12 methods similar to bind_kv.go pattern
func createMethods() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"method1": func(L *lua.LState) int {
			mod := checkModule(L)
			L.Push(lua.LNumber(mod.value + 1))
			return 1
		},
		"method2": func(L *lua.LState) int {
			mod := checkModule(L)
			L.Push(lua.LNumber(mod.value + 2))
			return 1
		},
		"method3": func(L *lua.LState) int {
			mod := checkModule(L)
			L.Push(lua.LNumber(mod.value + 3))
			return 1
		},
		"method4": func(L *lua.LState) int {
			mod := checkModule(L)
			L.Push(lua.LNumber(mod.value + 4))
			return 1
		},
		"method5": func(L *lua.LState) int {
			mod := checkModule(L)
			L.Push(lua.LNumber(mod.value + 5))
			return 1
		},
		"method6": func(L *lua.LState) int {
			mod := checkModule(L)
			L.Push(lua.LNumber(mod.value + 6))
			return 1
		},
		"method7": func(L *lua.LState) int {
			mod := checkModule(L)
			L.Push(lua.LNumber(mod.value + 7))
			return 1
		},
		"method8": func(L *lua.LState) int {
			mod := checkModule(L)
			L.Push(lua.LNumber(mod.value + 8))
			return 1
		},
		"method9": func(L *lua.LState) int {
			mod := checkModule(L)
			L.Push(lua.LNumber(mod.value + 9))
			return 1
		},
		"method10": func(L *lua.LState) int {
			mod := checkModule(L)
			L.Push(lua.LNumber(mod.value + 10))
			return 1
		},
		"method11": func(L *lua.LState) int {
			mod := checkModule(L)
			L.Push(lua.LNumber(mod.value + 11))
			return 1
		},
		"method12": func(L *lua.LState) int {
			mod := checkModule(L)
			L.Push(lua.LNumber(mod.value + 12))
			return 1
		},
	}
}

func checkModule(L *lua.LState) *TestModule {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*TestModule); ok {
		return v
	}
	L.ArgError(1, luaBindableTypeName+" expected")
	return nil
}

func registerBindableType(L *lua.LState, methods map[string]lua.LGFunction) {
	mt := L.NewTypeMetatable(luaBindableTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), methods))
}

func newBindableModule(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = &TestModule{value: 100}
	L.SetMetatable(ud, L.GetTypeMetatable(luaBindableTypeName))
	return ud
}

func newDirectModule(L *lua.LState) *lua.LTable {
	mod := &TestModule{value: 100}
	obj := L.NewTable()

	// Add methods directly to the table (without metatable)
	obj.RawSetString("method1", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(mod.value + 1))
		return 1
	}))
	obj.RawSetString("method2", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(mod.value + 2))
		return 1
	}))
	obj.RawSetString("method3", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(mod.value + 3))
		return 1
	}))
	obj.RawSetString("method4", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(mod.value + 4))
		return 1
	}))
	obj.RawSetString("method5", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(mod.value + 5))
		return 1
	}))
	obj.RawSetString("method6", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(mod.value + 6))
		return 1
	}))
	obj.RawSetString("method7", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(mod.value + 7))
		return 1
	}))
	obj.RawSetString("method8", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(mod.value + 8))
		return 1
	}))
	obj.RawSetString("method9", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(mod.value + 9))
		return 1
	}))
	obj.RawSetString("method10", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(mod.value + 10))
		return 1
	}))
	obj.RawSetString("method11", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(mod.value + 11))
		return 1
	}))
	obj.RawSetString("method12", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(mod.value + 12))
		return 1
	}))

	return obj
}

var methods = createMethods()

func setupLuaState(L *lua.LState, bindingType string) error {

	if bindingType == "metatable" {
		// Register metatable type (like RegisterPotatoBindableType)
		registerBindableType(L, methods)

		// Create a function to create the module
		L.SetGlobal("create_module", L.NewFunction(func(L *lua.LState) int {
			ud := newBindableModule(L)
			L.Push(ud)
			return 1
		}))
	} else {
		// Direct object without metatable
		L.SetGlobal("create_module", L.NewFunction(func(L *lua.LState) int {
			obj := newDirectModule(L)
			L.Push(obj)
			return 1
		}))
	}

	return nil
}

func main() {
	flag.Parse()

	if *bindingType != "metatable" && *bindingType != "direct" {
		fmt.Fprintf(flag.CommandLine.Output(), "Error: type must be 'metatable' or 'direct'\n")
		flag.Usage()
		return
	}

	startTime := time.Now()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	mem_before := m.Alloc
	fmt.Printf("Benchmark: %s binding type\n", *bindingType)
	fmt.Printf("Number of Lua VMs: %d\n", *numStates)
	fmt.Printf("Initial memory allocated: %s\n\n", humanize.Bytes(uint64(mem_before)))

	states := make([]*lua.LState, *numStates)
	var wg sync.WaitGroup
	done := make(chan struct{})

	// Start a goroutine to monitor memory usage
	go func() {
		ticker := time.NewTicker(time.Duration(*intervalMillis) * time.Millisecond)
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
	wg.Add(*numStates)
	for i := 0; i < *numStates; i++ {
		go func(id int) {
			defer wg.Done()
			L := lua.NewState()
			// defer L.Close()
			states[id] = L // Store the state reference to keep it alive

			// Setup bindings
			if err := setupLuaState(L, *bindingType); err != nil {
				fmt.Printf("Error setting up state %d: %v\n", id, err)
				return
			}

			// Run the benchmark script
			if err := L.DoFile("benchmark.lua"); err != nil {
				fmt.Printf("Error running script on state %d: %v\n", id, err)
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Signal the monitoring goroutine to stop
	close(done)

	// Final memory stats after all states have been used
	// runtime.GC()
	runtime.ReadMemStats(&m)
	final_mem := m.Alloc
	humanReadableMemory := humanize.Bytes(uint64(final_mem))
	memoryUsed := final_mem - mem_before
	fmt.Printf("\nFinal memory allocated: %s\n", humanReadableMemory)
	fmt.Printf("Memory used: %s\n", humanize.Bytes(uint64(memoryUsed)))
	fmt.Printf("Time taken: %s\n", time.Since(startTime))
	fmt.Printf("Average memory per VM: %s\n", humanize.Bytes(uint64(memoryUsed)/uint64(*numStates)))

	// Wait for the memory monitor to exit
	time.Sleep(500 * time.Millisecond) // Give the ticker a moment to stop
}
