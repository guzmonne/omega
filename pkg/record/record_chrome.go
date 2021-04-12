package record

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"
	"golang.org/x/sync/errgroup"
	"gux.codes/omega/pkg/chrome"
	"gux.codes/omega/pkg/utils"
)

type AnimationMessage struct {
	Session int `json:"session"`
	Type string `json:"type"`
	Message string `json:"message"`
	Action string `json:"action"`
}

type AnimationOptions struct {
	Session int
	Port int
}

func NewAnimationOptions() *AnimationOptions {
	return &AnimationOptions{
		Session: rand.Int(),
		Port: 38080,
	}
}

func Chrome() error {
	// Open Chrome in headless mode
	if err := chrome.Start(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	// Create the devtools connnection
	var devt *devtool.DevTools
	var pt *devtool.Target
	var err error
	for i := 0; i < 10; i++ {
		if i > 0 {
			time.Sleep(500 * time.Millisecond)
		}

		devt = devtool.New("http://localhost:9222")
		pt, err = devt.Get(ctx, devtool.Page)
		if err == nil {
			break
		}
	}
	if err != nil {
		return err
	}

	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Create a new CDP client that uses conn.
	c := cdp.NewClient(conn)

	// Give enough capacity to avoid blocking any event listeners
	abort := make(chan error, 2)

	// Watch the abort channel
	go func() {
		select {
		case <- ctx.Done():
		case err := <-abort:
			fmt.Printf("aborted: %s\n", err.Error())
			cancel()
		}
	}()

	if err = abortOnErrors(ctx, c, abort); err != nil {
		return err
	}

	// Enable all the domain events that we're interested in.
	if err = runBatch(
		func() error { return c.DOM.Enable(ctx) },
		func() error { return c.Network.Enable(ctx, nil) },
		func() error { return c.Page.Enable(ctx) },
		func() error { return c.Runtime.Enable(ctx) },
		func() error { return c.Console.Enable(ctx) },

	); err != nil {
		return err
	}

	// Start the HTTP server
	options := NewAnimationOptions()
	go startServer(options)

	// Create a messageAdded client
	messageAdded, err := c.Console.MessageAdded(ctx)
	if err != nil {
		return err
	}

	// Handle added messages
	go func() {
		defer messageAdded.Close()

		for {
			ev, err := messageAdded.Recv()
			if err != nil {
				log.Printf("Failed to receive MessageAdded: %v", err)
				return
			}
			fmt.Printf("Added message [%s]: %s\n", ev.Message.Level, ev.Message.Text)
		}
	}()

	// Navigate to handler URL
	domLoadTimeout := 5 * time.Second
	urlstr := fmt.Sprintf("http://localhost:%d/handler", options.Port)
	err = navigate(ctx, c.Page, urlstr, domLoadTimeout)
	if err != nil {
		return err
	}

	// Create a screencastFrame client
	screencastFrame, err := c.Page.ScreencastFrame(ctx)
	if err != nil {
		return err
	}

	// Handle sceencast frames
	go func() {
		defer screencastFrame.Close()

		for {
			ev, err := screencastFrame.Recv()
			if err != nil {
				log.Printf("Failed to receive ScreencastFrame: %v", err)
				return
			}

			err = c.Page.ScreencastFrameAck(ctx, page.NewScreencastFrameAckArgs(ev.SessionID))
			if err != nil {
				log.Printf("Failed to ack ScreencastFrame: %v", err)
				return
			}

			// Write the frame to file (without blocking).
			go func() {
				name := utils.ULID() + ".png"
				err = ioutil.WriteFile(name, ev.Data, 0644)
				if err != nil {
					log.Printf("Failed to write ScreencastFrame to %q: %v", name, err)
				}
			}()
		}
	}()

	screencastArgs := page.NewStartScreencastArgs().
		SetEveryNthFrame(1).
		SetFormat("png")
	err = c.Page.StartScreencast(ctx, screencastArgs)
	if err != nil {
		return err
	}

	// Random delay for our screencast.
	time.Sleep(5 * time.Second)

	err = c.Page.StopScreencast(ctx)
	if err != nil {
		return err
	}

	return nil
}

// navigate to the URL and wait for DOMContentEventFired. An error is
// returned if timeout happens before DOMContentEventFired.
func navigate(ctx context.Context, pageClient cdp.Page, url string, timeout time.Duration) error {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	// Make sure Page events are enabled.
	err := pageClient.Enable(ctx)
	if err != nil {
		return err
	}

	// Open client for DOMContentEventFired to block until DOM has fully loaded.
	domContentEventFired, err := pageClient.DOMContentEventFired(ctx)
	if err != nil {
		return err
	}
	defer domContentEventFired.Close()

	_, err = pageClient.Navigate(ctx, page.NewNavigateArgs(url))
	if err != nil {
		return err
	}

	_, err = domContentEventFired.Recv()
	return err
}

// runBatchFunc is the function signature for runBatch.
type runBatchFunc func() error

// runBatch runs all functions simultaneously and waits until
// execution has completed or an error is encountered.
func runBatch(fn ...runBatchFunc) error {
	eg := errgroup.Group{}
	for _, f := range fn {
		eg.Go(f)
	}
	return eg.Wait()
}

func abortOnErrors(ctx context.Context, c *cdp.Client, abort chan<- error) error {
	exceptionThrown, err := c.Runtime.ExceptionThrown(ctx)
	if err != nil {
		return err
	}

	loadingFailed, err := c.Network.LoadingFailed(ctx)
	if err != nil {
		return err
	}

	go func() {
		defer exceptionThrown.Close() // Cleanup.
		defer loadingFailed.Close()
		for {
			select {
			// Check for exceptions so we can abort as soon
			// as one is encountered.
			case <-exceptionThrown.Ready():
				ev, err := exceptionThrown.Recv()
				if err != nil {
					// This could be any one of: stream closed,
					// connection closed, context deadline or
					// unmarshal failed.
					abort <- err
					return
				}

				// Ruh-roh! Let the caller know something went wrong.
				abort <- ev.ExceptionDetails

			// Check for non-canceled resources that failed
			// to load.
			case <-loadingFailed.Ready():
				ev, err := loadingFailed.Recv()
				if err != nil {
					abort <- err
					return
				}

				// For now, most optional fields are pointers
				// and must be checked for nil.
				canceled := ev.Canceled != nil && *ev.Canceled

				if !canceled {
					abort <- fmt.Errorf("request %s failed: %s", ev.RequestID, ev.ErrorText)
				}
			}
		}
	}()
	return nil
}

func startServer(options *AnimationOptions) {
	// Force logs in color
	gin.ForceConsoleColor()
	// Set the "release" mode
	gin.SetMode(gin.ReleaseMode)
	// Create the default router
	router := gin.Default()
	// Load the templates
	router.LoadHTMLGlob("templates/*")
	// Create the routes
	router.GET("/handler", func(c *gin.Context) {
		c.HTML(http.StatusOK, "handler.tmpl", options)
	})
	// Run the server
	router.Run(fmt.Sprintf(":%d", options.Port))
}