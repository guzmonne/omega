package record

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
	"gux.codes/omega/pkg/configure"
)

// Record corresponds to a PTY interface stdout record
type Record struct {
  // Delay from the last record.
	Delay int `yaml:"delay"`
  // Content of the record.
	Content string `yaml:"content"`
}

// Recording is the output from a Recording Session
type Recording struct {
  // Config used for the recording.
	Config configure.Config `yaml:"config"`
  // Records correspond to the list of stdout outputs generated during the recording.
	Records []Record `yaml:"records,omitempty"`
}

// FrameDelayOptions modify the timing between records.
type FrameDelaysOptions struct {
	// MaxIdleTime fixes the maximum delay between frames in ms.
	// Ignored if the `frameDelay` option is set to `auto`.
	// Set to `auto` to prevent limiting the max idle time.
	MaxIdleTime configure.Auto
	// FrameDelay sets a fiz delay between records in ms.
	// If the value is "auto" use the actual recording delay.
	FrameDelay configure.Auto
	// SpeedFactor multiplies the delay between records.
	SpeedFactor int
}

// NewFrameDelayOptions returns a default FrameDelaysOptions struct.
func NewFrameDelayOptions() FrameDelaysOptions {
	return FrameDelaysOptions{
		MaxIdleTime: configure.Auto(-1),
		FrameDelay: configure.Auto(-1),
		SpeedFactor: 1,
	}
}

// AdjustFrameDelays adjusts the delays between records according to the
// provided options.
func (recording *Recording) AdjustFrameDelays(options FrameDelaysOptions) {
	// Adjust the timing between records accordin to the options
	for i := 0; i < len(recording.Records); i++ {
		delay := configure.Auto((recording.Records)[i].Delay)
		// Adjust the delay according to the frameDelay and maxIdleTime options
		if options.FrameDelay != -1 {
			recording.Records[i].Delay = int(options.FrameDelay)
		} else if (options.MaxIdleTime != -1 && options.MaxIdleTime < delay) {
			recording.Records[i].Delay = int(options.MaxIdleTime)
		}
		// Apply the speed factor
		recording.Records[i].Delay = recording.Records[i].Delay * options.SpeedFactor
	}
}

// WriteRecording writes the Recording to a YAML file on the path provided by
// the variable recordingPath.
func WriteRecording(recordingPath string, config *configure.Config, records []Record) error {
	var file bytes.Buffer

	// Create a custom YAML encoder
	encoder := yaml.NewEncoder(&file)
	encoder.SetIndent(2)

	// Marshall to YAML the Recording struct
	err := encoder.Encode(Recording{*config, records})
	if err != nil {
		return err
	}

	// Write the recording file
	err = ioutil.WriteFile(recordingPath, file.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

// ReadRecording reads a recording from a file and returns its contents.
func ReadRecording(recordingPath string) (*Recording, error) {
	// Check if the config exists at `configPath`
	if _, err := os.Stat(recordingPath); err != nil {
		return &Recording{}, errors.New("Can't find a file at: " + recordingPath)
	}
	// Open the configuration file
	configFile, err := ioutil.ReadFile(recordingPath)
	if err != nil {
		return &Recording{}, err
	}
	// Unmarshall the configuration file
	recording := &Recording{}
	if err := yaml.Unmarshal(configFile, &recording); err != nil {
		return &Recording{}, err
	}

	return recording, nil
}

// RecordShell runs a pty shell that will record stdout into a recordings file.
func RecordShell(recordingPath string, config *configure.Config) error {
	// Create a command
	c := exec.Command(config.Command)

	// Add the environment variables provied by the config
	c.Env = append(os.Environ(), config.Env.Values...)

	// Modify the Current Working Directory of the command.
	c.Dir = config.Cwd

	// Start command with a pty
	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}

	// Make sure the pty closes at the end
	defer func() { _ = ptmx.Close() }()

	// Handle the pty size
	ch := make(chan os.Signal, 1)
	// Send a Signal Windows Change to redraw the window.
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			// If Rows and Cols are defined then resize the window.
			if config.Rows != -1 && config.Cols != -1 {
				err := pty.Setsize(ptmx, &pty.Winsize{Rows: uint16(config.Rows), Cols: uint16(config.Cols)})
				// If an error occurred print it and let the window inherit sdtin size.
				if err != nil {
					log.Printf("error applying custom size to pty: %s", err)
				} else {
					break
				}
			}
			// Set the pty window to the same size as stdin.
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("error resizing pty: %s", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize

	// Set stdin in raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	// Restore the old state of stdin when done.
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

		// Create a RecordWriter
	writer := NewRecordWriter()

	// Create a MultiWriter
	multi := io.MultiWriter(writer, os.Stdout)

	// Copy stdin to the pty and the pty to stdout and writer
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	if _, err = io.Copy(multi, ptmx); err != nil {
		panic(err)
	}

	defer func() {
		WriteRecording(recordingPath, config, writer.records)
	}()

	return nil
}
