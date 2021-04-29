package web

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"gux.codes/omega/pkg/utils"
)

func Screenshot() error {
	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)
	defer cancel()

	// Run the web server
	webServerOptions := NewWebServerOptions()
	go startWebServer(*webServerOptions)

	// Navigate to the URL
	url := fmt.Sprintf("http://localhost:%d/handler", webServerOptions.Port)
	if err := navigate(ctx, url); err != nil {
		return err
	}

	// Take the screenshot
	var buf []byte
	if err := takeScreenshot(ctx, &buf, 1920, 1080); err != nil {
		return err
	}

	// Store the screenshot
	if err := ioutil.WriteFile("/tmp/" + utils.ULID() + ".png", buf, 0644); err != nil {
		return err
	}

	return nil
}

func takeScreenshot(ctx context.Context, buf *[]byte, width int64, height int64) error {
	// Take the screenshot
	if err := chromedp.Run(ctx, screenshot(width, height, buf)); err != nil {
		return err
	}

	return nil
}

// navigate makes the current context browser go to the provided URL.
func navigate(ctx context.Context, url string) error {
	if err := chromedp.Run(ctx, chromedp.Tasks{chromedp.Navigate(url)}); err != nil {
		return err
	}

	return nil
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