package main

import (
	"fmt"

	"github.com/simonnyman/DISY_Projects/synchronization/lamport"
	"github.com/simonnyman/DISY_Projects/synchronization/vectorclock"
)

// Example demonstrating basic usage of Lamport clocks and Vector clocks

func main() {
	fmt.Println("=== Basic Usage Examples ===")
	fmt.Println()

	// Lamport Clock Example
	fmt.Println("1. Lamport Clock Example:")
	lamportExample()

	fmt.Println("\n2. Vector Clock Example:")
	vectorClockExample()

	fmt.Println("\n3. Comparison Example:")
	comparisonExample()
}

func lamportExample() {
	// Create clocks for two processes
	p1 := lamport.NewClock()
	p2 := lamport.NewClock()

	// Process 1: Local event
	t1 := p1.Tick()
	fmt.Printf("Process 1 - Local event: timestamp = %d\n", t1)

	// Process 1: Send message
	msgTimestamp := p1.Send()
	fmt.Printf("Process 1 - Send message: timestamp = %d\n", msgTimestamp)

	// Process 2: Receive message
	t2 := p2.Receive(msgTimestamp)
	fmt.Printf("Process 2 - Receive message: timestamp = %d\n", t2)

	// Process 2: Local event
	t3 := p2.Tick()
	fmt.Printf("Process 2 - Local event: timestamp = %d\n", t3)

	// Create and compare messages
	msg1 := lamport.NewMessage(msgTimestamp, 1, "Hello")
	msg2 := lamport.NewMessage(t3, 2, "World")

	if msg1.HappensBefore(msg2) {
		fmt.Printf("Message from P1 (t=%d) happened before message from P2 (t=%d)\n",
			msg1.Timestamp, msg2.Timestamp)
	}
}

func vectorClockExample() {
	numProcesses := 3

	// Create vector clocks for three processes
	p0 := vectorclock.NewVectorClock(0, numProcesses)
	p1 := vectorclock.NewVectorClock(1, numProcesses)
	p2 := vectorclock.NewVectorClock(2, numProcesses)

	// Process 0: Local event
	v1 := p0.Tick()
	fmt.Printf("Process 0 - Local event: %v\n", v1)

	// Process 1: Local event
	v2 := p1.Tick()
	fmt.Printf("Process 1 - Local event: %v\n", v2)

	// Process 0: Send message to Process 1
	msgVector := p0.Send()
	fmt.Printf("Process 0 - Send message: %v\n", msgVector)

	// Process 1: Receive message from Process 0
	v3 := p1.Receive(msgVector)
	fmt.Printf("Process 1 - Receive message: %v\n", v3)

	// Process 2: Local event (concurrent with others)
	v4 := p2.Tick()
	fmt.Printf("Process 2 - Local event: %v\n", v4)

	// Compare messages
	msg1 := vectorclock.NewMessage(msgVector, 0, "From P0")
	msg2 := vectorclock.NewMessage(v3, 1, "From P1")
	msg3 := vectorclock.NewMessage(v4, 2, "From P2")

	fmt.Println("\nOrdering relationships:")

	ordering := msg1.CompareTo(msg2)
	switch ordering {
	case vectorclock.Before:
		fmt.Printf("Message from P0 happened before message from P1\n")
	case vectorclock.After:
		fmt.Printf("Message from P0 happened after message from P1\n")
	case vectorclock.Concurrent:
		fmt.Printf("Message from P0 is concurrent with message from P1\n")
	}

	ordering = msg1.CompareTo(msg3)
	switch ordering {
	case vectorclock.Concurrent:
		fmt.Printf("Message from P0 is concurrent with message from P2\n")
	default:
		fmt.Printf("Message from P0 and P2 have definite ordering\n")
	}
}

func comparisonExample() {
	fmt.Println("Key Differences:")
	fmt.Println()

	fmt.Println("Lamport Clocks:")
	fmt.Println("  ✓ Simple and efficient (O(1) time and space)")
	fmt.Println("  ✓ Small message overhead (single integer)")
	fmt.Println("  ✗ Only partial ordering")
	fmt.Println("  ✗ Cannot detect concurrent events")
	fmt.Println()

	fmt.Println("Vector Clocks:")
	fmt.Println("  ✓ Total ordering capability")
	fmt.Println("  ✓ Can detect concurrent events")
	fmt.Println("  ✓ More precise causality tracking")
	fmt.Println("  ✗ Higher overhead (O(n) time and space)")
	fmt.Println("  ✗ Larger message size (n integers)")
	fmt.Println()

	fmt.Println("Use Cases:")
	fmt.Println("  - Lamport: Event logging, distributed timestamps, simple ordering")
	fmt.Println("  - Vector Clock: Conflict detection, replicated databases, version control")
}
