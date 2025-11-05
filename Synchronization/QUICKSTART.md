# Quick Start Guide

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