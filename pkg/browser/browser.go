package browser

import (
	"context"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// browser abstracts the communication and handling of a browser instance.
type BrowserHandler interface {
	// NewContext opens a new browser instance if none were defined on the parent context.
	NewContext(parent context.Context) (context.Context, context.CancelFunc)
	// Evaluate evaluates any JavaScript script on the current context, and returns its output.
	Evaluate(ctx context.Context, script string) ([]byte, error)
	// Navigate navigates the current context to the provided url
	Navigate(ctx context.Context, urlstr string, width, height int64) error
	// Screenshot takes a screensho of the context's viewport, trimmed according to the provided params.
	Screenshot(ctx context.Context) ([]byte, error)
}

// ChromeBrowser is an implementation of the browserHandler interface to interact with a Chrome
// browser through its devtools protocol.
type ChromeBrowser struct {}

// NewContext Creates a new Browser context.
func (ChromeBrowser) NewContext(parent context.Context) (context.Context, context.CancelFunc) {
	return chromedp.NewContext(parent)
}

// Evaluate evaluates a JavaScript script on the browser page, and returns its output.
func (ChromeBrowser) Evaluate(ctx context.Context, script string) ([]byte, error) {
	var res []byte
	return res, chromedp.Run(ctx, chromedp.Evaluate(script, &res))
}

// Navigate navigates the current browser to the provided url.
func (ChromeBrowser) Navigate(ctx context.Context, urlstr string, width, height int64) error {
	return chromedp.Run(ctx, chromedp.Tasks{
		emulation.SetDeviceMetricsOverride(width, height, 1, false).
		WithScreenOrientation(&emulation.ScreenOrientation{
			Type: emulation.OrientationTypePortraitPrimary,
			Angle: 0,
		}),
		chromedp.Navigate(urlstr),
	})
}

// Screenshot takes a screenshot of what is being shown on the current browser's viewport
// cropped according to the provided coordinates.
func (ChromeBrowser) Screenshot(ctx context.Context) ([]byte, error) {
	var buf []byte
	return buf, chromedp.Run(ctx, screenshot(&buf))
}

// Chrome is a package-level variable of type BrowserHandle to hold a reference to ChromeBrowser
var Chrome BrowserHandler

// init initializes the brow package variable
func init() {
	Chrome = &ChromeBrowser{}
}

// Screenshot Action Task
func screenshot(buf *[]byte) chromedp.Tasks {
	return chromedp.Tasks{chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		*buf, err = page.CaptureScreenshot().
			WithFormat(page.CaptureScreenshotFormatPng).
			Do(ctx)
		return err
	})}
}