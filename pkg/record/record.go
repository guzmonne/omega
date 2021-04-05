package record

import (
	"bytes"
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
