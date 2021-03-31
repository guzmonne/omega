package main

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

//
// PTY interface record
//
type Record struct {
	//
  // Time from the last record.
  //
	Delay int `yaml:"delay"`
	//
  // Content of the record.
  //
	Content string `yaml:"content"`
}
//
// Recording file
//
type Recording struct {
	//
  // Configuration used for the recording.
  //
	Config Config `yaml:"config"`
	//
  // Stored records of the recording.
  //
	Records []Record `yaml:"records"`
}

// RecordShell runs a pty shell that will record stdout onto a recordings file.
func RecordShell(config *Config) error {
	// Create a command
	c := exec.Command(config.Command)

	c.Env = append(os.Environ(), config.Env.Values...)

	// Start command with a pty
	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}

	// Make sure the pty closes at the end
	defer func() { _ = ptmx.Close() }()

	// Handle the pty size
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
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
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	//var export = "export PS1='$ '\nclear\n"
	//ptmx.Write([]byte(export))

	// Copy stdin to the pty and the pty to stdout
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	_, _ = io.Copy(os.Stdout, ptmx)


	return nil
}