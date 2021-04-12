package record

import (
	"context"
	"io/ioutil"
	"math"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"gux.codes/omega/pkg/utils"
)

func screenshot() chromedp.Tasks {
	return chromedp.Tasks{
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
			res, err := page.CaptureScreenshot().
			//_, err = page.CaptureScreenshot().
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

			if err := ioutil.WriteFile(utils.ULID() + ".png", res, 0o644); err != nil {
				return err
			}

			return nil
		}),
	}
}