# Quick Start Guide

## Installation

This project requires Go 1.21 or later.

```bash
cd /Users/simon_nyman/GitHub/DISY_Projects/Synchronization
go mod download
```

## Running Tests

### Run all tests
```bash
go test ./... -v
```

### Run specific package tests
```bash
# Lamport clock tests
go test ./lamport -v

# Vector clock tests
go test ./vectorclock -v
```

### Run benchmarks
```bash
# All benchmarks
go test ./... -bench=. -benchmem

# Specific benchmarks
go test ./lamport -bench=. -benchmem
go test ./vectorclock -bench=. -benchmem
```

## Running Examples

### Basic Usage Example
```bash
go run cmd/examples/main.go
```

This demonstrates:
- Creating and using Lamport clocks
- Creating and using Vector clocks
- Comparing messages and detecting ordering
- Understanding when to use each algorithm

### Full Demo with Benchmarks
```bash
go run cmd/demo/main.go
```

This includes:
- Distributed system simulation
- Performance benchmarks
- Complexity analysis
- Trade-off comparison

## Basic Usage

### Lamport Clocks

```go
import "github.com/simonnyman/DISY_Projects/synchronization/lamport"

// Create a clock
clock := lamport.NewClock()

// Local event
timestamp := clock.Tick()

// Send message
sendTime := clock.Send()  // Include this in your message

// Receive message
newTime := clock.Receive(receivedTimestamp)

// Create and compare messages
msg1 := lamport.NewMessage(timestamp1, processID1, data)
msg2 := lamport.NewMessage(timestamp2, processID2, data)
if msg1.HappensBefore(msg2) {
    // msg1 happened before msg2
}
```

### Vector Clocks

```go
import "github.com/simonnyman/DISY_Projects/synchronization/vectorclock"

// Create a vector clock (processID, total number of processes)
vc := vectorclock.NewVectorClock(0, 3)

// Local event
vector := vc.Tick()

// Send message
sendVector := vc.Send()  // Include this in your message

// Receive message
newVector := vc.Receive(receivedVector)

// Create and compare messages
msg1 := vectorclock.NewMessage(vector1, processID1, data)
msg2 := vectorclock.NewMessage(vector2, processID2, data)

ordering := msg1.CompareTo(msg2)
switch ordering {
case vectorclock.Before:
    // msg1 happened before msg2
case vectorclock.After:
    // msg1 happened after msg2
case vectorclock.Concurrent:
    // msg1 and msg2 are concurrent
case vectorclock.Equal:
    // same timestamp
}
```

## Project Structure

```
Synchronization/
├── README.md              # Main documentation
├── ANALYSIS.md            # Detailed comparative analysis
├── QUICKSTART.md          # This file
├── go.mod                 # Go module definition
│
├── lamport/               # Lamport timestamp implementation
│   ├── lamport.go        # Core implementation
│   └── lamport_test.go   # Tests and benchmarks
│
├── vectorclock/           # Vector clock implementation
│   ├── vectorclock.go    # Core implementation
│   └── vectorclock_test.go # Tests and benchmarks
│
├── simulator/             # Distributed system simulator
│   └── simulator.go      # Simulation framework
│
├── benchmark/             # Benchmarking tools
│   └── benchmark.go      # Performance comparison
│
└── cmd/
    ├── demo/             # Full demonstration
    │   └── main.go
    └── examples/         # Basic usage examples
        └── main.go
```

## Understanding the Output

### Test Output
- `PASS` indicates all tests passed
- Each test shows execution time
- Use `-v` flag for verbose output

### Benchmark Output
Example:
```
BenchmarkTick-10    116938281    10.02 ns/op    0 B/op    0 allocs/op
```
- `116938281` iterations run
- `10.02 ns/op` average time per operation
- `0 B/op` bytes allocated per operation
- `0 allocs/op` number of allocations per operation

### Simulation Output
- **Total events**: All events in the system
- **Local events**: Events within a single process
- **Send/Receive events**: Communication between processes
- **Concurrency rate**: Percentage of concurrent event pairs

## Common Use Cases

### 1. Event Logging
Use **Lamport clocks** for timestamping events in distributed logs:
```go
clock := lamport.NewClock()
timestamp := clock.Tick()
log.Printf("[%d] Event occurred", timestamp)
```

### 2. Message Ordering
Use **Lamport clocks** for simple message ordering:
```go
sendTime := clock.Send()
message := Message{Timestamp: sendTime, Data: data}
// Send message...
```

### 3. Conflict Detection
Use **Vector clocks** for detecting conflicting updates:
```go
vc := vectorclock.NewVectorClock(processID, numProcesses)
updateVector := vc.Send()
update := Update{Version: updateVector, Data: data}

// On receiving two updates
if update1.IsConcurrent(update2) {
    // These are conflicting updates - need resolution
    resolveConflict(update1, update2)
}
```

### 4. Distributed Debugging
Use **Vector clocks** for understanding causality:
```go
event1 := Event{Vector: v1, Description: "User login"}
event2 := Event{Vector: v2, Description: "Data access"}

if event1.HappensBefore(event2) {
    fmt.Println("Login definitely happened before access")
} else if event1.IsConcurrent(event2) {
    fmt.Println("Potential security issue: concurrent events")
}
```

## Performance Tips

### Lamport Clocks
- ✓ Use for high-throughput systems
- ✓ Ideal when ordering most events is sufficient
- ✓ Minimal memory footprint
- ✓ Thread-safe operations with low contention

### Vector Clocks
- ✓ Use when you need complete causality
- ⚠ Consider limiting number of processes
- ⚠ Memory grows linearly with processes
- ✓ Good for moderate-scale systems (<100 processes)

## Troubleshooting

### Issue: Tests fail with "cannot find package"
**Solution:** Run `go mod download` first

### Issue: Benchmarks show inconsistent results
**Solution:** Run multiple times and use `-benchtime=10s` for longer runs

### Issue: Simulation deadlocks
**Solution:** Check that channels are properly closed and all goroutines can exit

### Issue: High memory usage in vector clocks
**Solution:** This is expected - memory scales with O(n) where n = number of processes

## Next Steps

1. Read `README.md` for comprehensive documentation
2. Read `ANALYSIS.md` for detailed comparative analysis
3. Run the examples to understand basic usage
4. Run the demo for comprehensive evaluation
5. Explore the test files for more usage patterns
6. Modify the simulator for your specific use case

## Additional Resources

- [Lamport's Original Paper (1978)](https://lamport.azurewebsites.net/pubs/time-clocks.pdf)
- [Vector Clocks Explanation](https://en.wikipedia.org/wiki/Vector_clock)
- [Distributed Systems Concepts](https://www.distributed-systems.net/)

## Support

For questions or issues:
1. Check the code comments in source files
2. Review test cases for usage examples
3. Read the comprehensive documentation in README.md
4. Examine the comparative analysis in ANALYSIS.md
