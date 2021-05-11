package chrome

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/cheggaaa/pb"
	"gux.codes/omega/pkg/browser"
	"gux.codes/omega/pkg/utils"
)

// RecordParams specifies how the recording should be done.
type RecordParams struct {
	// Duration to recording.
	Duration float64
	// URL specifies the URL to be recorded.
	URL string
	// Interface used to write the frames to disk.
	Writer utils.Writer
	// Workers used for the recording.
	Workers int
	// Viewport width.
	Width int64
	// Viewport height
	Height int64
}

// FPS represents the fixed FPS to record an animation.
const FPS float64 = 60.0
// FRAME_DURATION represent the interval between frames when animating at
// 60 fps, which is when `requestAnimationFrame` should update.
const FRAME_DURATION float64 = 1000.0 / FPS

type StdinWriter struct {
	stdin io.WriteCloser
}

// WriteFile exposes the ioutil.WriteFile function
func (w StdinWriter) WriteFile(filename string, data []byte, perm os.FileMode) error {
	_, err := w.stdin.Write(data)

	return err
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

	// Create the ffmpeg command
	cmd := exec.Command(`ffmpeg`,
		`-y`,
		`-framerate`, `60`,
		`-i`, `pipe:0`,
		`-c:v`, `libx264`,
		`-pix_fmt`, `yuv420p`,
  	`-r`, `60`,
		`/tmp/out.mp4`,
	)

	// Pipe cmd stderr and stdout to the console
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// Open stdin pipe
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	// Override the provided RecordParams
	params.Writer = StdinWriter{stdin}
	params.URL = fmt.Sprintf("http://localhost:%d/handler", webServerOptions.Port)

	// Start the ffmpeg command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Start the recording process
	if err := record(ctx, params); err != nil {
		return err
	}

	// Close stdin, or ffmpeg will wait forever
	if err := stdin.Close(); err != nil {
		return err
	}

	// Wait until ffmpeg finishes
	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func record(parent context.Context, params RecordParams) error {
	// Store all the available workers
	availableWorkers := params.Workers

	// Calculate the amount of frames ro record.
	framesToRecord  := params.Duration * FPS / 1000

	// Create a channel to communicate when a goroutine is ready to do more work
	// providing an appropiate context.
	routineReady := make(chan context.Context)

	// Create a map that holds the frames that each browser has last processed
	previousFrames := make(map[context.Context]float64)

	// Instantiate the progress bar.
	bar := pb.StartNew(int(framesToRecord))

	// Run a goroutine to take a screenshot of each frame.
	for f := 0.0; f < framesToRecord; f++ {
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
			if err := browser.Chrome.Navigate(ctx, params.URL, params.Width, params.Height); err != nil {
				return err
			}
			defer cancel()
		}
		// Launch a new routine to record the current frame
		go func(ctx context.Context, pf, f float64) {
			// Windup clock
			for i := pf + 1; i <= f; i++ {
				script := fmt.Sprintf("timeweb.goTo(%.3f)", float64(i) * FRAME_DURATION)
				if _, err := browser.Chrome.Evaluate(ctx, script); err != nil {
					panic(err)
				}
			}
			// Take screenshot
			frame, err := browser.Chrome.Screenshot(ctx)
			if err != nil {
				panic(err)
			}
			// Store screenshot
			if err := params.Writer.WriteFile(fmt.Sprintf("/tmp/%06.0f.png", f), frame, 0644); err != nil {
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