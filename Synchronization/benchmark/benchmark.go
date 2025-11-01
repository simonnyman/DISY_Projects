package benchmark

import (
	"fmt"
	"time"

	"github.com/simonnyman/DISY_Projects/synchronization/lamport"
	"github.com/simonnyman/DISY_Projects/synchronization/simulator"
	"github.com/simonnyman/DISY_Projects/synchronization/vectorclock"
)

// Result represents benchmark results
type Result struct {
	Algorithm        string
	NumProcesses     int
	Duration         time.Duration
	TotalEvents      int
	EventsPerSecond  float64
	MemoryPerProcess int // bytes
	Overhead         string
}

// RunBenchmark runs comprehensive benchmarks comparing algorithms
func RunBenchmark() []Result {
	results := make([]Result, 0)

	// Test configurations
	processCounts := []int{3, 5, 10, 20, 50}
	duration := 1 * time.Second

	for _, numProcs := range processCounts {
		// Benchmark Lamport clocks
		lamportResult := benchmarkLamport(numProcs, duration)
		results = append(results, lamportResult)

		// Benchmark Vector clocks
		vectorResult := benchmarkVector(numProcs, duration)
		results = append(results, vectorResult)
	}

	return results
}

func benchmarkLamport(numProcesses int, duration time.Duration) Result {
	sim := simulator.NewSimulator(numProcesses)

	startTime := time.Now()
	sim.RunSimulation(duration, 0.4, 0.4)
	elapsed := time.Since(startTime)

	stats := sim.GetStatistics()
	totalEvents := stats["total_events"].(int)
	eventsPerSec := float64(totalEvents) / elapsed.Seconds()

	// Memory calculation: Lamport uses 8 bytes per process (int64)
	memoryPerProcess := 8

	return Result{
		Algorithm:        "Lamport",
		NumProcesses:     numProcesses,
		Duration:         elapsed,
		TotalEvents:      totalEvents,
		EventsPerSecond:  eventsPerSec,
		MemoryPerProcess: memoryPerProcess,
		Overhead:         "O(1) time, O(1) space",
	}
}

func benchmarkVector(numProcesses int, duration time.Duration) Result {
	sim := simulator.NewSimulator(numProcesses)

	startTime := time.Now()
	sim.RunSimulation(duration, 0.4, 0.4)
	elapsed := time.Since(startTime)

	stats := sim.GetStatistics()
	totalEvents := stats["total_events"].(int)
	eventsPerSec := float64(totalEvents) / elapsed.Seconds()

	// Memory calculation: Vector clock uses 8 bytes * numProcesses per process
	memoryPerProcess := 8 * numProcesses

	return Result{
		Algorithm:        "Vector Clock",
		NumProcesses:     numProcesses,
		Duration:         elapsed,
		TotalEvents:      totalEvents,
		EventsPerSecond:  eventsPerSec,
		MemoryPerProcess: memoryPerProcess,
		Overhead:         fmt.Sprintf("O(n) time, O(n) space where n=%d", numProcesses),
	}
}

// CompareOperations compares individual operations
func CompareOperations(iterations int) map[string]time.Duration {
	results := make(map[string]time.Duration)
	numProcesses := 10

	// Benchmark Lamport Tick
	lClock := lamport.NewClock()
	start := time.Now()
	for i := 0; i < iterations; i++ {
		lClock.Tick()
	}
	results["Lamport_Tick"] = time.Since(start)

	// Benchmark Lamport Send
	lClock.Reset()
	start = time.Now()
	for i := 0; i < iterations; i++ {
		lClock.Send()
	}
	results["Lamport_Send"] = time.Since(start)

	// Benchmark Lamport Receive
	lClock.Reset()
	start = time.Now()
	for i := 0; i < iterations; i++ {
		lClock.Receive(int64(i))
	}
	results["Lamport_Receive"] = time.Since(start)

	// Benchmark Vector Clock Tick
	vClock := vectorclock.NewVectorClock(0, numProcesses)
	start = time.Now()
	for i := 0; i < iterations; i++ {
		vClock.Tick()
	}
	results["VectorClock_Tick"] = time.Since(start)

	// Benchmark Vector Clock Send
	vClock.Reset()
	start = time.Now()
	for i := 0; i < iterations; i++ {
		vClock.Send()
	}
	results["VectorClock_Send"] = time.Since(start)

	// Benchmark Vector Clock Receive
	vClock.Reset()
	received := make([]int64, numProcesses)
	start = time.Now()
	for i := 0; i < iterations; i++ {
		vClock.Receive(received)
	}
	results["VectorClock_Receive"] = time.Since(start)

	// Benchmark Vector Clock Compare
	v1 := make([]int64, numProcesses)
	v2 := make([]int64, numProcesses)
	for i := 0; i < numProcesses; i++ {
		v1[i] = int64(i)
		v2[i] = int64(i + 1)
	}
	start = time.Now()
	for i := 0; i < iterations; i++ {
		vectorclock.CompareVectorClocks(v1, v2)
	}
	results["VectorClock_Compare"] = time.Since(start)

	return results
}

// PrintResults prints benchmark results in a formatted table
func PrintResults(results []Result) {
	fmt.Println("\n=== Benchmark Results ===")
	fmt.Printf("%-15s %-12s %-12s %-15s %-20s %-25s\n",
		"Algorithm", "Processes", "Events", "Events/sec", "Memory/Process", "Complexity")
	fmt.Println("-----------------------------------------------------------------------------------------------------------")

	for _, r := range results {
		fmt.Printf("%-15s %-12d %-12d %-15.2f %-20d %-25s\n",
			r.Algorithm, r.NumProcesses, r.TotalEvents, r.EventsPerSecond,
			r.MemoryPerProcess, r.Overhead)
	}
	fmt.Println()
}

// PrintOperationComparison prints operation-level comparison
func PrintOperationComparison(results map[string]time.Duration, iterations int) {
	fmt.Println("\n=== Operation-Level Comparison ===")
	fmt.Printf("Operations performed: %d\n\n", iterations)
	fmt.Printf("%-25s %-15s %-20s\n", "Operation", "Total Time", "Time per Operation")
	fmt.Println("--------------------------------------------------------------")

	for op, duration := range results {
		timePerOp := duration / time.Duration(iterations)
		fmt.Printf("%-25s %-15s %-20s\n", op, duration, timePerOp)
	}
	fmt.Println()
}
