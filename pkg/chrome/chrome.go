package chrome

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	run "runtime"
	"time"

	"github.com/gobs/args"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/emulation"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/rpcc"
	"gux.codes/omega/pkg/utils"
)

// Chrome contains abstracts all the logic required to operate Chrome
// through the Chrome Development Tools Protocol.
type Chrome struct {
	AnimationMessages chan AnimationMessage
	AnimationCommands chan AnimationMessage
	Frames chan []byte
	abort chan error
	connection *rpcc.Conn
	cancel context.CancelFunc
	client *cdp.Client
	ctx context.Context
	cancelFuncs []func () error
	stop chan bool
	virtualTime int
	fps float64
}

// NewChrome returns a new Chrome value object with a new context.
func New() Chrome {
	ctx, cancel := context.WithCancel(context.Background())
	return Chrome{
		// Setup the context
		ctx: ctx,
		cancel: cancel,
		// Create the virtualtime handlers
		virtualTime: 0,
		fps: 60.0,
		// Create communication channels
		abort: make(chan error, 2),
		AnimationMessages: make(chan AnimationMessage),
		AnimationCommands: make(chan AnimationMessage),
		Frames: make(chan []byte),
	}
}

type AnimationMessage struct {
	Action string `json:"action"`
	Message string `json:"message"`
	Type string `json:"type"`
}

type OpenOptions struct {
	// DevToolsPort corresponds to the port to be used by the Chrome Developer Protocol
	DevToolsPort int
	// HideScrollbars disables the browser scrollbars.
	HideScrollbars bool
	// BWSI indicates that the browser will run a Guest session.
	BWSI bool
	// DisableExtensions disable the use of browser extensions
	DisableExtensions bool
	// AllowHttpScreenCapture allows non-secure origins to use the screen capture API and
	// the desktopCapture extension API.
	AllowHttpScreenCapture bool
	// AllowInsecuredLocalhost enables TLS/SSL errors on localhost to be ignored.
	AllowInsecuredLocalhost bool
	// CastInitialScreenWidth is used to pass initial screen resolution to GPU process.
	// This allows us to set screen size correctly (so no need to resize when first window is created).
	CastInitialScreenWidth int
	// CastInitialScreenHeight is used to pass initial screen resolution to GPU process.
	// This allows us to set screen size correctly (so no need to resize when first window is created).
	CastInitialScreenHeight int
	// DisableFrameRateLimit disables begin frame limiting in both cc scheduler and display scheduler.
	// Also implies --disable-gpu-vsync
	DisableFrameRateLimit bool
	// DisableGPU disables GPU hardware acceleration. If software renderer is not in place, then
	// the GPU process won't launch.
	DisableGPU bool
	// DisableWebSecurity makes the browser don't enforce the same-origin policy.
	DisableWebSecurity bool
	// EnableAccelerated2dCanvas enables accelerated 2D canvas.
	EnableAccelerated2dCanvas bool
}

func NewOpenOptions() OpenOptions {
	return OpenOptions{
		DevToolsPort: 9222,
		HideScrollbars: true,
		BWSI: true,
		DisableExtensions: true,
		AllowHttpScreenCapture: true,
		AllowInsecuredLocalhost: true,
		CastInitialScreenWidth: 1920,
		CastInitialScreenHeight: 1080,
		DisableFrameRateLimit: true,
		DisableGPU: true,
		DisableWebSecurity: true,
		EnableAccelerated2dCanvas: true,
	}
}

// Close runs the cancel function for the context and closes
// the browser.
func (c Chrome) Close() error {
	for _, f := range c.cancelFuncs {
		if err := f(); err != nil {
			return err
		}
	}
	c.cancel()
	return nil
}


// Open starts the Chrome browser with its devtools and a client
// CDP connection to it.
func (c *Chrome) Open(options OpenOptions) error {
	chromeapp := os.Getenv("OMEGA_CHROMEAPP")

	if chromeapp == "" {
		switch run.GOOS {
		case "darwin":
			for _, c := range []string{
				"/Applications/Google Chrome Canary.app",
				"/Applications/Google Chrome.app",
			} {
				// MacOS apps are actually folders
				if info, err := os.Stat(c); err == nil && info.IsDir() {
					chromeapp = fmt.Sprintf("open %q --args", c)
					break
				}
			}

		case "linux":
			for _, c := range []string{
				"headless_shell",
				"chromium",
				"google-chrome-beta",
				"google-chrome-unstable",
				"google-chrome-stable"} {
				if _, err := exec.LookPath(c); err == nil {
					chromeapp = c
					break
				}
			}

		case "windows":
			// TODO
		}
	}

	if chromeapp == "" {
		return errors.New("chromeapp not found")
	}

	if chromeapp == "headless_shell" {
		chromeapp += " --no-sandbox"
	} else {
		chromeapp += " --headless"
	}

	chromeapp += fmt.Sprintf(" --remote-debugging-port=%d", options.DevToolsPort)

	if options.HideScrollbars {
		chromeapp += " --hide-scrollbars"
	}
	if options.BWSI {
		chromeapp += " --bwsi"
	}
	if options.DisableExtensions {
		chromeapp += " --disable-extensions"
	}
	if options.AllowHttpScreenCapture {
		chromeapp += " --allow-http-screen-capture"
	}
	if options.AllowInsecuredLocalhost {
		chromeapp += " --allow-insecure-localhost"
	}
	if options.CastInitialScreenHeight > 0 {
		chromeapp += fmt.Sprintf(" --cast-initial-screen-height=%d", options.CastInitialScreenHeight)
	}
	if options.CastInitialScreenWidth > 0 {
		chromeapp += fmt.Sprintf(" --cast-initial-screen-width=%d", options.CastInitialScreenWidth)
	}
	if options.CastInitialScreenWidth > 0 && options.CastInitialScreenHeight > 0 {
		chromeapp += fmt.Sprintf(" --window-size=%d,%d", options.CastInitialScreenWidth, options.CastInitialScreenHeight)
	}
	if options.DisableFrameRateLimit {
		chromeapp += " --disable-frame-rate-limit"
	}
	if options.DisableGPU {
		chromeapp += " --disable-gpu"
	}
	if options.DisableWebSecurity {
		chromeapp += " --disable-web-security"
	}
	if options.EnableAccelerated2dCanvas {
		chromeapp += " --enable-accelerated-2d-canvas"
	}

	chromeapp += " about:blank"

	// Run the chromeapp command
	parts := args.GetArgs(chromeapp)
	cmd := exec.Command(parts[0], parts[1:]...)
	if err := cmd.Start(); err != nil {
		return err
	}

	// Create the devtools connnection
	var devt *devtool.DevTools
	var pt *devtool.Target
	var err error
	for i := 0; i < 30; i++ {
		if i > 0 {
			time.Sleep(1000 * time.Millisecond)
		}

		devt = devtool.New(fmt.Sprintf("http://localhost:%d", options.DevToolsPort))
		pt, err = devt.Get(c.ctx, devtool.Page)
		if err == nil {
			break
		}
	}
	// Check if an error occur while trying to connect to the Chrome devTools.
	if err != nil {
		return err
	}

	c.connection, err = rpcc.DialContext(c.ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		return err
	}

	// Create a new CDP client that uses conn.
	c.client = cdp.NewClient(c.connection)

	// Append the function to close the browser
	c.cancelFuncs = append(c.cancelFuncs, func () error {
		if err := c.client.Browser.Close(c.ctx); err != nil {
			return err
		}
		return nil
	})

	// Watch the abort channel
	go func() {
		select {
		case <- c.ctx.Done():
		case err := <- c.abort:
			fmt.Printf("%s %s\n", utils.BoxRed("error"), err)
			c.Close()
		}
	}()

	if err = c.abortOnErrors(); err != nil {
		return err
	}

	// Enable all the domain events that we're interested in.
	if err = utils.RunBatch(
		func() error { return c.client.DOM.Enable(c.ctx) },
		func() error { return c.client.Network.Enable(c.ctx, nil) },
		func() error { return c.client.Page.Enable(c.ctx) },
		func() error { return c.client.Runtime.Enable(c.ctx) },
		func() error { return c.client.Console.Enable(c.ctx) },
	); err != nil {
		return err
	}

	// Change the size of the page
	if err := c.client.Emulation.SetDeviceMetricsOverride(c.ctx, emulation.NewSetDeviceMetricsOverrideArgs(
		options.CastInitialScreenWidth,
		options.CastInitialScreenHeight,
		1.0,
		false,
	)); err != nil {
		return err
	}
	// Set a default transparent background
	if err := c.client.Emulation.SetDefaultBackgroundColorOverride(c.ctx, emulation.NewSetDefaultBackgroundColorOverrideArgs().
		SetColor(dom.RGBA{R: 0,G: 0,B: 0,A: utils.Float64(0.0)}),
	); err != nil {
		return err
	}

	// Create a messageAdded client
	messageAdded, err := c.client.Console.MessageAdded(c.ctx)
	if err != nil {
		return err
	}

	mydir, err := os.Getwd()
	if err != nil {
			fmt.Println(err)
	}
	fmt.Println(mydir)

	// Handle added messages
	go func() {
		defer messageAdded.Close()

		for {
			ev, err := messageAdded.Recv()
			if err != nil {
				c.abort <- fmt.Errorf("failed to receive MessageAdded: %v", err)
			}

			message := AnimationMessage{}
			if err := json.Unmarshal([]byte(ev.Message.Text), &message); err != nil {
				c.abort <- fmt.Errorf("failed to unmarshal AnimationMessage: %v", err)
			}

			switch message.Type {
			case "message":
				c.AnimationMessages <- message
			case "command":
				c.AnimationCommands <- message
			}
		}
	}()

	return nil
}

// abortOnErrors handles errors occured while navigating the page
func (c Chrome) abortOnErrors() error {
	exceptionThrown, err := c.client.Runtime.ExceptionThrown(c.ctx)
	if err != nil {
		return err
	}

	loadingFailed, err := c.client.Network.LoadingFailed(c.ctx)
	if err != nil {
		return err
	}

	go func() {
		defer exceptionThrown.Close()
		defer loadingFailed.Close()
		for {
			select {
			// Check for exceptions so we can abort as soon
			// as one is encountered.
			case <- exceptionThrown.Ready():
				ev, err := exceptionThrown.Recv()
				if err != nil {
					// This could be any one of: stream closed,
					// connection closed, context deadline or
					// unmarshal failed.
					c.abort <- err
					return
				}

				// Ruh-roh! Let the caller know something went wrong.
				c.abort <- ev.ExceptionDetails

			// Check for non-canceled resources that failed
			// to load.
			case <- loadingFailed.Ready():
				ev, err := loadingFailed.Recv()
				if err != nil {
					c.abort <- err
					return
				}

				// For now, most optional fields are pointers
				// and must be checked for nil.
				canceled := ev.Canceled != nil && *ev.Canceled

				if !canceled {
					c.abort <- fmt.Errorf("request %s failed: %s", ev.RequestID, ev.ErrorText)
				}
			}
		}
	}()

	return nil
}

// StartRecording starts recording the page using the provided method.
func (c Chrome) StartRecording(method string) error {
	switch method {
	case "screencast":
		if err := c.StartScreencast(); err != nil {
			return err
		}
	case "timeweb":
		if err := c.StartTimeweb(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid method %s", method)
	}
	return nil
}

// StartScreencast send an event through the CDP client to start sending
// screeencast frames.
func (c Chrome) StartScreencast() error {
	screencastArgs := page.NewStartScreencastArgs().SetFormat("png").SetEveryNthFrame(1).SetQuality(100)
	err := c.client.Page.StartScreencast(c.ctx, screencastArgs)
	if err != nil {
		c.abort <- fmt.Errorf("failed to start page screencast: %v", err)
		return err
	}

	// Create a screencastFrame client
	screencastFrame, err := c.client.Page.ScreencastFrame(c.ctx)
	if err != nil {
		return err
	}

	// Handle sceencast frames
	go func() {
		defer screencastFrame.Close()

		for {
			ev, err := screencastFrame.Recv()
			if err != nil {
				c.abort <- fmt.Errorf("failed to receive ScreencastFrame: %v", err)
			}

			err = c.client.Page.ScreencastFrameAck(c.ctx, page.NewScreencastFrameAckArgs(ev.SessionID))
			if err != nil {
				c.abort <- fmt.Errorf("failed to ack ScreencastFrame: %v", err)
			}

			c.Frames <- ev.Data
		}
	}()

	return nil
}

// StartTimeweb starts moving the time forward frame by frame, taking
// a screenshot on every move.
func (c *Chrome) StartTimeweb() error {
	c.stop = make(chan bool, 1)

	go func() {
		screenshotArgs := page.NewCaptureScreenshotArgs().SetFormat("png")

		timeweb:
		for {
			select {
			case <- c.stop:
				break timeweb
			default:
				// Take screenshot
				screenshot, err := c.client.Page.CaptureScreenshot(c.ctx, screenshotArgs)
				if err != nil {
					c.abort <- err
				}
				// Send frame to the Frames channel
				c.Frames <- screenshot.Data
				// Move the time one frame further
				c.nextVirtualTime()
				c.Evaluate(fmt.Sprintf("timeweb.goTo(%d)", c.virtualTime))
			}
		}
	}()

	return nil
}

// StopRecording stops the current recording.
func (c Chrome) StopRecording(method string) error {
	switch method {
	case "screencast":
		if err := c.StopScreencast(); err != nil {
			return err
		}
	case "timeweb":
		if err := c.StopTimeweb(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid method %s", method)
	}
	return nil
}

// StopScreencast send an event through the CDP client to stop sending
// screeencast frames.
func (c Chrome) StopScreencast() error {
	err := c.client.Page.StopScreencast(c.ctx)
	if err != nil {
		c.abort <- fmt.Errorf("failed to stop page screencast %v", err)
		return err
	}
	return nil
}

// StopTimeweb stops the moving of time on the page using timeweb.
func (c *Chrome) StopTimeweb() error {
	c.stop <- true

	return nil
}

// Navigate navigates Chrome to the URL and waits for DOMContentEventFired.
// An error is returned if timeout happens before DOMContentEventFired.
func (c Chrome) Navigate(url string, timeout time.Duration) error {
	var cancel context.CancelFunc
	c.ctx, cancel = context.WithTimeout(c.ctx, timeout)
	c.cancelFuncs = append(c.cancelFuncs, func() error {cancel(); return nil})

	// Make sure Page events are enabled.
	err := c.client.Page.Enable(c.ctx)
	if err != nil {
		return err
	}

	// Open client for DOMContentEventFired to block until DOM has fully loaded.
	domContentEventFired, err := c.client.Page.DOMContentEventFired(c.ctx)
	if err != nil {
		return err
	}
	defer domContentEventFired.Close()

	_, err = c.client.Page.Navigate(c.ctx, page.NewNavigateArgs(url))
	if err != nil {
		return err
	}

	_, err = domContentEventFired.Recv()
	return err
}

// Evaluate allows the evaluation of an expression on the global scope.
func (c Chrome) Evaluate(expression string) {
	evalArgs := runtime.NewEvaluateArgs(expression)
	_, err := c.client.Runtime.Evaluate(c.ctx, evalArgs)
	if err != nil {
		c.abort <- fmt.Errorf("failed to Evaluate expression: %v", err)
	}
}

func (c *Chrome) nextVirtualTime() {
	c.virtualTime += int(math.Floor(1000 / c.fps))
}