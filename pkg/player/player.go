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

// FrameDelayOptions modify the timing between records.
type FrameDelaysOptions struct {
	// Maximum delay between frames in ms.
	// Ignored if the `frameDelay` option is set to `auto`.
	// Set to `auto` to prevent limiting the max idle time.
	maxIdleTime configure.Auto
	// Delay between frames in ms.
	// If the value is "auto" use the actual recording delay.
	frameDelay configure.Auto
	// Multiply the frames delay by a speed factor
	speedFactor int
}

// NewFrameDelayOptions returns a default FrameDelaysOptions struct.
func NewFrameDelayOptions() FrameDelaysOptions {
	return FrameDelaysOptions{configure.Auto(-1), configure.Auto(-1), 1}
}

func Play (recordingPath string, options PlayOptions) error {
	// Parse the recording file
	recording, err := record.ReadRecording(recordingPath)
	if err != nil {
		return err
	}

	// Create a default FrameDelayOptions stuct
	frameDelayOptions := NewFrameDelayOptions()

	// Override frameDelayOptions with the configuration in the recording.config
	frameDelayOptions.frameDelay = recording.Config.FrameDelay
	frameDelayOptions.maxIdleTime = recording.Config.MaxIdleTime

	// Modify the frameDelayOptions according to the PlayOptions
	if options.realTiming {
		frameDelayOptions.frameDelay = configure.Auto(-1)
		frameDelayOptions.maxIdleTime = configure.Auto(-1)
	}

	// Override frameDelayOptions according to the PlayOptions
	frameDelayOptions.speedFactor = options.speedFactor

	// Modify the delay between records according to FramDelayOptions
	adjustFrameDelays(&recording.Records, frameDelayOptions)

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

// Adjust the timing between records accordin to the options
func adjustFrameDelays(records *[]record.Record, options FrameDelaysOptions) {
	for i := 0; i < len(*records); i++ {
		delay := configure.Auto((*records)[i].Delay)
		// Adjust the delay according to the frameDelay and maxIdleTime options
		if options.frameDelay != -1 {
			(*records)[i].Delay = int(options.frameDelay)
		} else if (options.maxIdleTime != -1 && options.maxIdleTime < delay) {
			(*records)[i].Delay = int(options.maxIdleTime)
		}
		// Apply the speed factor
		(*records)[i].Delay = (*records)[i].Delay * options.speedFactor
	}
}