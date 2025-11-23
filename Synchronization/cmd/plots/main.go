package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"time"

	"github.com/simonnyman/DISY_Projects/Synchronization/simulator"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// simulation configuration
const (
	simulationTime = 1 * time.Second
	localEventProb = 0.3
	sendEventProb  = 0.4
	numRuns        = 3 // number of runs per scenario (for averaging)
)

// scenarios to test (varying number of processes)
var scenarios = []struct {
	NumProcesses int
	Label        string
}{
	{2, "2 Proc"},
	{5, "5 Proc"},
	{10, "10 Proc"},
	{15, "15 Proc"},
	{20, "20 Proc"},
	{30, "30 Proc"},
}

type SimulationResult struct {
	Label               string
	NumProcesses        int
	TotalEvents         int
	LocalEvents         int
	SendEvents          int
	ReceiveEvents       int
	ConcurrentPairs     int // concurrent relationships
	TotalPairs          int
	LamportBytesPerProc int // Space overhead per process
	VectorBytesPerProc  int // Space overhead per process
	LamportMsgBytes     int // Average message overhead
	VectorMsgBytes      int // Average message overhead
	TotalMessages       int
	ConcurrencyRate     float64
}

func main() {
	fmt.Println("Running simulations to compare Lamport vs Vector clocks...")
	fmt.Println()

	results := runSimulations()

	fmt.Println("Generating individual plots...")
	generateEventStatisticsPlot(results)
	generateSpaceOverheadPlot(results)
	generateMessageOverheadPlot(results)
	generateConcurrencyCausalityPlot(results)

	fmt.Println("\nCombining plots into 2x2 grid...")
	combinePlots()

	fmt.Println("\n✓ All plots generated successfully!")
	fmt.Println("\nGenerated files (in plot_pictures/):")
}

func runSimulations() []SimulationResult {
	results := make([]SimulationResult, 0)

	for _, scenario := range scenarios {
		fmt.Printf("Running simulation: %s...\n", scenario.Label)

		// Run multiple times and average results
		var avgResult SimulationResult
		avgResult.Label = scenario.Label
		avgResult.NumProcesses = scenario.NumProcesses

		for run := 0; run < numRuns; run++ {
			sim := simulator.NewSimulator(scenario.NumProcesses)
			sim.RunSimulation(simulationTime, localEventProb, sendEventProb)

			stats := sim.GetStatistics()
			metrics := sim.AnalyzeComplexity()

			avgResult.TotalEvents += stats["total_events"].(int)
			avgResult.LocalEvents += stats["local_events"].(int)
			avgResult.SendEvents += stats["send_events"].(int)
			avgResult.ReceiveEvents += stats["receive_events"].(int)

			// Count concurrent events
			concurrentPairs := sim.CountConcurrentEvents()
			totalEvents := stats["total_events"].(int)
			totalPairs := totalEvents * (totalEvents - 1) / 2

			avgResult.ConcurrentPairs += concurrentPairs
			avgResult.TotalPairs += totalPairs

			// Space overhead per process
			avgResult.LamportBytesPerProc += metrics.LamportClockSize
			avgResult.VectorBytesPerProc += metrics.VectorClockSize

			// Message overhead (timestamp data in each message)
			avgResult.LamportMsgBytes += 8                          // Lamport: just the int64 timestamp
			avgResult.VectorMsgBytes += (8 * scenario.NumProcesses) // Vector: full vector

			avgResult.TotalMessages += metrics.TotalMessages
		}

		// Calculate averages
		avgResult.TotalEvents /= numRuns
		avgResult.LocalEvents /= numRuns
		avgResult.SendEvents /= numRuns
		avgResult.ReceiveEvents /= numRuns
		avgResult.ConcurrentPairs /= numRuns
		avgResult.TotalPairs /= numRuns
		avgResult.LamportBytesPerProc /= numRuns
		avgResult.VectorBytesPerProc /= numRuns
		avgResult.LamportMsgBytes /= numRuns
		avgResult.VectorMsgBytes /= numRuns
		avgResult.TotalMessages /= numRuns

		if avgResult.TotalPairs > 0 {
			avgResult.ConcurrencyRate = float64(avgResult.ConcurrentPairs) / float64(avgResult.TotalPairs) * 100
		}

		results = append(results, avgResult)
	}

	return results
}

// Plot 1: Event Statistics (Workload Overview)
func generateEventStatisticsPlot(results []SimulationResult) {
	p := plot.New()
	p.Title.Text = "Plot 1: Event Statistics (Workload Overview)"
	p.Y.Label.Text = "Event Count"
	p.Legend.Top = true

	localVals := make(plotter.Values, len(results))
	sendVals := make(plotter.Values, len(results))
	receiveVals := make(plotter.Values, len(results))
	labels := make([]string, len(results))

	for i, r := range results {
		localVals[i] = float64(r.LocalEvents)
		sendVals[i] = float64(r.SendEvents)
		receiveVals[i] = float64(r.ReceiveEvents)
		labels[i] = r.Label
	}

	width := vg.Points(20)

	// Create stacked bar chart effect
	localBars, _ := plotter.NewBarChart(localVals, width)
	localBars.Color = color.RGBA{R: 76, G: 175, B: 80, A: 255}
	localBars.Offset = -width

	sendBars, _ := plotter.NewBarChart(sendVals, width)
	sendBars.Color = color.RGBA{R: 244, G: 67, B: 54, A: 255}

	receiveBars, _ := plotter.NewBarChart(receiveVals, width)
	receiveBars.Color = color.RGBA{R: 33, G: 150, B: 243, A: 255}
	receiveBars.Offset = width

	p.Add(localBars, sendBars, receiveBars)
	p.Legend.Add("Local Events", localBars)
	p.Legend.Add("Send Events", sendBars)
	p.Legend.Add("Receive Events", receiveBars)
	p.NominalX(labels...)
	p.X.Label.Text = "Scenario"

	if err := p.Save(8*vg.Inch, 6*vg.Inch, "plot_pictures/1_event_statistics.png"); err != nil {
		panic(err)
	}
}

// Plot 2: Space Overhead vs Number of Processes
func generateSpaceOverheadPlot(results []SimulationResult) {
	p := plot.New()
	p.Title.Text = "Plot 2: Space Overhead (Bytes per Process)"
	p.X.Label.Text = "Number of Processes"
	p.Y.Label.Text = "Timestamp Size (bytes)"
	p.Legend.Top = true

	lamportPts := make(plotter.XYs, len(results))
	vectorPts := make(plotter.XYs, len(results))

	for i, r := range results {
		lamportPts[i].X = float64(r.NumProcesses)
		lamportPts[i].Y = float64(r.LamportBytesPerProc)
		vectorPts[i].X = float64(r.NumProcesses)
		vectorPts[i].Y = float64(r.VectorBytesPerProc)
	}

	lamportLine, lamportPoints, _ := plotter.NewLinePoints(lamportPts)
	lamportLine.Color = color.RGBA{R: 255, G: 152, B: 0, A: 255}
	lamportLine.Width = vg.Points(2)
	lamportPoints.Color = color.RGBA{R: 255, G: 152, B: 0, A: 255}
	lamportPoints.Radius = vg.Points(4)

	vectorLine, vectorPoints, _ := plotter.NewLinePoints(vectorPts)
	vectorLine.Color = color.RGBA{R: 156, G: 39, B: 176, A: 255}
	vectorLine.Width = vg.Points(2)
	vectorPoints.Color = color.RGBA{R: 156, G: 39, B: 176, A: 255}
	vectorPoints.Radius = vg.Points(4)

	p.Add(lamportLine, lamportPoints, vectorLine, vectorPoints)
	p.Add(plotter.NewGrid())
	p.Legend.Add("Lamport (O(1) - constant)", lamportLine, lamportPoints)
	p.Legend.Add("Vector (O(n) - linear)", vectorLine, vectorPoints)

	if err := p.Save(8*vg.Inch, 6*vg.Inch, "plot_pictures/2_space_overhead.png"); err != nil {
		panic(err)
	}
}

// Plot 3: Time Complexity vs Number of Processes
func generateMessageOverheadPlot(results []SimulationResult) {
	p := plot.New()
	p.Title.Text = "Plot 3: Time Complexity vs Number of Processes"
	p.X.Label.Text = "Number of Processes"
	p.Y.Label.Text = "Operations per Event"
	p.Legend.Top = true

	lamportPts := make(plotter.XYs, len(results))
	vectorPts := make(plotter.XYs, len(results))

	for i, r := range results {
		// Operations per event: for updates and comparisons
		// Lamport: 1 operation per update/compare (O(1))
		// Vector: n operations per update/compare (O(n))
		lamportOpsPerEvent := 1.0
		vectorOpsPerEvent := float64(r.NumProcesses)

		lamportPts[i].X = float64(r.NumProcesses)
		lamportPts[i].Y = lamportOpsPerEvent
		vectorPts[i].X = float64(r.NumProcesses)
		vectorPts[i].Y = vectorOpsPerEvent
	}

	lamportLine, lamportPoints, _ := plotter.NewLinePoints(lamportPts)
	lamportLine.Color = color.RGBA{R: 255, G: 152, B: 0, A: 255}
	lamportLine.Width = vg.Points(2)
	lamportPoints.Color = color.RGBA{R: 255, G: 152, B: 0, A: 255}
	lamportPoints.Radius = vg.Points(4)

	vectorLine, vectorPoints, _ := plotter.NewLinePoints(vectorPts)
	vectorLine.Color = color.RGBA{R: 156, G: 39, B: 176, A: 255}
	vectorLine.Width = vg.Points(2)
	vectorPoints.Color = color.RGBA{R: 156, G: 39, B: 176, A: 255}
	vectorPoints.Radius = vg.Points(4)

	p.Add(lamportLine, lamportPoints, vectorLine, vectorPoints)
	p.Add(plotter.NewGrid())
	p.Legend.Add("Lamport: O(1) constant", lamportLine, lamportPoints)
	p.Legend.Add("Vector: O(n) linear", vectorLine, vectorPoints)

	if err := p.Save(8*vg.Inch, 6*vg.Inch, "plot_pictures/3_time_complexity.png"); err != nil {
		panic(err)
	}
} // Plot 4: Concurrency & Causality Breakdown
func generateConcurrencyCausalityPlot(results []SimulationResult) {
	p := plot.New()
	p.Title.Text = "Plot 4: Correctness Trade-off: Concurrency Detected by Each Algorithm"
	p.Y.Label.Text = "Concurrent Event Pairs (Percentage)"
	p.X.Label.Text = "Scenario"
	p.Legend.Top = true

	lamportConcurrent := make(plotter.Values, len(results))
	vectorConcurrent := make(plotter.Values, len(results))
	labels := make([]string, len(results))

	for i, r := range results {
		lamportConcurrent[i] = 0.0

		vectorConcurrent[i] = r.ConcurrencyRate
		labels[i] = r.Label
	}

	width := vg.Points(25)

	lamportBars, _ := plotter.NewBarChart(lamportConcurrent, width)
	lamportBars.Color = color.RGBA{R: 255, G: 152, B: 0, A: 255} // Orange
	lamportBars.Offset = -width / 2

	vectorBars, _ := plotter.NewBarChart(vectorConcurrent, width)
	vectorBars.Color = color.RGBA{R: 156, G: 39, B: 176, A: 255} // Purple
	vectorBars.Offset = width / 2

	p.Add(lamportBars, vectorBars)
	p.Legend.Add("Lamport: Concurrent (≈0%)", lamportBars)
	p.Legend.Add("Vector: Concurrent", vectorBars)
	p.NominalX(labels...)

	if err := p.Save(8*vg.Inch, 6*vg.Inch, "plot_pictures/4_concurrency_causality.png"); err != nil {
		panic(err)
	}
}

// combinePlots combines the 4 individual plots into a 2x2 grid
func combinePlots() {
	// Define the 4 plots to combine (2x2 grid of trade-off analysis)
	plotFiles := []string{
		"plot_pictures/1_event_statistics.png",
		"plot_pictures/2_space_overhead.png",
		"plot_pictures/3_time_complexity.png",
		"plot_pictures/4_concurrency_causality.png",
	}

	// Create a 2x2 grid
	rows := 2
	cols := 2

	// Load all images
	images := make([]image.Image, len(plotFiles))
	for i, file := range plotFiles {
		f, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		img, err := png.Decode(f)
		if err != nil {
			panic(err)
		}
		images[i] = img
	}

	// Get dimensions from first image
	imgWidth := images[0].Bounds().Dx()
	imgHeight := images[0].Bounds().Dy()

	// Create combined image
	combined := image.NewRGBA(image.Rect(0, 0, imgWidth*cols, imgHeight*rows))

	// Draw images in 2x2 grid
	positions := []image.Point{
		{0, 0},                // Top-left
		{imgWidth, 0},         // Top-right
		{0, imgHeight},        // Bottom-left
		{imgWidth, imgHeight}, // Bottom-right
	}

	for i, img := range images {
		draw.Draw(combined, image.Rectangle{
			Min: positions[i],
			Max: positions[i].Add(image.Point{imgWidth, imgHeight}),
		}, img, image.Point{0, 0}, draw.Src)
	}

	// Save the combined image
	outFile, err := os.Create("plot_pictures/lamport_vs_vector_tradeoffs.png")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, combined); err != nil {
		panic(err)
	}

	fmt.Println("  Layout (2x2 grid):")
	fmt.Println("    ┌──────────────────────────┬──────────────────────────┐")
	fmt.Println("    │ 1. Event Statistics      │ 2. Space Overhead        │")
	fmt.Println("    │    (Workload)            │    (O(1) vs O(n))        │")
	fmt.Println("    ├──────────────────────────┼──────────────────────────┤")
	fmt.Println("    │ 3. Time Complexity       │ 4. Concurrency Detection │")
	fmt.Println("    │    (O(1) vs O(n))        │    (Correctness)         │")
	fmt.Println("    └──────────────────────────┴──────────────────────────┘")
}
