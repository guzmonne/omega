package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/paulbellamy/ratecounter"
	"gux.codes/omega/pkg/utils"
)

// StartRecording starts the process of recording a Chrome animation.
func StartRecording() error {
	doneCh		:= make(chan struct{}, 1)
	errCh			:= make(chan error, 1)
	messageCh	:= make(chan ConsoleMessage)

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Setup a Ctrl+C handler
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func(){
		<- signals
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		cancel()
		os.Exit(0)
	}()

	// Run the web server
	webServerOptions := NewWebServerOptions()
	go startWebServer(*webServerOptions)

	// Start the console listener
	consoleListener(ctx, errCh, messageCh)

	// Start the frame capturer
	c := framesCapturer(ctx, doneCh, errCh, messageCh)

	// Navigate to the handler function.
	urlstr := fmt.Sprintf("http://localhost:%d/handler", webServerOptions.Port)
	utils.Info("navigating to " + urlstr)
	if err := chromedp.Run(ctx, chromedp.Tasks{chromedp.Navigate(urlstr)}); err != nil {
		errCh <- err
	}

	// Block until done, or an error occurs.
	select {
	case err := <- errCh:
		return err
	case <- doneCh:
	}

	// Close channels as we don't need them anymore
	close(errCh)
	close(messageCh)

	// Create a bar to let the user know of the dump status.
	bar := pb.StartNew(len(c.frames))

	c.mu.Lock()
	for i, frame := range c.frames {
		err := ioutil.WriteFile(fmt.Sprintf("/tmp/%06d.png", i), frame, 0644)
		if err != nil {
			return err
		}
		// Move the bar forward
		bar.Increment()
	}
	c.mu.Unlock()

	return nil
}

// consoleListener starts a function that listens to the client console messages.
func consoleListener(ctx context.Context, errCh chan error, messageCh chan ConsoleMessage) {
	go func () {
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			go func() {
				switch ev := ev.(type) {
				case *runtime.EventConsoleAPICalled:
					for _, arg := range ev.Args {
						message := ConsoleMessage{}
						data := strings.ReplaceAll(fmt.Sprintf("%s", arg.Value[1 : len(arg.Value) -1]), "\\", "")
						err := json.Unmarshal([]byte(data), &message)
						if err != nil {
							errCh <- err
							break
						}
						messageCh <- message
					}
				case *runtime.EventExceptionThrown:
					errCh <- errors.New(ev.ExceptionDetails.Error())
				}
			}()
		})
	}()
}

// framesCapturer handles the caputing of frames.
func framesCapturer(ctx context.Context, doneCh chan struct{}, errCh chan error, messageCh chan ConsoleMessage) *capturer {
	c := capturer{
		counter: ratecounter.NewRateCounter(1 * time.Second),
		mu: sync.Mutex{},
		status: "idle",
		virtualTime: 0,
		frames: [][]byte{},
	}

	go func(){
		for {
			select {
			case <- ctx.Done():
				c.Done()
				return
			case message := <- messageCh:
				switch message.Action {
				case "start":
					c.Start(ctx, errCh)
				case "close":
					c.Stop()
				case "done":
					c.Done()
					close(doneCh)
				}
			}
		}
	}()

	return &c
}

type capturer struct {
	counter *ratecounter.RateCounter

	mu sync.Mutex
	status string
	virtualTime int
	frames [][]byte
}

func (c *capturer) Start(ctx context.Context, errCh chan error) {
	if c.status == "done" {
		return
	}

	c.mu.Lock()
	c.status = "capturing"
	c.mu.Unlock()

	go c.capture(ctx, errCh)
}

func (c *capturer) Stop() {
	if c.status == "done" {
		return
	}

	c.mu.Lock()
	c.status = "idle"
	c.mu.Unlock()
}

func (c *capturer) Done() {
	c.mu.Lock()
	c.status = "done"
	c.mu.Unlock()
}

func (c *capturer) capture(ctx context.Context, errCh chan error) {
	if c.status != "capturing" {
		return
	}

	var frame []byte
	c.mu.Lock()
	err := chromedp.Run(ctx, chromedp.Tasks{screenshot(1920, 1080, &frame)})
	if err != nil {
		errCh <- err
		return
	}
	c.frames = append(c.frames, frame)
	c.counter.Incr(1)
	fmt.Printf("%s %d fps | virtualTime: %d ms\r", utils.BoxBlue("info"), c.counter.Rate(), c.virtualTime)
	c.mu.Unlock()

	go c.nextVT(ctx, errCh)
}

func (c *capturer) nextVT(ctx context.Context, errCh chan error) {
	if c.status != "capturing" {
		return
	}

	var res []byte
	c.mu.Lock()
	c.virtualTime += 16
	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Evaluate(fmt.Sprintf("timeweb.goTo(%d)", c.virtualTime), &res),
	})
	if err != nil {
		errCh <- err
		return
	}
	c.mu.Unlock()

	go c.capture(ctx, errCh)
}