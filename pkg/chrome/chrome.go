package chrome

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/fatih/color"
)

func ScreenShot() error {
	// Create context
	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)
	defer cancel()

	// Create function to print green colors
	green := color.New(color.FgGreen).SprintFunc()

	// Capture the screenshot of an element
	var buf []byte
	if err := chromedp.Run(ctx, elementScreenshot(`https://pkg.go.dev/`, `img.Homepage-logo`, &buf)); err != nil {
		return err
	}
	if err := ioutil.WriteFile("elementScreenshot.png", buf, 0o644); err != nil {
		return err
	}

	fmt.Printf("Took a screenshot of an element. Take a look:\n%s\n", green("./elementScreenshot.png"))

	// Capture entire browser viewport, returning with quality=90
	if err := chromedp.Run(ctx, fullScreenshot(`https://www.conatel.com.uy`, 90, &buf)); err != nil {
		return err
	}
	if err := ioutil.WriteFile("fullScreenshot.png", buf, 0o644); err != nil {
		return err
	}

	fmt.Printf("Took a screenshot of a page. Take a look:\n%s\n", green("./fullScreenshot.png"))

	return nil
}

func elementScreenshot(urlstr, sel string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.Screenshot(sel, res, chromedp.NodeVisible),
	}
}

func fullScreenshot(urlstr string, quality int64, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Get layout metrics
			_, _, cssContentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := int64(math.Ceil(cssContentSize.Width)), int64(math.Ceil(cssContentSize.Height))

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

			// Capture the screenshot
			*res, err = page.CaptureScreenshot().
				WithQuality(quality).
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