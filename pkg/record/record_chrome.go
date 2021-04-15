package record

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
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
}

// NewChromeRecordingSpecification creates a new specification to record
// a Chrome animation.
func NewChromeRecordingSpecification() ChromeRecordingSpecification {
	return ChromeRecordingSpecification{
		Port: 38080,
		Method: "timescale",
		ChromeFlags: chrome.NewOpenOptions(),
	}
}

const port = 38080

func Chrome(specification ChromeRecordingSpecification) error {
	c := chrome.New()

	err := c.Open(specification.ChromeFlags)
	if err != nil {
		return err
	}
	defer c.Close()

	// Subscribe to the client messages
	go func() {
		utils.Info("subscribing to animation messages")
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

	// Create a new sceeencast writer
	writer := NewFramesWriter()

	// Write new frames
	// We're recording marks-per-1second
	counter := ratecounter.NewRateCounter(1 * time.Second)
	go func() {
		count := 0
		for frame := range c.Frames {
			writer.Write(frame)
			counter.Incr(1)
			count += 1
			// Print a new line to avoid deleting an existing line.
			fmt.Printf("%s %d fps | %d frames captured\r", utils.BoxBlue("info"), counter.Rate(), count)
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

// startServers starts the server from which the handler function is served.
func startServer(method string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("error in: ", r)
		}
	}()
	// Force logs in color
	gin.ForceConsoleColor()
	// Set the "release" mode
	gin.SetMode(gin.ReleaseMode)
	// Create the default router
	router := gin.Default()
	// Load the templates
	router.LoadHTMLGlob("./templates/*")
	// Serve static assets
	router.Static("/assets", "./assets")
	// Create the routes
	router.GET("/handler", func(c *gin.Context) {
		c.HTML(http.StatusOK, "handler.html.tmpl", gin.H{"width": 1920, "height": 1080, "method": method})
	})
	// Run the server
	router.Run(fmt.Sprintf(":%d", port))
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
