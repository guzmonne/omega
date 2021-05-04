package web

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"

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
	frames := int(math.Ceil(1000.0 / 16)) + 1 // We add one more frame to handle edge cases.

	// Instantiate the progress bar.
	bar := pb.StartNew(frames)

	for i := 0; i < frames; i++ {
		frame, err := browser.Screenshot(ctx, 0)
		if err != nil {
			return err
		}
		go func(){
			err := ioutil.WriteFile(fmt.Sprintf("/tmp/%06d.png", i), frame, 0644)
			if err != nil {
				panic(err)
			}
			bar.Increment()
		}()
		if _, err = browser.GoTo(ctx, 0, i * 16); err != nil {
			return err
		}
	}

	// Finish the bar
	bar.Finish()

	return nil
}