package simulator

import (
	"fmt"
	"testing"
	"time"
)

// benchmarks the performance of local event generation
func BenchmarkLocalEvent(b *testing.B) {
	sim := NewSimulator(5)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sim.generateLocalEvent(i % 5)
	}
}

// benchmarks the performance of message sending operations
func BenchmarkSendMessage(b *testing.B) {
	sim := NewSimulator(5)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		from := i % 5
		to := (i + 1) % 5
		sim.sendMessage(from, to)
	}
}

// benchmarks the performance of concurrent event detection
func BenchmarkConcurrencyDetection(b *testing.B) {
	sim := NewSimulator(5)

	// Generate test events
	for i := 0; i < 100; i++ {
		sim.generateLocalEvent(i % 5)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sim.CountConcurrentEvents()
	}
}

// benchmarks simulation performance with small number of processes
func BenchmarkSimulation_SmallScale(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sim := NewSimulator(5)
		sim.RunSimulation(50*time.Millisecond, 0.3, 0.4)
	}
}

// benchmarks simulation performance with larger number of processes
func BenchmarkSimulation_LargeScale(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sim := NewSimulator(10)
		sim.RunSimulation(100*time.Millisecond, 0.3, 0.4)
	}
}

// benchmarks vector clock overhead scaling with different process counts
func BenchmarkVectorClockOverhead(b *testing.B) {
	// Compare different process counts to measure vector clock scaling
	sizes := []int{5, 10, 20, 50}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Processes_%d", size), func(b *testing.B) {
			sim := NewSimulator(size)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				sim.generateLocalEvent(i % size)
			}
		})
	}
}
