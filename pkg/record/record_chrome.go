package record

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/paulbellamy/ratecounter"
	"gux.codes/omega/pkg/chrome"
	"gux.codes/omega/pkg/utils"
)

type ChromeRecordingSpecification struct {
	// Port defines the port to be opened on the web server.
	Port int
	// Method defines the recording method to be used. Must be one of:
	// "timescale" or "screencast".
	Method string
	// ChromeFlags defines the set of flags to be used when opening a
	// new Chrome instance.
	ChromeFlags chrome.OpenOptions
	// VirtualTime is the initial time to be used when the method of
	// recording is `timeweb`.
	VirtualTime int
	// FPS is the rate at which the animation will be recorded when
	// using the `timeweb` method of recordin.
	FPS float64
}

// NewChromeRecordingSpecification creates a new specification to record
// a Chrome animation.
func NewChromeRecordingSpecification() ChromeRecordingSpecification {
	return ChromeRecordingSpecification{
		Port: 38080,
		Method: "timescale",
		VirtualTime: 0,
		FPS: 60.0,
		ChromeFlags: chrome.NewOpenOptions(),
	}
}

const port = 38080

func Chrome(specification ChromeRecordingSpecification) error {
	c := chrome.New()
	// Override default Chrome options
	c.VirtualTime = specification.VirtualTime
	c.FPS = specification.FPS

	err := c.Open(specification.ChromeFlags)
	if err != nil {
		return err
	}
	defer c.Close()

	// Write info to the console
	utils.Info(fmt.Sprintf("FPS: %2.f", c.FPS))
	utils.Info(fmt.Sprintf("VirtualTime: %d", c.VirtualTime))

	// Subscribe to the client messages
	go func() {
		utils.Info("Subscribing to animation messages")
		for message := range c.AnimationMessages {
			utils.Message(message.Message)
		}
	}()

	// Setup our Ctrl+C handler
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go setupCloseHandler(signals, c.Close)

	// Start the HTTP server
	go startServer(specification.Method)

	// Create hold channel
	close := make(chan bool, 1)

	// Setup the commands handlers
	go func() {
		for command := range c.AnimationCommands {
			// Print the command command
			utils.Command(command.Message)
			// Handle the command action
			switch command.Action {
			case "start":
				c.StartRecording(specification.Method)
			case "close":
				c.StopRecording(specification.Method)
			case "done":
				close <- true
			}
		}
	}()

	// Create a new frames writer
	writer := NewFramesWriter()

	// Create a counter to keep track of the number of recorded frames.
	counter := ratecounter.NewRateCounter(1 * time.Second)

	// Setup the Frames writer
	go func() {
		for frame := range c.Frames {
			writer.Write(frame)
			counter.Incr(1)
			fmt.Printf("%s %d fps\r", utils.BoxBlue("info"), counter.Rate())
		}
	}()

	// Navigate to handler URL
	utils.Info("Navigating to handler endpoint.")
	urlstr := fmt.Sprintf("http://localhost:%d/handler", port)
	if err := c.Navigate(urlstr, 30 * time.Second); err != nil {
		return err
	}

	<- close

	// Dump all the screenshots to disk
	utils.Info("Dumping frames into files")
	if err := writer.Dump("/tmp/%06d.png"); err != nil {
		return err
	}

	return nil
}


// SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
func setupCloseHandler(c chan os.Signal, handler func() error) {
	<-c
	fmt.Println("\r- Ctrl+C pressed in Terminal")
	if err := handler(); err != nil {
		fmt.Printf("%s %s\n", utils.BoxRed("message"), err)
	}
	os.Exit(0)
}
