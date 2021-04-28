package shell

import (
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

// Record corresponds to a PTY interface stdout record
type Record struct {
  // Delay from the last record.
	Delay int `yaml:"delay"`
  // Content of the record.
	Content string `yaml:"content"`
}

// ShellSpecification dictates how the pty session will be recorded.
type ShellSpecification struct {
  // Command to execute on the pty interface.
  Command string
  // CWD corresponds to the Current Working Directory
  Cwd string
  // Env is a map of environment variables that will override default environment variables.
  Env []string
	// MinDelay specifies the minimum delay in ms between two records.
	MinDelay int
  // Cols represent the number of columns to display on the pty interface.
  Cols int
  // Rows represent the number of rows to display on the pty interface.
  Rows int
	// OutputPath indicates the path where the recording will be saved.
	OutputPath string
}

// NewShellSpecification returns a default ShellSpecification.
func NewShellSpecification() *ShellSpecification {
	return &ShellSpecification{
		Cols:	-1,
		Command: "/bin/bash",
		Cwd: "/tmp",
		Env: make([]string, 0),
		MinDelay: 5,
		OutputPath: "/tmp/recording.yaml",
		Rows: -1,
	}
}

// Shell runs a pty shell that will record stdout into a recordings file.
func Shell(specification ShellSpecification) error {
	// Create a command
	c := exec.Command(specification.Command)

	// Add the environment variables provied by the config
	c.Env = append(os.Environ(), specification.Env...)

	// Modify the Current Working Directory of the command.
	c.Dir = specification.Cwd

	// Start command with a pty
	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}

	// Make sure the pty closes at the end
	defer func() { _ = ptmx.Close() }()

	// Listen to the Signal Windows Change to redraw the window.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)

	go func() {
		for range ch {
			// If Rows and Cols are defined then resize the window.
			if specification.Rows != -1 && specification.Cols != -1 {
				err := pty.Setsize(ptmx, &pty.Winsize{Rows: uint16(specification.Rows), Cols: uint16(specification.Cols)})
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

 	// Initial resize
	ch <- syscall.SIGWINCH

	// Set stdin in raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	// Restore the old state of stdin when done.
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	// Create a RecordWriter
	writer := NewShellWriter(&specification)

	// Create a MultiWriter
	multi := io.MultiWriter(writer, os.Stdout)

	// Copy stdin to the pty and the pty to stdout and writer
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	if _, err = io.Copy(multi, ptmx); err != nil {
		panic(err)
	}

	defer func() {
		writer.Dump()
	}()

	return nil
}