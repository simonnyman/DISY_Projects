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
	numProcesses   = 10 // fixed number of processes
	localEventProb = 0.3
	sendEventProb  = 0.4
	numRuns        = 3 // number of runs per duration (for averaging)
)

// simulation durations to test (in milliseconds)
var simulationDurations = []int{100, 250, 500, 750, 1000, 1500, 2000, 3000, 5000}

type TimeResult struct {
	DurationMS      int
	TotalEvents     int
	LocalEvents     int
	SendEvents      int
	ReceiveEvents   int
	ConcurrentPairs int
	EventRate       float64 // events per second
	MessageRate     float64 // messages per second
	ConcurrencyRate float64 // percentage
	AvgLamportValue float64
	MaxLamportValue int64
	AvgVectorSum    float64
	MaxVectorSum    int64
}

func main() {
	fmt.Println("Running simulations with varying simulation times...")
	fmt.Println("This may take a minute...")
	fmt.Println()

	results := runTimeAnalysis()

	fmt.Println("Generating plots...")
	generateEventGrowthPlot(results)
	generateEventRatePlot(results)
	generateEventTypeTimePlot(results)
	generateConcurrencyTimePlot(results)
	generateClockValuePlot(results)
	generateEfficiencyPlot(results)

	fmt.Println("\n✓ All time analysis plots generated successfully!")
	fmt.Println("\nGenerated files:")
	fmt.Println("  - time_event_growth.png")
	fmt.Println("  - time_event_rate.png")
	fmt.Println("  - time_event_types.png")
	fmt.Println("  - time_concurrency.png")
	fmt.Println("  - time_clock_values.png")
	fmt.Println("  - time_efficiency.png")
}

func runTimeAnalysis() []TimeResult {
	results := make([]TimeResult, 0)

	for _, durationMS := range simulationDurations {
		duration := time.Duration(durationMS) * time.Millisecond
		fmt.Printf("Running simulation for %d ms...\n", durationMS)

		// Run multiple times and average results
		var avgResult TimeResult
		avgResult.DurationMS = durationMS

		for run := 0; run < numRuns; run++ {
			sim := simulator.NewSimulator(numProcesses)
			sim.RunSimulation(duration, localEventProb, sendEventProb)

			stats := sim.GetStatistics()

			totalEvents := stats["total_events"].(int)
			avgResult.TotalEvents += totalEvents
			avgResult.LocalEvents += stats["local_events"].(int)
			avgResult.SendEvents += stats["send_events"].(int)
			avgResult.ReceiveEvents += stats["receive_events"].(int)

			concurrentPairs := sim.CountConcurrentEvents()
			totalPairs := totalEvents * (totalEvents - 1) / 2
			avgResult.ConcurrentPairs += concurrentPairs

			// Calculate clock statistics
			var lamportSum int64
			var maxLamport int64
			var vectorSum int64
			var maxVector int64

			for _, process := range sim.Processes {
				if len(process.Events) > 0 {
					lastEvent := process.Events[len(process.Events)-1]
					lamportSum += lastEvent.Timestamp
					if lastEvent.Timestamp > maxLamport {
						maxLamport = lastEvent.Timestamp
					}

					vecSum := int64(0)
					for _, v := range lastEvent.VectorTime {
						vecSum += v
					}
					vectorSum += vecSum
					if vecSum > maxVector {
						maxVector = vecSum
					}
				}
			}

			avgResult.AvgLamportValue += float64(lamportSum) / float64(numProcesses)
			avgResult.MaxLamportValue += maxLamport
			avgResult.AvgVectorSum += float64(vectorSum) / float64(numProcesses)
			avgResult.MaxVectorSum += maxVector

			if totalPairs > 0 {
				avgResult.ConcurrencyRate += float64(concurrentPairs) / float64(totalPairs) * 100
			}
		}

		// Calculate averages
		avgResult.TotalEvents /= numRuns
		avgResult.LocalEvents /= numRuns
		avgResult.SendEvents /= numRuns
		avgResult.ReceiveEvents /= numRuns
		avgResult.ConcurrentPairs /= numRuns
		avgResult.ConcurrencyRate /= float64(numRuns)
		avgResult.AvgLamportValue /= float64(numRuns)
		avgResult.MaxLamportValue /= int64(numRuns)
		avgResult.AvgVectorSum /= float64(numRuns)
		avgResult.MaxVectorSum /= int64(numRuns)

		// Calculate rates (per second)
		durationSeconds := float64(durationMS) / 1000.0
		avgResult.EventRate = float64(avgResult.TotalEvents) / durationSeconds
		avgResult.MessageRate = float64(avgResult.SendEvents) / durationSeconds

		results = append(results, avgResult)
	}

	return results
}

func generateEventGrowthPlot(results []TimeResult) {
	p := plot.New()
	p.Title.Text = "Total Events vs Simulation Duration"
	p.X.Label.Text = "Simulation Duration (ms)"
	p.Y.Label.Text = "Total Events"

	pts := make(plotter.XYs, len(results))
	for i, r := range results {
		pts[i].X = float64(r.DurationMS)
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

	if err := p.Save(8*vg.Inch, 6*vg.Inch, "time_event_growth.png"); err != nil {
		panic(err)
	}
}

func generateEventRatePlot(results []TimeResult) {
	p := plot.New()
	p.Title.Text = "Event Generation Rate vs Simulation Duration"
	p.X.Label.Text = "Simulation Duration (ms)"
	p.Y.Label.Text = "Events per Second"

	pts := make(plotter.XYs, len(results))
	for i, r := range results {
		pts[i].X = float64(r.DurationMS)
		pts[i].Y = r.EventRate
	}

	line, points, err := plotter.NewLinePoints(pts)
	if err != nil {
		panic(err)
	}
	line.Color = color.RGBA{R: 0, G: 200, B: 0, A: 255}
	points.Color = color.RGBA{R: 0, G: 200, B: 0, A: 255}
	points.Shape = plotutil.DefaultGlyphShapes[0]

	p.Add(line, points)
	p.Add(plotter.NewGrid())

	if err := p.Save(8*vg.Inch, 6*vg.Inch, "time_event_rate.png"); err != nil {
		panic(err)
	}
}

func generateEventTypeTimePlot(results []TimeResult) {
	p := plot.New()
	p.Title.Text = "Event Types vs Simulation Duration"
	p.X.Label.Text = "Simulation Duration (ms)"
	p.Y.Label.Text = "Event Count"
	p.Legend.Top = true

	localPts := make(plotter.XYs, len(results))
	sendPts := make(plotter.XYs, len(results))
	receivePts := make(plotter.XYs, len(results))

	for i, r := range results {
		localPts[i].X = float64(r.DurationMS)
		localPts[i].Y = float64(r.LocalEvents)
		sendPts[i].X = float64(r.DurationMS)
		sendPts[i].Y = float64(r.SendEvents)
		receivePts[i].X = float64(r.DurationMS)
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

	if err := p.Save(8*vg.Inch, 6*vg.Inch, "time_event_types.png"); err != nil {
		panic(err)
	}
}

func generateConcurrencyTimePlot(results []TimeResult) {
	p := plot.New()
	p.Title.Text = "Concurrency Rate vs Simulation Duration"
	p.X.Label.Text = "Simulation Duration (ms)"
	p.Y.Label.Text = "Concurrency Rate (%)"

	pts := make(plotter.XYs, len(results))
	for i, r := range results {
		pts[i].X = float64(r.DurationMS)
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

	if err := p.Save(8*vg.Inch, 6*vg.Inch, "time_concurrency.png"); err != nil {
		panic(err)
	}
}

func generateClockValuePlot(results []TimeResult) {
	p := plot.New()
	p.Title.Text = "Average Clock Values vs Simulation Duration"
	p.X.Label.Text = "Simulation Duration (ms)"
	p.Y.Label.Text = "Average Clock Value"
	p.Legend.Top = true

	lamportPts := make(plotter.XYs, len(results))
	vectorPts := make(plotter.XYs, len(results))

	for i, r := range results {
		lamportPts[i].X = float64(r.DurationMS)
		lamportPts[i].Y = r.AvgLamportValue
		vectorPts[i].X = float64(r.DurationMS)
		vectorPts[i].Y = r.AvgVectorSum
	}

	lamportLine, lamportPoints, _ := plotter.NewLinePoints(lamportPts)
	lamportLine.Color = color.RGBA{R: 255, G: 165, B: 0, A: 255}
	lamportPoints.Color = color.RGBA{R: 255, G: 165, B: 0, A: 255}

	vectorLine, vectorPoints, _ := plotter.NewLinePoints(vectorPts)
	vectorLine.Color = color.RGBA{R: 0, G: 128, B: 255, A: 255}
	vectorPoints.Color = color.RGBA{R: 0, G: 128, B: 255, A: 255}

	p.Add(lamportLine, lamportPoints, vectorLine, vectorPoints)
	p.Add(plotter.NewGrid())
	p.Legend.Add("Avg Lamport Clock", lamportLine, lamportPoints)
	p.Legend.Add("Avg Vector Clock Sum", vectorLine, vectorPoints)

	if err := p.Save(8*vg.Inch, 6*vg.Inch, "time_clock_values.png"); err != nil {
		panic(err)
	}
}

func generateEfficiencyPlot(results []TimeResult) {
	p := plot.New()
	p.Title.Text = "System Efficiency: Events per Time Unit"
	p.X.Label.Text = "Simulation Duration (ms)"
	p.Y.Label.Text = "Events per 100ms"
	p.Legend.Top = true

	totalPts := make(plotter.XYs, len(results))
	messagePts := make(plotter.XYs, len(results))

	for i, r := range results {
		// Normalize to events per 100ms for easier interpretation
		totalPts[i].X = float64(r.DurationMS)
		totalPts[i].Y = (float64(r.TotalEvents) / float64(r.DurationMS)) * 100
		messagePts[i].X = float64(r.DurationMS)
		messagePts[i].Y = (float64(r.SendEvents) / float64(r.DurationMS)) * 100
	}

	totalLine, totalPoints, _ := plotter.NewLinePoints(totalPts)
	totalLine.Color = color.RGBA{R: 70, G: 130, B: 180, A: 255}
	totalPoints.Color = color.RGBA{R: 70, G: 130, B: 180, A: 255}

	msgLine, msgPoints, _ := plotter.NewLinePoints(messagePts)
	msgLine.Color = color.RGBA{R: 220, G: 20, B: 60, A: 255}
	msgPoints.Color = color.RGBA{R: 220, G: 20, B: 60, A: 255}

	p.Add(totalLine, totalPoints, msgLine, msgPoints)
	p.Add(plotter.NewGrid())
	p.Legend.Add("Total Event Rate", totalLine, totalPoints)
	p.Legend.Add("Message Rate", msgLine, msgPoints)

	if err := p.Save(8*vg.Inch, 6*vg.Inch, "time_efficiency.png"); err != nil {
		panic(err)
	}
}
