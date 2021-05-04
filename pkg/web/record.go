package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/paulbellamy/ratecounter"
	"gux.codes/omega/pkg/utils"
)

// ConsoleMessage are messages passed through the client.
type ConsoleMessage struct {
	Type string `json:"type"`
	Action string `json:"action"`
	Message string `json:"message"`
}

func Record() error {
	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)
	defer cancel()

	// Run the web server
	webServerOptions := NewWebServerOptions()
	go Serve(webServerOptions)

	// Set up a channel so we can block later while we get each animation frame
	done := make(chan bool, 1)

	// Setup a message channel
	messages := make(chan *ConsoleMessage)

	// Setup a slice of frames
	frames := [][]byte{}

	// Start to listen to the console.log() messages
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			for _, arg := range ev.Args {
				message := &ConsoleMessage{}
				data := strings.ReplaceAll(fmt.Sprintf("%s", arg.Value[1 : len(arg.Value) -1]), "\\", "")
				err := json.Unmarshal([]byte(data), &message)
				if err != nil {
					utils.Error(err.Error())
					utils.Error(data)
					done <- true
					break
				}
				messages <- message
			}
		case *runtime.EventExceptionThrown:
			fmt.Printf("%s %s\n", utils.BoxRed("exception"), ev.ExceptionDetails.Error())
			done <- true
		}
	})

	// Start the event listener to handle message commands
	go func(){
		close := make(chan bool, 1)

		for message := range messages {
			if message.Type == "command" {
				utils.Command(message.Message)
			}

			switch message.Action {
			case "start":
				go startRecording(ctx, &frames, done, close)
			case "close":
				close <- true
			case "done":
				done <- true
			}
		}
	}()

	// Navigate to the URL
	url := fmt.Sprintf("http://localhost:%d/handler", webServerOptions.Port)
	if err := navigate(ctx, url); err != nil {
		return err
	}

	// Block until recording all the animations is done.
	<- done

	// Dump all the screenshots to disk
	utils.Info("\nDumping frames into files")
	// Create a bar to let the user know of the dump status.
	bar := pb.StartNew(len(frames))

	// Write the files to the filesystem
	for index, value := range frames {
		err := ioutil.WriteFile(fmt.Sprintf("/tmp/%06d.png", index), value, 0644)
		if err != nil {
			return err
		}
		// Move the bar forward
		bar.Increment()
	}

	// Close the bar progress
	bar.Finish()

	return nil
}

func startRecording(ctx context.Context, frames *[][]byte, done chan bool, close chan bool) {
	next := make(chan int)
	counter := ratecounter.NewRateCounter(1 * time.Second)

	// Initiate loop
	go func() {next <- 0}()

	// Start the virtual time loop
	for {
		select {
		case <- close:
			return
		case virtualTime := <- next:
			var res []byte
			var frame []byte

			err := chromedp.Run(ctx, chromedp.Tasks{
				screenshot(1920, 1080, &frame),
				chromedp.ActionFunc(func(ctx context.Context) error {
					*frames = append(*frames, frame)

					return nil
				}),
				chromedp.Evaluate(fmt.Sprintf("timeweb.goTo(%d)", virtualTime), &res),
				chromedp.ActionFunc(func(ctx context.Context) error {
					counter.Incr(1)
					virtualTime += 16
					fmt.Printf("%s %d fps | virtualTime: %d ms\r", utils.BoxBlue("info"), counter.Rate(), virtualTime)

					// Continue loop
					go func() {next <- virtualTime}()

					return nil
				}),
			})
			if err != nil {
				utils.Error(err.Error())
				close <- true
			}
		}
	}
}