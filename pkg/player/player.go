package player

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/fatih/color"
	"gux.codes/omega/pkg/configure"
	"gux.codes/omega/pkg/record"
)

// PlayOptions modify the way the recording is played.
type PlayOptions struct {
	// RealTiming makes the playback use the actual delays between records.
	RealTiming bool
	// Silent eliminates the recording message before playback.
	Silent bool
	// SpeedFactor applies a custom factor between record delays.
	SpeedFactor int
}

// NewPlayOptions returns a default PlayOptions struct.
func NewPlayOptions() PlayOptions {
	return PlayOptions{
		RealTiming: true,
		Silent: false,
		SpeedFactor: 1,
	}
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
	if options.RealTiming {
		frameDelayOptions.FrameDelay = configure.Auto(-1)
		frameDelayOptions.MaxIdleTime = configure.Auto(-1)
	}

	// Update SpeedFactor
	frameDelayOptions.SpeedFactor = options.SpeedFactor

	// Modify the delay between records according to FramDelayOptions
	recording.AdjustFrameDelays(frameDelayOptions)

	if !options.Silent {
		showPlaybackMessage(recordingPath, frameDelayOptions)
	}

	// Capture the interrupt and kill signals
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, os.Kill)

	// Interrupt channel
	interrupt := make(chan bool)

	go func() {
		Clear()
		for _, record := range recording.Records {
			fmt.Printf(record.Content)
			time.Sleep(time.Duration(record.Delay) * time.Millisecond)
		}
		interrupt <- true
	}()

	// Block
	select {
	case <- signals:
	case <- interrupt:
	}

	if !options.Silent {
		showDoneMessage()
	}

	return nil
}

func showPlaybackMessage(recordingPath string, options record.FrameDelaysOptions) {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Printf("\nPlaying %s\n", green(recordingPath))
	fmt.Printf("\nPlayback options:\n")
	fmt.Printf(("\tFrame Delay:\t%s\n"), options.FrameDelay.String())
	fmt.Printf(("\tMax Idle Time:\t%s\n"), options.MaxIdleTime.String())
	fmt.Printf(("\tSpeed Factor:\t%d\n"), options.SpeedFactor)
	fmt.Printf("\n---\n\n")
	fmt.Printf("Press %s to exit the recording at any time\n", red("CTRL+C"))
	for i := 5; i > -1; i-- {
		if i == 0 {
			fmt.Printf("\rYour recording will begin in... %s", green("Action!"))
		} else {
			fmt.Printf("\rYour recording will begin in... %d", i)
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func showDoneMessage() {
	magenta := color.New(color.FgHiMagenta).SprintFunc()

	Clear()

	fmt.Printf("\033[2;5H")
	color.Green("Done")
	fmt.Printf("\033[4;5H")
	fmt.Printf("Thank you for using %s!\n", magenta("Omega"))
}