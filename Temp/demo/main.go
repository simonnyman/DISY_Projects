package main

import (
	"fmt"
	"time"

	"github.com/simonnyman/DISY_Projects/synchronization/benchmark"
	"github.com/simonnyman/DISY_Projects/synchronization/simulator"
)

func main() {
	fmt.Println("=== Lamport Timestamp and Vector Clock Implementation ===")
	fmt.Println()
	
	// Run demonstration simulation
	fmt.Println("Running simulation with 5 processes for 2 seconds...")
	runSimulation()
	
	// Run benchmarks
	fmt.Println("\nRunning comprehensive benchmarks...")
	runBenchmarks()
	
	// Run operation-level comparison
	fmt.Println("\nRunning operation-level comparison...")
	runOperationComparison()
}

func runSimulation() {
	numProcesses := 5
	duration := 2 * time.Second
	localEventRate := 0.3
	messageRate := 0.4
	
	sim := simulator.NewSimulator(numProcesses)
	
	startTime := time.Now()
	sim.RunSimulation(duration, localEventRate, messageRate)
	elapsed := time.Since(startTime)
	
	// Print statistics
	stats := sim.GetStatistics()
	fmt.Println("\n--- Simulation Statistics ---")
	fmt.Printf("Duration: %v\n", elapsed)
	fmt.Printf("Number of processes: %d\n", stats["num_processes"])
	fmt.Printf("Total events: %d\n", stats["total_events"])
	fmt.Printf("Local events: %d\n", stats["local_events"])
	fmt.Printf("Send events: %d\n", stats["send_events"])
	fmt.Printf("Receive events: %d\n", stats["receive_events"])
	
	// Analyze ordering
	ordering := sim.AnalyzeOrdering()
	fmt.Println("\n--- Ordering Analysis ---")
	fmt.Printf("Total pairs analyzed: %d\n", ordering["total_pairs_analyzed"])
	fmt.Printf("Concurrent pairs: %d\n", ordering["concurrent_pairs"])
	fmt.Printf("Concurrency rate: %.2f%%\n", ordering["concurrency_rate"].(float64)*100)
	
	// Print sample events from each process
	fmt.Println("\n--- Sample Events per Process ---")
	for i := 0; i < numProcesses; i++ {
		process := sim.Processes[i]
		eventCount := len(process.Events)
		fmt.Printf("Process %d: %d events\n", i, eventCount)
		
		// Show first 3 events
		maxShow := 3
		if eventCount < maxShow {
			maxShow = eventCount
		}
		for j := 0; j < maxShow; j++ {
			event := process.Events[j]
			fmt.Printf("  [%s] Lamport: %d, Vector: %v\n",
				event.EventType, event.Timestamp, event.VectorTime)
		}
	}
}

func runBenchmarks() {
	results := benchmark.RunBenchmark()
	benchmark.PrintResults(results)
	
	// Analysis
	fmt.Println("=== Analysis ===")
	fmt.Println("\nKey Observations:")
	fmt.Println("1. Time Complexity:")
	fmt.Println("   - Lamport: O(1) for all operations (tick, send, receive)")
	fmt.Println("   - Vector Clock: O(n) for send and receive, where n is number of processes")
	fmt.Println()
	fmt.Println("2. Space Complexity:")
	fmt.Println("   - Lamport: O(1) - single integer per process")
	fmt.Println("   - Vector Clock: O(n) - array of n integers per process")
	fmt.Println()
	fmt.Println("3. Message Complexity:")
	fmt.Println("   - Lamport: O(1) - single timestamp per message")
	fmt.Println("   - Vector Clock: O(n) - vector of n timestamps per message")
	fmt.Println()
	fmt.Println("4. Ordering Guarantees:")
	fmt.Println("   - Lamport: Partial ordering (if a->b then T(a)<T(b), but not vice versa)")
	fmt.Println("   - Vector Clock: Total ordering (can determine all causality relationships)")
	fmt.Println()
	fmt.Println("5. Concurrency Detection:")
	fmt.Println("   - Lamport: Cannot detect concurrent events")
	fmt.Println("   - Vector Clock: Can detect concurrent events")
}

func runOperationComparison() {
	iterations := 1000000
	results := benchmark.CompareOperations(iterations)
	benchmark.PrintOperationComparison(results, iterations)
	
	fmt.Println("=== Performance Summary ===")
	fmt.Println("Lamport clocks are significantly faster due to O(1) operations.")
	fmt.Println("Vector clocks provide more information but at the cost of O(n) complexity.")
	fmt.Println()
	fmt.Println("Trade-offs:")
	fmt.Println("- Use Lamport when: Performance is critical and partial ordering is sufficient")
	fmt.Println("- Use Vector Clock when: Full causality detection is needed (e.g., conflict resolution)")
}
