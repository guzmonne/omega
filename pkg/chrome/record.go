package chrome

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/cheggaaa/pb"
)

// Record starts the process of recording a Chrome animation.
func Record() error {
	// Create a canceable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup a Ctrl+C handler
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func(){
		<- signals
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		cancel()
		os.Exit(0)
	}()

	// Start the web server on a different goroutine
	webServerOptions := NewWebServerOptions()
	go Serve(webServerOptions)

	// Calculate the total number of frames.
	duration        := 10000
	fps             := 60
	frameDuration   := 1000 / fps
	frames          := duration / frameDuration + 1
	workers         := 4
	framesPerWorker := frames / workers

	// Instantiate the progress bar.
	bar := pb.StartNew(framesPerWorker * workers)

	// Create the handler url
	urlstr := fmt.Sprintf("http://localhost:%d/handler", webServerOptions.Port)

	// Create a new Wait Group
	var wg sync.WaitGroup

	// Run all the workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int){
			// Create a browser
			browser := NewBrowser()
			// Open the browser
			ctx, cancel := browser.Open(ctx, urlstr)
			// Defer calling Done on the WG and closing the browser
			defer func(){ wg.Done(); cancel() }()
			// Get the start and end frame for this worker
			start := id * framesPerWorker
			end   := start + framesPerWorker
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
				// Save the frame to disk
				if err := ioutil.WriteFile(fmt.Sprintf("/tmp/%06d.png", f), frame, 0644); err != nil {
					panic(err)
				}
				// Move the vt forward
				if _, err := browser.GoTo(ctx, f); err != nil {
					panic(err)
				}
				// Update the bar
				bar.Increment()
			}
		}(i)
	}

	// Wait for the workers to finish
	wg.Wait()

	// Finish the bar
	bar.Finish()

	return nil
}
