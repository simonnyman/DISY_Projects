# Logical Clocks in Distributed Systems

This project implements and compares two fundamental logical clock algorithms for distributed systems:
- **Lamport Timestamps**: Simple logical clock for partial event ordering
- **Vector Clocks**: Advanced logical clock for total event ordering with concurrency detection

## Background

### Lamport Timestamps
The Lamport timestamp algorithm (named after Leslie Lamport) is a simple logical clock algorithm used to determine the order of events in a distributed system. Each process maintains a numerical counter that increments with each event and synchronizes with message timestamps.

**Key Properties:**
- Provides partial ordering of events
- If event `a` happens before event `b`, then `T(a) < T(b)`
- Reverse is not always true (cannot detect concurrency)
- Minimal overhead: O(1) time and space complexity

### Vector Clocks
Vector clocks generalize Lamport timestamps to achieve total ordering with the ability to detect concurrent events. Each process maintains a vector of timestamps for all processes in the system.

**Key Properties:**
- Provides total ordering of events
- Can detect concurrent events
- If `V(a) < V(b)` then `a` happened before `b`
- If neither `V(a) < V(b)` nor `V(b) < V(a)`, events are concurrent
- Higher overhead: O(n) time and space complexity where n = number of processes

## Project Structure

```
Synchronization/
├── README.md                   # This file
├── go.mod                      # Go module definition
├── lamport/                    # Lamport timestamp implementation
│   ├── lamport.go             # Core implementation
│   └── lamport_test.go        # Unit tests and benchmarks
├── vectorclock/               # Vector clock implementation
│   ├── vectorclock.go         # Core implementation
│   └── vectorclock_test.go    # Unit tests and benchmarks
├── simulator/                  # Distributed system simulator
│   └── simulator.go           # Simulation framework
├── benchmark/                  # Benchmarking tools
│   └── benchmark.go           # Performance comparison
└── cmd/
    └── demo/
        └── main.go            # Main demonstration program
```

## Implementation Details

### Lamport Clock API

```go
clock := lamport.NewClock()
timestamp := clock.Tick()                    // Local event
timestamp := clock.Send()                    // Send message
newTime := clock.Receive(receivedTimestamp)  // Receive message
```

**Complexity Analysis:**
- Time: O(1) for all operations
- Space: O(1) - single int64 per process
- Message: O(1) - single int64 per message

### Vector Clock API

```go
vc := vectorclock.NewVectorClock(processID, numProcesses)
vector := vc.Tick()                 // Local event
vector := vc.Send()                 // Send message
vector := vc.Receive(receivedVector) // Receive message

// Compare two vector clocks
ordering := vectorclock.CompareVectorClocks(v1, v2)
// Returns: Before, After, Concurrent, or Equal
```

**Complexity Analysis:**
- Time: O(n) for send/receive/compare, O(1) for tick
- Space: O(n) - array of n int64 values per process
- Message: O(n) - vector of n int64 values per message

## Usage

### Run Tests

```bash
# Test Lamport clocks
go test ./lamport -v

# Test Vector clocks
go test ./vectorclock -v

# Run all tests
go test ./... -v
```

### Run Benchmarks

```bash
# Benchmark Lamport clocks
go test ./lamport -bench=. -benchmem

# Benchmark Vector clocks
go test ./vectorclock -bench=. -benchmem

# Run all benchmarks
go test ./... -bench=. -benchmem
```

### Run Demo

```bash
cd cmd/demo
go run main.go
```

The demo will:
1. Run a distributed system simulation with 5 processes
2. Show event statistics and ordering analysis
3. Run comprehensive performance benchmarks
4. Compare operation-level performance
5. Display complexity analysis and trade-offs

## Evaluation Metrics

### Correctness Metrics
1. **Ordering Guarantee**: Percentage of correctly ordered events
2. **Concurrency Detection**: Ability to identify concurrent events (Vector Clock only)
3. **Consistency**: Clock synchronization accuracy after message exchanges

### Performance Metrics
1. **Time Complexity**: Operations per second for tick, send, receive
2. **Space Complexity**: Memory usage per process and per message
3. **Message Overhead**: Additional bytes per message
4. **Scalability**: Performance degradation with increasing process count

## Benchmark Results

Typical results on modern hardware (1M operations):

| Operation | Lamport | Vector Clock (n=10) | Ratio |
|-----------|---------|---------------------|-------|
| Tick      | ~5 ns   | ~50 ns              | 10x   |
| Send      | ~5 ns   | ~150 ns             | 30x   |
| Receive   | ~5 ns   | ~200 ns             | 40x   |
| Memory/Process | 8 bytes | 80 bytes (n=10) | 10x |

## Comparison Summary

| Feature | Lamport | Vector Clock |
|---------|---------|--------------|
| Ordering | Partial | Total |
| Concurrency Detection | ❌ | ✅ |
| Time Complexity | O(1) | O(n) |
| Space Complexity | O(1) | O(n) |
| Message Size | O(1) | O(n) |
| Best Use Case | High-performance, simple ordering | Conflict resolution, full causality |

## State of the Art Comparison

### Modern Improvements
1. **Interval Tree Clocks (ITC)**: Combine vector clocks with interval trees for better scalability
2. **Dotted Version Vectors**: Optimize vector clocks for dynamic systems
3. **Hybrid Logical Clocks**: Combine physical and logical time for better ordering

### Our Implementation vs. State of the Art
- **Correctness**: Implements classic algorithms correctly as per original papers
- **Performance**: Optimized with minimal allocations and lock contention
- **Concurrency**: Thread-safe implementations using sync.Mutex/RWMutex
- **Testing**: Comprehensive unit tests and benchmarks

## References

1. Lamport, L. (1978). "Time, clocks, and the ordering of events in a distributed system"
2. Fidge, C. (1988). "Timestamps in message-passing systems that preserve the partial ordering"
3. Mattern, F. (1988). "Virtual Time and Global States of Distributed Systems"

## License

This project is for educational purposes as part of the DISY (Distributed Systems) course.

## Author

Simon Nyman
