package web

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

type Browser struct {
	webServerOptions WebServerOptions
	ready bool
}

// Opens the Browser and starts the handler web server.
func (b *Browser) Open(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)

	// Fail if this function is called more than once.
	if b.ready {
		log.Fatal(errors.New("browser has already been initialized"))
	}

	// Start the web server on a different goroutine
	webServerOptions := NewWebServerOptions()
	go Serve(webServerOptions)

	// Set the browser as ready
	b.ready = true

	// Setup a Ctrl+C handler
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func(){
		<- signals
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		cancel()
		os.Exit(0)
	}()

	return ctx, cancel
}

// Opens a new handler tab, and waits until the handler is ready to be recorded.
func (b *Browser) NewHandler(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := chromedp.NewContext(parent)

	// ensure the tab is created
	if err := chromedp.Run(ctx); err != nil {
		log.Fatal(err)
	}

	// Ensure the tab is created.
	urlstr := fmt.Sprintf("http://localhost:%d/handler", b.webServerOptions.Port)
	if err := chromedp.Run(ctx, chromedp.Navigate(urlstr)); err != nil {
		log.Fatal(err)
	}

	return ctx, cancel
}

// Handlers lists all the handlers in the browser attached to the given context.
func (b *Browser) Handlers(ctx context.Context) ([]*target.Info, error) {
	return chromedp.Targets(ctx)
}

// GetHandler gets a single handler from the handlers in the browser attached to the given context
// identified by its index.
func (b *Browser) GetHandler(ctx context.Context, index int) (*target.Info, error) {
	handlers, err := b.Handlers(ctx)
	if err != nil {
		fmt.Println("browser.GetHandler error")
		return nil, err
	}

	if len(handlers) <= index {
		fmt.Println("browser.GetHandler error")
		return nil, errors.New("handler index out of range")
	}

	return handlers[index], nil
}

// GetHandlerContext returns a context configured to interact with the handler identified
// by its index.
func (b *Browser) GetHandlerContext(parent context.Context, index int) (context.Context, context.CancelFunc) {
	handler, err := b.GetHandler(parent, index)
	if err != nil {
		fmt.Println("browser.GetHandlerContext error")
		log.Fatal(err)
	}
	return chromedp.NewContext(parent, chromedp.WithTargetID(handler.TargetID))
}

// Takes a screenshot of the active handler
func (b *Browser) Screenshot(ctx context.Context) ([]byte, error) {
	var frame []byte

	//ctx, _ := b.GetHandlerContext(parent, index)

	err := chromedp.Run(ctx, chromedp.Tasks{screenshot(1920, 1080, &frame)})
	if err != nil {
		fmt.Println("browser.Screenshot error")
		return nil, err
	}

	return frame, err
}

// Move the active handler to the provided vt. The virtual time should always increment.
func (b *Browser) GoTo(ctx context.Context, vt int) ([]byte, error) {
	var res []byte

	//ctx, _ := b.GetHandlerContext(parent, index)

	err := chromedp.Run(ctx, chromedp.Evaluate(fmt.Sprintf("timeweb.goTo(%d)", vt), &res))
	if err != nil {
		fmt.Println("browser.GoTo error")
		return nil, err
	}

	return res, err
}

// NewBrowser creates a new Browser.
func NewBrowser() Browser {
	webServerOptions := NewWebServerOptions()

	return Browser{
		webServerOptions: webServerOptions,
		ready: false,
	}
}

// screenshot will take a screenshot of the current page
func screenshot(width int64, height int64, buf *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func (ctx context.Context) error {
			// Set a default transparent background
			err := emulation.SetDefaultBackgroundColorOverride().
				WithColor(&cdp.RGBA{R: 0,G: 0,B: 0,A: 0.0}).
				Do(ctx)
			if err != nil {
				return err
			}

			// Force viewport emulation
			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type: emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			// Get layout metrics
			_, _, cssContentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			*buf, err = page.CaptureScreenshot().
				WithFormat(page.CaptureScreenshotFormatPng).
				WithClip(&page.Viewport{
					X: cssContentSize.X,
					Y: cssContentSize.Y,
					Width: cssContentSize.Width,
					Height: cssContentSize.Height,
					Scale: 1,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			return nil
		}),
	}
}