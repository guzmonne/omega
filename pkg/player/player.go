package player

import (
	"os"
	"os/signal"
)

// PlayOptions modify the way the recording is played.
type PlayOptions struct {
	// Use the actual delays between frames as recorded?
	realTiming bool
	// Show message before playing the recording?
	silent bool
	// Factor for which the delays will be multiplied by.
	speedFactor int
}

func NewPlayOptions() PlayOptions {
	return PlayOptions{true, false, 1}
}

func Play (recordingPath string, options PlayOptions) error {
	// Get a reference to stdout
	//var stdout *bufio.Writer = bufio.NewWriter(os.Stdout)

	// Capture the interrupt and kill signals
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, os.Kill)

	go func() {
		Clear()

	}()

	// Block
	select {
	case <- signals:
	}

	return nil
}