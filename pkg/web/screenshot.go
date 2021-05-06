package web

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/chromedp/chromedp"
	"gux.codes/omega/pkg/utils"
)

func Screenshot() error {
	// Create the browser
	browser := NewBrowser()

	// Start the web server on a different goroutine
	webServerOptions := NewWebServerOptions()
	go Serve(webServerOptions)

	// Create the handler url
	urlstr := fmt.Sprintf("http://localhost:%d/handler", webServerOptions.Port)

	// Open the browser
	ctx, cancel := browser.Open(context.Background(), urlstr)
	defer cancel()

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

