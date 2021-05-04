package web

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/cheggaaa/pb"
)

// StartRecording starts the process of recording a Chrome animation.
func StartRecording() error {
	// Create the browser
	browser := NewBrowser()

	// Open the browser
	ctx, cancel := browser.Open(context.Background())
	defer cancel()

	// Calculate the total number of frames.
	duration := 10000
	fps := 60
	frameDuration := 1000 / fps
	frames := duration / frameDuration + 1
	workers := 4
	chunkSize := frames / workers

	// Open one tab for each worker
	for i := 0; i < workers; i++ {
		ctx, _ = browser.NewHandler(ctx)
	}

	// Instantiate the progress bar.
	bar := pb.StartNew(chunkSize * workers)

	// Create a new Wait Group
	var wg sync.WaitGroup

	// Create the worker
	worker := func(id, start, end int) {
		var wwg sync.WaitGroup

		for f := 0; f < end; f++ {
			if _, err := browser.GoTo(ctx, id, f); err != nil {
				panic(err)
			}
			// Windup until we can start capturing frames.
			if f < start {
				continue
			}
			// Take screenshot
			frame, err := browser.Screenshot(ctx, id)
			if err != nil {
				panic(err)
			}
			// Save screenshot on the disk
			wwg.Add(1)
			go func(frame []byte, f int) {
				if err := ioutil.WriteFile(fmt.Sprintf("/tmp/%06d.png", f), frame, 0644); err != nil {
					panic(err)
				}
				bar.Increment()
				defer wwg.Done()
			}(frame, f)
		}
		// Wait for all the goroutines that save to the disk to finish.
		wwg.Wait()

		fmt.Println("worker", id, "done")
		defer wg.Done()
	}

	for i := 0; i < workers; i++ {
		offset := i * chunkSize
		go worker(i, offset, offset + chunkSize)
		wg.Add(1)
	}

	// Wait for the workers to finish
	wg.Wait()

	// Finish the bar
	bar.Finish()

	return nil
}