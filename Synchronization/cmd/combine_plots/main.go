package main

import (
	"image"
	"image/draw"
	"image/png"
	"os"
)

func main() {
	// Define the 4 plots to combine
	plotFiles := []string{
		"plotpictures/concurrency_rate.png",
		"plotpictures/memory_usage.png",
		"plotpictures/prob_total_events.png",
		"plotpictures/time_clock_values.png",
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
	outFile, err := os.Create("plotpictures/combined_2x2.png")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, combined); err != nil {
		panic(err)
	}

	println("✓ Combined plot created: plotpictures/combined_2x2.png")
	println("\nLayout:")
	println("  ┌─────────────────────┬─────────────────────┐")
	println("  │ concurrency_rate    │ memory_usage        │")
	println("  ├─────────────────────┼─────────────────────┤")
	println("  │ prob_total_events   │ time_clock_values   │")
	println("  └─────────────────────┴─────────────────────┘")
}
