package simulator

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkLocalEvent(b *testing.B) {
	sim := NewSimulator(5)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sim.generateLocalEvent(i % 5)
	}
}

func BenchmarkSendMessage(b *testing.B) {
	sim := NewSimulator(5)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		from := i % 5
		to := (i + 1) % 5
		sim.sendMessage(from, to)
	}
}

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

func BenchmarkSimulation_SmallScale(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sim := NewSimulator(5)
		sim.RunSimulation(50*time.Millisecond, 0.3, 0.4)
	}
}

func BenchmarkSimulation_LargeScale(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sim := NewSimulator(10)
		sim.RunSimulation(100*time.Millisecond, 0.3, 0.4)
	}
}

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
