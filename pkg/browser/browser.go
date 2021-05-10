package browser

import (
	"context"

	"github.com/chromedp/cdproto/cdp"
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
	Navigate(ctx context.Context, urlstr string) error
	// Screenshot takes a screensho of the context's viewport, trimmed according to the provided params.
	Screenshot(ctx context.Context, viewport page.Viewport) ([]byte, error)
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
func (ChromeBrowser) Navigate(ctx context.Context, urlstr string) error {
	return chromedp.Run(ctx, chromedp.Navigate(urlstr))
}

// Screenshot takes a screenshot of what is being shown on the current browser's viewport
// cropped according to the provided coordinates.
func (ChromeBrowser) Screenshot(ctx context.Context, viewport page.Viewport) ([]byte, error) {
	var buf []byte
	var err error
	return buf, chromedp.Run(ctx, chromedp.Tasks{
		emulation.SetDefaultBackgroundColorOverride().
			WithColor(&cdp.RGBA{R: 0,G: 0,B: 0,A: 0.0}),
		emulation.SetDeviceMetricsOverride(int64(viewport.Width), int64(viewport.Height), 1, false).
			WithScreenOrientation(&emulation.ScreenOrientation{
				Type: emulation.OrientationTypePortraitPrimary,
				Angle: 0,
			}),
		chromedp.ActionFunc(func (ctx context.Context) error {
			buf, err = page.CaptureScreenshot().
				WithFormat(page.CaptureScreenshotFormatPng).
				WithClip(&page.Viewport{
					X: viewport.X,
					Y: viewport.Y,
					Width: float64(viewport.Width),
					Height: float64(viewport.Height),
					Scale: viewport.Scale,
				}).
				Do(ctx)
			return err
		}),
	})
}

// Chrome is a package-level variable of type BrowserHandle to hold a reference to ChromeBrowser
var Chrome BrowserHandler

// init initializes the brow package variable
func init() {
	Chrome = &ChromeBrowser{}
}