package player

import (
	"os"
	"os/signal"

	"gux.codes/omega/pkg/configure"
	"gux.codes/omega/pkg/record"
)

// PlayOptions modify the way the recording is played.
type PlayOptions struct {
	// Use the actual delays between frames as recorded?
	realTiming bool
	// Show message before playing the recording?
	silent bool
	// Multiply the frames delay by a speed factor
	speedFactor int
}

// NewPlayOptions returns a default PlayOptions struct.
func NewPlayOptions() PlayOptions {
	return PlayOptions{true, false, 1}
}

func Play (recordingPath string, options PlayOptions) error {
	// Parse the recording file
	recording, err := record.ReadRecording(recordingPath)
	if err != nil {
		return err
	}

	// Create a default FrameDelayOptions stuct
	frameDelayOptions := record.NewFrameDelayOptions()

	// Override frameDelayOptions with the configuration in the recording.config
	frameDelayOptions.FrameDelay = recording.Config.FrameDelay
	frameDelayOptions.MaxIdleTime = recording.Config.MaxIdleTime

	// Modify the frameDelayOptions according to the PlayOptions
	if options.realTiming {
		frameDelayOptions.FrameDelay = configure.Auto(-1)
		frameDelayOptions.MaxIdleTime = configure.Auto(-1)
	}

	// Override frameDelayOptions according to the PlayOptions
	frameDelayOptions.SpeedFactor = options.speedFactor

	// Modify the delay between records according to FramDelayOptions
	recording.AdjustFrameDelays(frameDelayOptions)

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