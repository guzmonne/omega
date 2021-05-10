package chrome

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"

	"github.com/cheggaaa/pb"
	"github.com/chromedp/cdproto/page"
	"gux.codes/omega/pkg/browser"
	"gux.codes/omega/pkg/utils"
)

// RecordParams specifies how the recording should be done.
type RecordParams struct {
	// Duration to recording.
	Duration int
	// FPS determines the frames per second of the recording.
	FPS int
	// URL specifies the URL to be recorded.
	URL string
	// Interface used to write the frames to disk.
	Writer utils.Writer
	// Workers used for the recording.
	Workers int
	// Viewport params.
	Viewport page.Viewport
}

// Record starts the process of recording a Chrome animation.
func Record(params RecordParams) error {
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

	// Override the provided RecordParams
	params.Writer = utils.FWriter
	params.URL = fmt.Sprintf("http://localhost:%d/handler", webServerOptions.Port)

	// Start the browsers and start recording each session
	if err := record(ctx, params); err != nil {
		return err
	}

	return nil
}

func record(parent context.Context, params RecordParams) error {
	// Store all the available workers
	availableWorkers := params.Workers

	// Calculate the amount of frames ro record.
	frameDuration   := 1000 / float64(params.FPS)
	framesToRecord  := params.Duration * params.FPS / 1000

	// Create a channel to communicate when a goroutine is ready to do more work
	// providing an appropiate context.
	routineReady := make(chan context.Context)

	// Create a map that holds the frames that each browser has last processed
	previousFrames := make(map[context.Context]int)

	// Instantiate the progress bar.
	bar := pb.StartNew(int(framesToRecord))

	// Run a goroutine to take a screenshot of each frame.
	for f := 0; f < framesToRecord; f++ {
		var ctx context.Context
		ctx = nil
		// Check if there are no more routines available
		if availableWorkers == 0 {
			// Wait until a new routine is ready
			ctx = <- routineReady
			// Increase the number of available routines
			availableWorkers += 1
		}
		// Decrease the number of available routines
		availableWorkers -= 1
		// If no context is defined, we create a new one.
		if ctx == nil {
			var cancel context.CancelFunc
			ctx, cancel = browser.Chrome.NewContext(parent)
			if err := browser.Chrome.Navigate(ctx, params.URL); err != nil {
				return err
			}
			defer cancel()
		}
		// Launch a new routine to record the current frame
		go func(ctx context.Context, pf, f int) {
			// Windup clock
			for i := pf + 1; i <= f; i++ {
				script := fmt.Sprintf("timeweb.goTo(%.0f)", math.Ceil(float64(i) * frameDuration))
				if _, err := browser.Chrome.Evaluate(ctx, script); err != nil {
					panic(err)
				}
			}
			// Take screenshot
			frame, err := browser.Chrome.Screenshot(ctx, params.Viewport)
			if err != nil {
				panic(err)
			}
			// Store screenshot
			if err := params.Writer.WriteFile(fmt.Sprintf("/tmp/%06d.png", f), frame, 0644); err != nil {
				panic(err)
			}
			// Update the bar
			bar.Increment()
			// Notify that a goroutine is available
			routineReady <- ctx
		}(ctx, previousFrames[ctx], f)
		// Update the previousFrames map
		previousFrames[ctx] = f
	}

	// Wait for all running goroutines to finish
	for availableWorkers < params.Workers {
		<- routineReady
		availableWorkers += 1
	}

	// Finish the bar
	bar.Finish()

	return nil
}