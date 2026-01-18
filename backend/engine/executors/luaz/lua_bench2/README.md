# Lua Binding Memory Benchmark

This benchmark compares memory usage between two binding approaches:
- **Metatable bindings**: Uses `RegisterPotatoBindableType` pattern with metatable `__index` (like `bind_kv.go`)
- **Direct object methods**: Regular Lua table with `obj.xyz()` methods (no metatable)

## Usage

Run with metatable bindings (default):
```bash
go run main.go -vms=100 -type=metatable
```

Run with direct object methods:
```bash
go run main.go -vms=100 -type=direct
```

## Options

- `-vms`: Number of Lua VMs to create (default: 100)
- `-interval`: Memory monitoring interval in milliseconds (default: 200)
- `-type`: Binding type - either "metatable" or "direct" (default: "metatable")

## Example

Compare memory usage with 500 VMs:
```bash
# Test metatable bindings
go run main.go -vms=500 -type=metatable > metatable_results.txt

# Test direct object methods
go run main.go -vms=500 -type=direct > direct_results.txt

# Compare results
diff metatable_results.txt direct_results.txt
```

## What it tests

Each Lua VM:
- Creates an object with 12 methods (similar to `bind_kv.go` pattern)
- Performs 100,000 iterations
- Each iteration calls all 12 methods
- Total: 1,200,000 method calls per VM

The benchmark measures:
- Initial memory
- Memory during execution (monitored at intervals)
- Final memory after all VMs complete
- Average memory per VM

## Implementation Details

- **Metatable version**: Uses `RegisterPotatoBindableType` pattern with userdata and metatable `__index`
- **Direct version**: Creates a regular Lua table with methods attached directly, no metatable
