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
	duration        := 10000
	fps             := 60
	frameDuration   := 1000 / fps
	frames          := duration / frameDuration + 1
	workers         := 1
	framesPerWorker := frames / workers

	// Instantiate the progress bar.
	bar := pb.StartNew(framesPerWorker * workers)

	// Create a new Wait Group
	var wg sync.WaitGroup

	// Run all the workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int){
			// Defer calling Done on the WG
			defer wg.Done()
			// Create a wait group for the goroutines that store the frames
			var fwg sync.WaitGroup
			ctx, _ = browser.NewHandler(ctx)
			// Get the start and end frame for this worker
			start := id * framesPerWorker
			end   := start + framesPerWorker - 1
			// Wind the clock of the worker
			for f := 0; f < start; f++ {
				if _, err := browser.GoTo(ctx, f); err != nil {
					panic(err)
				}
			}
			// Start recording the frames
			for f := start; f < end; f++ {
				frame, err := browser.Screenshot(ctx)
				if err != nil {
					panic(err)
				}
				// Start a goroutine to save the frame to the disk
				fwg.Add(1)
				go func(frame []byte, number int) {
					// Defer calling done on the frames wg
					defer fwg.Done()
					path := fmt.Sprintf("/tmp/%06d.png", number)
					if err := ioutil.WriteFile(path, frame, 0644); err != nil {
						panic(err)
					}
					bar.Increment()
				}(frame, f)
			}
			// Wait until all the frames are stored on the disk
			fwg.Wait()
		}(i)
	}

	// Wait for the workers to finish
	wg.Wait()

	// Finish the bar
	bar.Finish()

	return nil
}
