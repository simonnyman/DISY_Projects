package main

import (
	"fmt"
	"image/color"
	"time"

	"github.com/simonnyman/DISY_Projects/Synchronization/simulator"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// simulation configuration
const (
	numProcesses   = 10
	simulationTime = 1 * time.Second
	numRuns        = 3 // number of runs per configuration
)

// probability configurations to test
type ProbConfig struct {
	LocalProb float64
	SendProb  float64
	Label     string
}

var probConfigs = []ProbConfig{
	{0.1, 0.1, "Low Activity (10/10)"},
	{0.2, 0.2, "Low-Med (20/20)"},
	{0.3, 0.2, "Med Local/Low Send (30/20)"},
	{0.2, 0.3, "Low Local/Med Send (20/30)"},
	{0.3, 0.3, "Medium (30/30)"},
	{0.4, 0.3, "High Local/Med Send (40/30)"},
	{0.3, 0.4, "Med Local/High Send (30/40)"},
	{0.4, 0.4, "High (40/40)"},
	{0.5, 0.3, "Very High Local (50/30)"},
	{0.3, 0.5, "Very High Send (30/50)"},
	{0.5, 0.5, "Very High Activity (50/50)"},
}

type ProbResult struct {
	Config          ProbConfig
	TotalEvents     int
	LocalEvents     int
	SendEvents      int
	ReceiveEvents   int
	ConcurrentPairs int
	TotalPairs      int
	ConcurrencyRate float64
	LocalRatio      float64 // actual local / total
	SendRatio       float64 // actual send / total
	ReceiveRatio    float64 // actual receive / total
	MessageBalance  float64 // receive / send ratio (should be ~1.0)
}

func main() {
	fmt.Println("Running simulations with varying event probabilities...")
	fmt.Println("This may take a couple minutes...")
	fmt.Println()

	results := runProbabilityAnalysis()

	fmt.Println("Generating plots...")
	generateProbTotalEventsPlot(results)
	generateProbEventTypesPlot(results)
	generateProbConcurrencyPlot(results)
	generateProbDistributionPlot(results)
	generateProbMessageBalancePlot(results)
	generateProbEfficiencyPlot(results)

	fmt.Println("\n✓ All probability analysis plots generated successfully!")
	fmt.Println("\nGenerated files:")
	fmt.Println("  - prob_total_events.png")
	fmt.Println("  - prob_event_types.png")
	fmt.Println("  - prob_concurrency.png")
	fmt.Println("  - prob_distribution.png")
	fmt.Println("  - prob_message_balance.png")
	fmt.Println("  - prob_efficiency.png")
}

func runProbabilityAnalysis() []ProbResult {
	results := make([]ProbResult, 0)

	for idx, config := range probConfigs {
		fmt.Printf("Running simulation %d/%d: %s...\n", idx+1, len(probConfigs), config.Label)

		var avgResult ProbResult
		avgResult.Config = config

		for run := 0; run < numRuns; run++ {
			sim := simulator.NewSimulator(numProcesses)
			sim.RunSimulation(simulationTime, config.LocalProb, config.SendProb)

			stats := sim.GetStatistics()

			totalEvents := stats["total_events"].(int)
			localEvents := stats["local_events"].(int)
			sendEvents := stats["send_events"].(int)
			receiveEvents := stats["receive_events"].(int)

			avgResult.TotalEvents += totalEvents
			avgResult.LocalEvents += localEvents
			avgResult.SendEvents += sendEvents
			avgResult.ReceiveEvents += receiveEvents

			concurrentPairs := sim.CountConcurrentEvents()
			totalPairs := totalEvents * (totalEvents - 1) / 2

			avgResult.ConcurrentPairs += concurrentPairs
			avgResult.TotalPairs += totalPairs
		}

		// Calculate averages
		avgResult.TotalEvents /= numRuns
		avgResult.LocalEvents /= numRuns
		avgResult.SendEvents /= numRuns
		avgResult.ReceiveEvents /= numRuns
		avgResult.ConcurrentPairs /= numRuns
		avgResult.TotalPairs /= numRuns

		// Calculate ratios
		if avgResult.TotalEvents > 0 {
			avgResult.LocalRatio = float64(avgResult.LocalEvents) / float64(avgResult.TotalEvents) * 100
			avgResult.SendRatio = float64(avgResult.SendEvents) / float64(avgResult.TotalEvents) * 100
			avgResult.ReceiveRatio = float64(avgResult.ReceiveEvents) / float64(avgResult.TotalEvents) * 100
		}

		if avgResult.TotalPairs > 0 {
			avgResult.ConcurrencyRate = float64(avgResult.ConcurrentPairs) / float64(avgResult.TotalPairs) * 100
		}

		if avgResult.SendEvents > 0 {
			avgResult.MessageBalance = float64(avgResult.ReceiveEvents) / float64(avgResult.SendEvents)
		}

		results = append(results, avgResult)
	}

	return results
}

func generateProbTotalEventsPlot(results []ProbResult) {
	p := plot.New()
	p.Title.Text = "Total Events by Probability Configuration"
	p.Y.Label.Text = "Total Events"
	p.NominalX("") // Will use labels

	bars := make(plotter.Values, len(results))
	labels := make([]string, len(results))

	for i, r := range results {
		bars[i] = float64(r.TotalEvents)
		labels[i] = fmt.Sprintf("%.1f/%.1f", r.Config.LocalProb*100, r.Config.SendProb*100)
	}

	barChart, err := plotter.NewBarChart(bars, vg.Points(20))
	if err != nil {
		panic(err)
	}
	barChart.Color = color.RGBA{R: 100, G: 149, B: 237, A: 255}

	p.Add(barChart)
	p.NominalX(labels...)
	p.X.Label.Text = "Probabilities (Local%/Send%)"
	p.X.Tick.Label.Rotation = 0.5

	if err := p.Save(10*vg.Inch, 6*vg.Inch, "prob_total_events.png"); err != nil {
		panic(err)
	}
}

func generateProbEventTypesPlot(results []ProbResult) {
	p := plot.New()
	p.Title.Text = "Event Type Breakdown by Configuration"
	p.Y.Label.Text = "Event Count"
	p.Legend.Top = true
	p.Legend.Left = true

	localVals := make(plotter.Values, len(results))
	sendVals := make(plotter.Values, len(results))
	receiveVals := make(plotter.Values, len(results))
	labels := make([]string, len(results))

	for i, r := range results {
		localVals[i] = float64(r.LocalEvents)
		sendVals[i] = float64(r.SendEvents)
		receiveVals[i] = float64(r.ReceiveEvents)
		labels[i] = fmt.Sprintf("%.1f/%.1f", r.Config.LocalProb*100, r.Config.SendProb*100)
	}

	width := vg.Points(15)

	localBars, _ := plotter.NewBarChart(localVals, width)
	localBars.Color = color.RGBA{R: 0, G: 200, B: 0, A: 255}
	localBars.Offset = -width

	sendBars, _ := plotter.NewBarChart(sendVals, width)
	sendBars.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}

	receiveBars, _ := plotter.NewBarChart(receiveVals, width)
	receiveBars.Color = color.RGBA{R: 0, G: 0, B: 255, A: 255}
	receiveBars.Offset = width

	p.Add(localBars, sendBars, receiveBars)
	p.Legend.Add("Local", localBars)
	p.Legend.Add("Send", sendBars)
	p.Legend.Add("Receive", receiveBars)
	p.NominalX(labels...)
	p.X.Label.Text = "Probabilities (Local%/Send%)"
	p.X.Tick.Label.Rotation = 0.5

	if err := p.Save(10*vg.Inch, 6*vg.Inch, "prob_event_types.png"); err != nil {
		panic(err)
	}
}

func generateProbConcurrencyPlot(results []ProbResult) {
	p := plot.New()
	p.Title.Text = "Concurrency Rate by Configuration"
	p.Y.Label.Text = "Concurrency Rate (%)"

	bars := make(plotter.Values, len(results))
	labels := make([]string, len(results))

	for i, r := range results {
		bars[i] = r.ConcurrencyRate
		labels[i] = fmt.Sprintf("%.1f/%.1f", r.Config.LocalProb*100, r.Config.SendProb*100)
	}

	barChart, err := plotter.NewBarChart(bars, vg.Points(20))
	if err != nil {
		panic(err)
	}
	barChart.Color = color.RGBA{R: 147, G: 112, B: 219, A: 255}

	p.Add(barChart)
	p.NominalX(labels...)
	p.X.Label.Text = "Probabilities (Local%/Send%)"
	p.X.Tick.Label.Rotation = 0.5

	if err := p.Save(10*vg.Inch, 6*vg.Inch, "prob_concurrency.png"); err != nil {
		panic(err)
	}
}

func generateProbDistributionPlot(results []ProbResult) {
	p := plot.New()
	p.Title.Text = "Actual Event Distribution (% of Total)"
	p.Y.Label.Text = "Percentage of Total Events"
	p.Legend.Top = true

	localVals := make(plotter.Values, len(results))
	sendVals := make(plotter.Values, len(results))
	receiveVals := make(plotter.Values, len(results))
	labels := make([]string, len(results))

	for i, r := range results {
		localVals[i] = r.LocalRatio
		sendVals[i] = r.SendRatio
		receiveVals[i] = r.ReceiveRatio
		labels[i] = fmt.Sprintf("%.1f/%.1f", r.Config.LocalProb*100, r.Config.SendProb*100)
	}

	width := vg.Points(15)

	localBars, _ := plotter.NewBarChart(localVals, width)
	localBars.Color = color.RGBA{R: 50, G: 205, B: 50, A: 255}
	localBars.Offset = -width

	sendBars, _ := plotter.NewBarChart(sendVals, width)
	sendBars.Color = color.RGBA{R: 255, G: 99, B: 71, A: 255}

	receiveBars, _ := plotter.NewBarChart(receiveVals, width)
	receiveBars.Color = color.RGBA{R: 30, G: 144, B: 255, A: 255}
	receiveBars.Offset = width

	p.Add(localBars, sendBars, receiveBars)
	p.Legend.Add("Local %", localBars)
	p.Legend.Add("Send %", sendBars)
	p.Legend.Add("Receive %", receiveBars)
	p.NominalX(labels...)
	p.X.Label.Text = "Probabilities (Local%/Send%)"
	p.X.Tick.Label.Rotation = 0.5

	if err := p.Save(10*vg.Inch, 6*vg.Inch, "prob_distribution.png"); err != nil {
		panic(err)
	}
}

func generateProbMessageBalancePlot(results []ProbResult) {
	p := plot.New()
	p.Title.Text = "Message Balance: Receive/Send Ratio"
	p.Y.Label.Text = "Receive/Send Ratio (should be ~1.0)"

	bars := make(plotter.Values, len(results))
	labels := make([]string, len(results))

	for i, r := range results {
		bars[i] = r.MessageBalance
		labels[i] = fmt.Sprintf("%.1f/%.1f", r.Config.LocalProb*100, r.Config.SendProb*100)
	}

	barChart, err := plotter.NewBarChart(bars, vg.Points(20))
	if err != nil {
		panic(err)
	}
	barChart.Color = color.RGBA{R: 255, G: 140, B: 0, A: 255}

	// Add reference line at 1.0
	refLine := plotter.NewFunction(func(x float64) float64 { return 1.0 })
	refLine.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	refLine.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}

	p.Add(barChart, refLine)
	p.NominalX(labels...)
	p.X.Label.Text = "Probabilities (Local%/Send%)"
	p.X.Tick.Label.Rotation = 0.5

	if err := p.Save(10*vg.Inch, 6*vg.Inch, "prob_message_balance.png"); err != nil {
		panic(err)
	}
}

func generateProbEfficiencyPlot(results []ProbResult) {
	p := plot.New()
	p.Title.Text = "Communication Efficiency by Configuration"
	p.Y.Label.Text = "Events"
	p.Legend.Top = true

	// Show relationship between local and communication events
	totalVals := make(plotter.Values, len(results))
	localVals := make(plotter.Values, len(results))
	commVals := make(plotter.Values, len(results))
	labels := make([]string, len(results))

	for i, r := range results {
		totalVals[i] = float64(r.TotalEvents)
		localVals[i] = float64(r.LocalEvents)
		commVals[i] = float64(r.SendEvents + r.ReceiveEvents)
		labels[i] = fmt.Sprintf("%.1f/%.1f", r.Config.LocalProb*100, r.Config.SendProb*100)
	}

	width := vg.Points(20)

	totalBars, _ := plotter.NewBarChart(totalVals, width)
	totalBars.Color = color.RGBA{R: 70, G: 130, B: 180, A: 255}
	totalBars.Offset = -width / 2

	localBars, _ := plotter.NewBarChart(localVals, width)
	localBars.Color = color.RGBA{R: 34, G: 139, B: 34, A: 255}
	localBars.Offset = width / 2

	commBars, _ := plotter.NewBarChart(commVals, width)
	commBars.Color = color.RGBA{R: 220, G: 20, B: 60, A: 255}
	commBars.Offset = width * 1.5

	p.Add(totalBars, localBars, commBars)
	p.Legend.Add("Total Events", totalBars)
	p.Legend.Add("Local Events", localBars)
	p.Legend.Add("Communication Events", commBars)
	p.NominalX(labels...)
	p.X.Label.Text = "Probabilities (Local%/Send%)"
	p.X.Tick.Label.Rotation = 0.5

	if err := p.Save(10*vg.Inch, 6*vg.Inch, "prob_efficiency.png"); err != nil {
		panic(err)
	}
}
