package simulator

import "fmt"

// ComplexityMetrics holds overhead measurements
type ComplexityMetrics struct {
	// Space complexity
	LamportClockSize   int // bytes per process
	VectorClockSize    int // bytes per process
	AverageMessageSize int // bytes
	TotalMemoryUsage   int // bytes

	// Message complexity
	TotalMessages         int
	AverageMessagePerProc int
	MessageOverhead       float64 // ratio of control vs data
}

// AnalyzeComplexity measures time, space, and message complexities
func (s *Simulator) AnalyzeComplexity() ComplexityMetrics {
	metrics := ComplexityMetrics{}

	// Space complexity
	// Lamport: 8 bytes (int64)
	metrics.LamportClockSize = 8

	// Vector: 8 bytes * number of processes
	metrics.VectorClockSize = 8 * s.NumProcesses

	// Message size: From(8) + To(8) + LamportTime(8) + VectorTime(8*n) + MessageID(8)
	metrics.AverageMessageSize = 32 + (8 * s.NumProcesses)

	// Total memory: (Lamport + Vector) * processes + all events
	clockMemory := (metrics.LamportClockSize + metrics.VectorClockSize) * s.NumProcesses

	// Each event: ProcessID(8) + EventType(16) + Timestamp(8) + VectorTime(8*n) + TargetID(8) + MessageID(8)
	eventSize := 48 + (8 * s.NumProcesses)
	eventsMemory := eventSize * len(s.Events)

	metrics.TotalMemoryUsage = clockMemory + eventsMemory

	// Message complexity
	stats := s.GetStatistics()
	metrics.TotalMessages = stats["send_events"].(int)
	if s.NumProcesses > 0 {
		metrics.AverageMessagePerProc = metrics.TotalMessages / s.NumProcesses
	}

	// Message overhead: vector clock adds (n-1)*8 bytes vs Lamport
	vectorOverhead := float64((s.NumProcesses - 1) * 8)
	lamportSize := float64(8)
	metrics.MessageOverhead = vectorOverhead / lamportSize

	return metrics
}

// CompareAlgorithms compares Lamport vs Vector clock overhead
func (s *Simulator) CompareAlgorithms() map[string]interface{} {
	metrics := s.AnalyzeComplexity()

	return map[string]interface{}{
		"lamport": map[string]interface{}{
			"space_per_process":     metrics.LamportClockSize,
			"message_overhead":      8, // just the timestamp
			"can_detect_concurrent": false,
		},
		"vector": map[string]interface{}{
			"space_per_process":     metrics.VectorClockSize,
			"message_overhead":      8 * s.NumProcesses, // full vector
			"can_detect_concurrent": true,
			"overhead_ratio":        float64(metrics.VectorClockSize) / float64(metrics.LamportClockSize),
		},
		"tradeoff": map[string]interface{}{
			"space_increase":       fmt.Sprintf("%.1fx", float64(metrics.VectorClockSize)/float64(metrics.LamportClockSize)),
			"message_increase":     fmt.Sprintf("%.1fx", metrics.MessageOverhead+1),
			"concurrent_detection": "Only Vector clocks can detect",
		},
	}
}
