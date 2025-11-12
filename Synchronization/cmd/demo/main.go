package main

import (
	"fmt"
	"time"

	"github.com/simonnyman/DISY_Projects/Synchronization/simulator"
)

// simulation configuration
const (
	numProcesses    = 5               // number of processes
	simulationTime  = 2 * time.Second // seconds the simulation runs
	localEventProb  = 0.3             // probability of local event
	sendEventProb   = 0.4             // probability of send event
	sampleEventsMax = 20              // sample events to show per process
)

func main() {
	sim := createSimulation()
	runSimulation(sim)

	displayStatistics(sim)
	displayConcurrencyAnalysis(sim)
	displaySampleEvents(sim)
}

func createSimulation() *simulator.Simulator {
	fmt.Printf("Configuration:\n")
	fmt.Printf("Processes: %d\n", numProcesses)
	fmt.Printf("Duration: %s\n", simulationTime)
	fmt.Printf("Local event probability: %.0f%%\n", localEventProb*100)
	fmt.Printf("Send message probability: %.0f%%\n", sendEventProb*100)
	fmt.Println()

	return simulator.NewSimulator(numProcesses)
}

func runSimulation(sim *simulator.Simulator) {
	fmt.Println("Running simulation...")
	sim.RunSimulation(simulationTime, localEventProb, sendEventProb)
	fmt.Println("Simulation complete")
	fmt.Println()
}

func displayStatistics(sim *simulator.Simulator) {
	stats := sim.GetStatistics()
	total := stats["total_events"].(int)

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Event Statistics")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Total events:    %6d\n", total)
	fmt.Printf("Local events:    %6d (%.1f%%)\n",
		stats["local_events"].(int),
		percentage(stats["local_events"].(int), total))
	fmt.Printf("Send events:     %6d (%.1f%%)\n",
		stats["send_events"].(int),
		percentage(stats["send_events"].(int), total))
	fmt.Printf("Receive events:  %6d (%.1f%%)\n",
		stats["receive_events"].(int),
		percentage(stats["receive_events"].(int), total))
	fmt.Println()
}

func displayConcurrencyAnalysis(sim *simulator.Simulator) {
	stats := sim.GetStatistics()
	totalEvents := stats["total_events"].(int)
	totalPairs := totalEvents * (totalEvents - 1) / 2
	concurrentPairs := sim.CountConcurrentEvents()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Concurrency Analysis")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Concurrent pairs:  %6d\n", concurrentPairs)
	fmt.Printf("Total pairs:       %6d\n", totalPairs)
	fmt.Printf("Concurrency rate:  %6.2f%%\n", percentage(concurrentPairs, totalPairs))
	fmt.Println()
}

func displaySampleEvents(sim *simulator.Simulator) {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Sample Events (first %d per process)\n", sampleEventsMax)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for i := 0; i < numProcesses; i++ {
		displayProcessEvents(sim.Processes[i], i)
	}
}

func displayProcessEvents(process *simulator.Process, id int) {
	eventCount := len(process.Events)
	fmt.Printf("\n Process %d (%d total events):\n", id, eventCount)

	maxShow := min(sampleEventsMax, eventCount)
	for j := 0; j < maxShow; j++ {
		event := process.Events[j]

		switch event.EventType {
		case "local":
			fmt.Printf("   [%8s] Lamport: %3d, Vector: %v\n",
				event.EventType, event.Timestamp, event.VectorTime)
		case "send":
			fmt.Printf("   [%8s] Lamport: %3d, Vector: %v → P%d (msg#%d)\n",
				event.EventType, event.Timestamp, event.VectorTime,
				event.TargetID, event.MessageID)
		case "receive":
			fmt.Printf("   [%8s] Lamport: %3d, Vector: %v ← P%d (msg#%d)\n",
				event.EventType, event.Timestamp, event.VectorTime,
				event.TargetID, event.MessageID)
		}
	}
}

// helper functions
func percentage(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
