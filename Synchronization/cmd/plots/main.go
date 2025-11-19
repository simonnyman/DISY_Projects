package main

import (
	"fmt"
	"image/color"
	"time"

	"github.com/simonnyman/DISY_Projects/Synchronization/simulator"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

// simulation configuration
const (
	simulationTime = 1 * time.Second
	localEventProb = 0.3
	sendEventProb  = 0.4
	numRuns        = 3 // number of runs per process count (for averaging)
)

// process counts to test
var processCounts = []int{2, 4, 6, 8, 10, 15, 20, 25, 30}

type SimulationResult struct {
	NumProcesses    int
	TotalEvents     int
	LocalEvents     int
	SendEvents      int
	ReceiveEvents   int
	ConcurrentPairs int
	TotalPairs      int
	LamportMemory   int
	VectorMemory    int
	TotalMessages   int
	ConcurrencyRate float64
	MemoryOverhead  float64
}

func main() {
	fmt.Println("Running simulations with varying process counts...")
	fmt.Println("This may take a minute...")
	fmt.Println()

	results := runSimulations()

	fmt.Println("Generating plots...")
	generateEventCountPlot(results)
	generateEventTypePlot(results)
	generateConcurrencyPlot(results)
	generateMemoryUsagePlot(results)
	generateMessageCountPlot(results)
	generateMemoryOverheadPlot(results)

	fmt.Println("\n✓ All plots generated successfully!")
	fmt.Println("\nGenerated files:")
	fmt.Println("  - event_count.png")
	fmt.Println("  - event_types.png")
	fmt.Println("  - concurrency_rate.png")
	fmt.Println("  - memory_usage.png")
	fmt.Println("  - message_count.png")
	fmt.Println("  - memory_overhead.png")
}

func runSimulations() []SimulationResult {
	results := make([]SimulationResult, 0)

	for _, numProcs := range processCounts {
		fmt.Printf("Running simulation with %d processes...\n", numProcs)

		// Run multiple times and average results
		var avgResult SimulationResult
		avgResult.NumProcesses = numProcs

		for run := 0; run < numRuns; run++ {
			sim := simulator.NewSimulator(numProcs)
			sim.RunSimulation(simulationTime, localEventProb, sendEventProb)

			stats := sim.GetStatistics()
			metrics := sim.AnalyzeComplexity()

			avgResult.TotalEvents += stats["total_events"].(int)
			avgResult.LocalEvents += stats["local_events"].(int)
			avgResult.SendEvents += stats["send_events"].(int)
			avgResult.ReceiveEvents += stats["receive_events"].(int)

			concurrentPairs := sim.CountConcurrentEvents()
			totalEvents := stats["total_events"].(int)
			totalPairs := totalEvents * (totalEvents - 1) / 2

			avgResult.ConcurrentPairs += concurrentPairs
			avgResult.TotalPairs += totalPairs

			avgResult.LamportMemory += metrics.LamportClockSize * numProcs
			avgResult.VectorMemory += metrics.VectorClockSize * numProcs
			avgResult.TotalMessages += metrics.TotalMessages
		}

		// Calculate averages
		avgResult.TotalEvents /= numRuns
		avgResult.LocalEvents /= numRuns
		avgResult.SendEvents /= numRuns
		avgResult.ReceiveEvents /= numRuns
		avgResult.ConcurrentPairs /= numRuns
		avgResult.TotalPairs /= numRuns
		avgResult.LamportMemory /= numRuns
		avgResult.VectorMemory /= numRuns
		avgResult.TotalMessages /= numRuns

		if avgResult.TotalPairs > 0 {
			avgResult.ConcurrencyRate = float64(avgResult.ConcurrentPairs) / float64(avgResult.TotalPairs) * 100
		}

		if avgResult.LamportMemory > 0 {
			avgResult.MemoryOverhead = float64(avgResult.VectorMemory) / float64(avgResult.LamportMemory)
		}

		results = append(results, avgResult)
	}

	return results
}

func generateEventCountPlot(results []SimulationResult) {
	p := plot.New()
	p.Title.Text = "Total Events vs Number of Processes"
	p.X.Label.Text = "Number of Processes"
	p.Y.Label.Text = "Total Events"

	pts := make(plotter.XYs, len(results))
	for i, r := range results {
		pts[i].X = float64(r.NumProcesses)
		pts[i].Y = float64(r.TotalEvents)
	}

	line, points, err := plotter.NewLinePoints(pts)
	if err != nil {
		panic(err)
	}
	line.Color = color.RGBA{R: 255, A: 255}
	points.Color = color.RGBA{R: 255, A: 255}
	points.Shape = plotutil.DefaultGlyphShapes[0]

	p.Add(line, points)
	p.Add(plotter.NewGrid())

	if err := p.Save(8*vg.Inch, 6*vg.Inch, "event_count.png"); err != nil {
		panic(err)
	}
}

func generateEventTypePlot(results []SimulationResult) {
	p := plot.New()
	p.Title.Text = "Event Types vs Number of Processes"
	p.X.Label.Text = "Number of Processes"
	p.Y.Label.Text = "Event Count"
	p.Legend.Top = true

	localPts := make(plotter.XYs, len(results))
	sendPts := make(plotter.XYs, len(results))
	receivePts := make(plotter.XYs, len(results))

	for i, r := range results {
		localPts[i].X = float64(r.NumProcesses)
		localPts[i].Y = float64(r.LocalEvents)
		sendPts[i].X = float64(r.NumProcesses)
		sendPts[i].Y = float64(r.SendEvents)
		receivePts[i].X = float64(r.NumProcesses)
		receivePts[i].Y = float64(r.ReceiveEvents)
	}

	localLine, localPoints, _ := plotter.NewLinePoints(localPts)
	localLine.Color = color.RGBA{R: 0, G: 255, B: 0, A: 255}
	localPoints.Color = color.RGBA{R: 0, G: 255, B: 0, A: 255}

	sendLine, sendPoints, _ := plotter.NewLinePoints(sendPts)
	sendLine.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	sendPoints.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}

	receiveLine, receivePoints, _ := plotter.NewLinePoints(receivePts)
	receiveLine.Color = color.RGBA{R: 0, G: 0, B: 255, A: 255}
	receivePoints.Color = color.RGBA{R: 0, G: 0, B: 255, A: 255}

	p.Add(localLine, localPoints, sendLine, sendPoints, receiveLine, receivePoints)
	p.Add(plotter.NewGrid())
	p.Legend.Add("Local Events", localLine, localPoints)
	p.Legend.Add("Send Events", sendLine, sendPoints)
	p.Legend.Add("Receive Events", receiveLine, receivePoints)

	if err := p.Save(8*vg.Inch, 6*vg.Inch, "event_types.png"); err != nil {
		panic(err)
	}
}

func generateConcurrencyPlot(results []SimulationResult) {
	p := plot.New()
	p.Title.Text = "Concurrency Rate vs Number of Processes"
	p.X.Label.Text = "Number of Processes"
	p.Y.Label.Text = "Concurrency Rate (%)"

	pts := make(plotter.XYs, len(results))
	for i, r := range results {
		pts[i].X = float64(r.NumProcesses)
		pts[i].Y = r.ConcurrencyRate
	}

	line, points, err := plotter.NewLinePoints(pts)
	if err != nil {
		panic(err)
	}
	line.Color = color.RGBA{R: 128, G: 0, B: 128, A: 255}
	points.Color = color.RGBA{R: 128, G: 0, B: 128, A: 255}

	p.Add(line, points)
	p.Add(plotter.NewGrid())

	if err := p.Save(8*vg.Inch, 6*vg.Inch, "concurrency_rate.png"); err != nil {
		panic(err)
	}
}

func generateMemoryUsagePlot(results []SimulationResult) {
	p := plot.New()
	p.Title.Text = "Memory Usage vs Number of Processes"
	p.X.Label.Text = "Number of Processes"
	p.Y.Label.Text = "Memory (bytes)"
	p.Legend.Top = true

	lamportPts := make(plotter.XYs, len(results))
	vectorPts := make(plotter.XYs, len(results))

	for i, r := range results {
		lamportPts[i].X = float64(r.NumProcesses)
		lamportPts[i].Y = float64(r.LamportMemory)
		vectorPts[i].X = float64(r.NumProcesses)
		vectorPts[i].Y = float64(r.VectorMemory)
	}

	lamportLine, lamportPoints, _ := plotter.NewLinePoints(lamportPts)
	lamportLine.Color = color.RGBA{R: 255, G: 165, B: 0, A: 255}
	lamportPoints.Color = color.RGBA{R: 255, G: 165, B: 0, A: 255}

	vectorLine, vectorPoints, _ := plotter.NewLinePoints(vectorPts)
	vectorLine.Color = color.RGBA{R: 0, G: 128, B: 255, A: 255}
	vectorPoints.Color = color.RGBA{R: 0, G: 128, B: 255, A: 255}

	p.Add(lamportLine, lamportPoints, vectorLine, vectorPoints)
	p.Add(plotter.NewGrid())
	p.Legend.Add("Lamport Clock", lamportLine, lamportPoints)
	p.Legend.Add("Vector Clock", vectorLine, vectorPoints)

	if err := p.Save(8*vg.Inch, 6*vg.Inch, "memory_usage.png"); err != nil {
		panic(err)
	}
}

func generateMessageCountPlot(results []SimulationResult) {
	p := plot.New()
	p.Title.Text = "Total Messages vs Number of Processes"
	p.X.Label.Text = "Number of Processes"
	p.Y.Label.Text = "Total Messages Sent"

	pts := make(plotter.XYs, len(results))
	for i, r := range results {
		pts[i].X = float64(r.NumProcesses)
		pts[i].Y = float64(r.TotalMessages)
	}

	line, points, err := plotter.NewLinePoints(pts)
	if err != nil {
		panic(err)
	}
	line.Color = color.RGBA{R: 255, G: 20, B: 147, A: 255}
	points.Color = color.RGBA{R: 255, G: 20, B: 147, A: 255}

	p.Add(line, points)
	p.Add(plotter.NewGrid())

	if err := p.Save(8*vg.Inch, 6*vg.Inch, "message_count.png"); err != nil {
		panic(err)
	}
}

func generateMemoryOverheadPlot(results []SimulationResult) {
	p := plot.New()
	p.Title.Text = "Vector Clock Memory Overhead vs Number of Processes"
	p.X.Label.Text = "Number of Processes"
	p.Y.Label.Text = "Overhead Ratio (Vector/Lamport)"

	pts := make(plotter.XYs, len(results))
	for i, r := range results {
		pts[i].X = float64(r.NumProcesses)
		pts[i].Y = r.MemoryOverhead
	}

	line, points, err := plotter.NewLinePoints(pts)
	if err != nil {
		panic(err)
	}
	line.Color = color.RGBA{R: 220, G: 20, B: 60, A: 255}
	points.Color = color.RGBA{R: 220, G: 20, B: 60, A: 255}

	p.Add(line, points)
	p.Add(plotter.NewGrid())

	if err := p.Save(8*vg.Inch, 6*vg.Inch, "memory_overhead.png"); err != nil {
		panic(err)
	}
}
