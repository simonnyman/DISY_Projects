package main

import (
	"fmt"
	"time"

	"github.com/simonnyman/DISY_Projects/Synchronization/simulator"
)

// simulation configuration
const (
	numProcesses    = 20              // number of processes
	simulationTime  = 2 * time.Second // seconds the simulation runs
	localEventProb  = 0.3             // probability of local event
	sendEventProb   = 0.4             // probability of send event
	sampleEventsMax = 1               // sample events to show per process
)

func main() {

	sim := createSimulation()
	runSimulation(sim)

	displayStatistics(sim)
	displayComplexityAnalysis(sim)
	displayTimeComplexity(sim)
	displayCausalAnalysis(sim)
	displayConcurrencyAnalysis(sim)
	displayAlgorithmComparison(sim)
	displayCommunicationMatrix(sim)
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

func displayCommunicationMatrix(sim *simulator.Simulator) {
	matrix := sim.GetCommunicationMatrix()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📡 Communication Matrix (messages sent)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Print("     ")
	for i := 0; i < numProcesses; i++ {
		fmt.Printf("P%d  ", i)
	}
	fmt.Println()

	for i := 0; i < numProcesses; i++ {
		fmt.Printf("P%d   ", i)
		for j := 0; j < numProcesses; j++ {
			if i == j {
				fmt.Print(" -  ")
			} else {
				fmt.Printf("%2d  ", matrix[i][j])
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

func displayAlgorithmComparison(sim *simulator.Simulator) {
	comparison := sim.CompareAlgorithms()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("⚖️  Lamport vs Vector Clock Comparison")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Lamport Timestamp:")
	lamport := comparison["lamport"].(map[string]interface{})
	fmt.Printf("  ✓ Space per process:  %d bytes\n", lamport["space_per_process"])
	fmt.Printf("  ✓ Message overhead:   %d bytes\n", lamport["message_overhead"])
	fmt.Printf("  ✗ Concurrent detect:  %v\n", lamport["can_detect_concurrent"])

	fmt.Println("\nVector Clock:")
	vec := comparison["vector"].(map[string]interface{})
	fmt.Printf("  ✓ Space per process:  %d bytes\n", vec["space_per_process"])
	fmt.Printf("  ✓ Message overhead:   %d bytes\n", vec["message_overhead"])
	fmt.Printf("  ✓ Concurrent detect:  %v\n", vec["can_detect_concurrent"])
	fmt.Printf("  ℹ  Overhead ratio:     %.1fx\n", vec["overhead_ratio"])

	fmt.Println("\nTrade-off Summary:")
	tradeoff := comparison["tradeoff"].(map[string]interface{})
	fmt.Printf("  • Space increase:     %s\n", tradeoff["space_increase"])
	fmt.Printf("  • Message increase:   %s\n", tradeoff["message_increase"])
	fmt.Printf("  • Benefit:            %s\n", tradeoff["concurrent_detection"])
	fmt.Println()
}

func displayComplexityAnalysis(sim *simulator.Simulator) {
	metrics := sim.AnalyzeComplexity()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("⚡ Complexity Analysis")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Space Complexity:\n")
	fmt.Printf("  Lamport per process:  %6d bytes\n", metrics.LamportClockSize)
	fmt.Printf("  Vector per process:   %6d bytes (%.1fx overhead)\n",
		metrics.VectorClockSize,
		float64(metrics.VectorClockSize)/float64(metrics.LamportClockSize))
	fmt.Printf("\nMessage Complexity:\n")
	fmt.Printf("  Total messages:       %6d\n", metrics.TotalMessages)
	fmt.Printf("  Avg per process:      %6d\n", metrics.AverageMessagePerProc)
	fmt.Printf("  Avg message size:     %6d bytes\n", metrics.AverageMessageSize)
	fmt.Printf("  Vector overhead:      %.1fx\n", metrics.MessageOverhead+1)
	fmt.Printf("\nTotal Memory Usage:     %6d bytes (%.2f KB)\n",
		metrics.TotalMemoryUsage,
		float64(metrics.TotalMemoryUsage)/1024)
	fmt.Println()
}

func displayTimeComplexity(sim *simulator.Simulator) {
	theoretical := sim.CalculateTimeComplexity()
	empirical := sim.MeasureEmpiricalComplexity()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("⏱️  Time Complexity Analysis")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	fmt.Println("\nTheoretical Complexity:")
	fmt.Println("\n  Lamport Clock:")
	fmt.Printf("    Update operation:   %s\n", theoretical.LamportUpdate)
	fmt.Printf("    Compare operation:  %s\n", theoretical.LamportCompare)

	fmt.Println("\n  Vector Clock:")
	fmt.Printf("    Update operation:   %s\n", theoretical.VectorUpdate)
	fmt.Printf("    Compare operation:  %s\n", theoretical.VectorCompare)
	fmt.Printf("    Merge operation:    %s\n", theoretical.VectorMerge)

	fmt.Println("\nEmpirical Measurements (this simulation):")
	fmt.Println("\n  Lamport Clock:")
	fmt.Printf("    Total updates:      %6d operations\n", empirical.LamportUpdates)
	fmt.Printf("    Total comparisons:  %6d operations\n", empirical.LamportCompares)
	fmt.Printf("    Ops per update:     ~1 operation\n")
	fmt.Printf("    Ops per compare:    ~1 operation\n")

	fmt.Println("\n  Vector Clock:")
	fmt.Printf("    Total updates:      %6d operations\n", empirical.VectorUpdates)
	fmt.Printf("    Total comparisons:  %6d operations\n", empirical.VectorCompares)
	fmt.Printf("    Ops per update:     ~%.0f operations (vector size)\n", empirical.VectorOpsPerUpdate)
	fmt.Printf("    Ops per compare:    ~%.0f operations (vector size)\n", empirical.VectorOpsPerCompare)

	fmt.Println("\n  Complexity Ratio:")
	totalLamportOps := float64(empirical.LamportUpdates + empirical.LamportCompares)
	totalVectorOps := float64(empirical.VectorUpdates+empirical.VectorCompares) * empirical.VectorOpsPerUpdate
	fmt.Printf("    Vector/Lamport:     %.1fx more operations\n", totalVectorOps/totalLamportOps)
	fmt.Println()
}

func displayCausalAnalysis(sim *simulator.Simulator) {
	relations := sim.CountCausalRelationships()
	total := relations["before"] + relations["after"] + relations["concurrent"] + relations["equal"]

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🔗 Causal Relationship Analysis")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Before relationships:     %6d (%.1f%%)\n",
		relations["before"], percentage(relations["before"], total))
	fmt.Printf("After relationships:      %6d (%.1f%%)\n",
		relations["after"], percentage(relations["after"], total))
	fmt.Printf("Concurrent relationships: %6d (%.1f%%)\n",
		relations["concurrent"], percentage(relations["concurrent"], total))
	fmt.Printf("Equal timestamps:         %6d (%.1f%%)\n",
		relations["equal"], percentage(relations["equal"], total))
	fmt.Printf("Total event pairs:        %6d\n", total)
	fmt.Println()
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
