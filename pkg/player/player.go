package player

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
	"gux.codes/omega/pkg/record"
)

// PlayOptions modify the way the recording is played.
type PlayOptions struct {
	// MaxIdleTime fixes the maximum delay between frames in ms.
	// Ignored if the `frameDelay` option is set to `auto`.
	// Set to `auto` to prevent limiting the max idle time.
	MaxIdleTime int
	// FrameDelay sets a fixed delay between records in ms.
	// If the value is "auto" use the actual recording delay.
	FrameDelay int
	// Silent eliminates the recording message before playback.
	Silent bool
	// SpeedFactor applies a custom factor between record delays.
	SpeedFactor float64
}

// NewPlayOptions returns a default PlayOptions struct.
func NewPlayOptions() PlayOptions {
	return PlayOptions{
		MaxIdleTime: -1,
		FrameDelay: -1,
		Silent: false,
		SpeedFactor: 1.0,
	}
}

// ReadRecording reads a recording from a file and returns its contents.
func ReadRecording(recordingPath string) ([]record.Record, error) {
	var records []record.Record
	// Check if the config exists at `configPath`
	if _, err := os.Stat(recordingPath); err != nil {
		return records, errors.New("Can't find a file at: " + recordingPath)
	}
	// Open the configuration file
	configFile, err := ioutil.ReadFile(recordingPath)
	if err != nil {
		return records, err
	}
	// Unmarshall the configuration file
	if err := yaml.Unmarshal(configFile, &records); err != nil {
		return records, err
	}

	return records, nil
}

// AdjustFrameDelays adjusts the delays between records according to the
// provided options.
func AdjustFrameDelay(records []record.Record, options PlayOptions) []record.Record {
	modifiedRecords := make([]record.Record, 0)

	for _, record := range records {
		if options.FrameDelay != -1 {
			record.Delay = options.FrameDelay
		}

		if (options.MaxIdleTime != -1 && options.MaxIdleTime < record.Delay) {
			record.Delay = options.MaxIdleTime
		}

		record.Delay = int(float64(record.Delay) * options.SpeedFactor)

		modifiedRecords = append(modifiedRecords, record)
	}

	return modifiedRecords
}


func Play (recordingPath string, options PlayOptions) error {
	// Parse the recording file
	records, err := ReadRecording(recordingPath)
	if err != nil {
		return err
	}

	// Modify the delay between records according to FramDelayOptions
	records = AdjustFrameDelay(records, options)

	if !options.Silent {
		showPlaybackMessage(recordingPath, options)
	}

	// Capture the interrupt and kill signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// Interrupt channel
	interrupt := make(chan bool)

	go func() {
		Clear()
		for _, record := range records {
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

func showPlaybackMessage(recordingPath string, options PlayOptions) {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Printf("\nPlaying %s\n", green(recordingPath))
	fmt.Printf("\nPlayback options:\n")
	fmt.Printf(("\tFrame Delay:\t%d\n"), options.FrameDelay)
	fmt.Printf(("\tMax Idle Time:\t%d\n"), options.MaxIdleTime)
	fmt.Printf(("\tSpeed Factor:\t%.2f\n"), options.SpeedFactor)
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
	fmt.Printf("Thank you for using %s!\n", magenta("Ωmega"))
}