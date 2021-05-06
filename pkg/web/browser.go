package web

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type Browser struct {
	webServerOptions WebServerOptions
	ready bool
}

// Opens the Browser and starts the handler web server.
func (b *Browser) Open(parent context.Context, url string) (context.Context, context.CancelFunc) {
	ctx, cancel := chromedp.NewContext(parent)

	// Fail if this function is called more than once.
	if b.ready {
		log.Fatal(errors.New("browser has already been initialized"))
	}

	// Set the browser as ready
	b.ready = true

	// Navigate to the url
	if err := chromedp.Run(ctx, chromedp.Navigate(url)); err != nil {
		log.Fatal(err)
	}

	return ctx, cancel
}

// Takes a screenshot of the active handler
func (b *Browser) Screenshot(ctx context.Context) ([]byte, error) {
	var frame []byte

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