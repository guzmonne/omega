package web

import (
	"context"
	"io/ioutil"

	"github.com/chromedp/chromedp"
	"gux.codes/omega/pkg/utils"
)

func Screenshot() error {
	// Create the browser
	browser := NewBrowser()

	// Open the browser
	ctx, cancel := browser.Open(context.Background())
	defer cancel()

	// Open a new handler
	ctx, _ = browser.NewHandler(ctx)

	// Take the screenshot
	buf, err := browser.Screenshot(ctx)
	if err != nil {
		return err
	}

	// Store the screenshot
	if err := ioutil.WriteFile("/tmp/" + utils.ULID() + ".png", buf, 0644); err != nil {
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

