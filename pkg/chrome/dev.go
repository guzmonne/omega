package chrome

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/chromedp/chromedp"
	"github.com/evanw/esbuild/pkg/api"
	"gux.codes/omega/pkg/esbuild"
	"gux.codes/omega/pkg/utils"
)

type DevelopmentEnvironment interface {
	// SetScript sets the current development script
	SetScript(script string)
	// GetScript gets the current development script
	GetScript() string
	// SetStyle sets the current development script
	SetStyles(style string)
	// GetStyle gets the current development script
	GetStyles() string
}

type Dev struct {
	// mu mutex protects the rest of the properties
	mu sync.Mutex
	// Script stores the current built script
	Script string
	// Styles stores the current build styles
	Styles string
}

// SetScript sets the current development script safely
func (d *Dev) SetScript(script string) {
	d.mu.Lock()
	d.Script = script
	d.mu.Unlock()
}

// GetScript gets the current development script safely
func (d *Dev) GetScript() string {
	return d.Script
}

// SetStyles sets the current development script safely
func (d *Dev) SetStyles(script string) {
	d.mu.Lock()
	d.Styles = script
	d.mu.Unlock()
}

// GetStyles gets the current development script safely
func (d *Dev) GetStyles() string {
	return d.Styles
}

var D DevelopmentEnvironment

func init() {
	D = &Dev{Script: `console.log("Hello, World!")`}
}

func NewDev(entryPoint string) error {
	// Clear the screen
	utils.Info("Starting the development environment...")

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
    chromedp.Flag("headless", false),
    chromedp.Flag("disable-gpu", false),
    chromedp.Flag("enable-automation", false),
    chromedp.Flag("disable-extensions", false),
    chromedp.Flag("auto-open-devtools-for-tabs", true),
    chromedp.Flag("enable-kiosk-mode", true),
    chromedp.Flag("window-size", "1200,700"),
	)

	// Start the web server on a different goroutine
	webServerOptions := NewWebServerOptions()
	go Serve(webServerOptions)
	utils.Info("Starting the webserver...")

	// Create the done channel
	doneCh := make(chan bool, 1)

	// Create custom allocator
	actx, acancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer acancel()

	// Create context
	ctx, cancel := chromedp.NewContext(actx, chromedp.WithLogf(log.Printf))
	defer cancel()
	utils.Info("Starting the browser...")

	// Create the build struct
	build := esbuild.NewBuild(entryPoint)

	// Run an initial build
	result := build.Run()
	buildDone(result)
	utils.Info("Initial build done")

	// Redirect to the dev site
	if err := chromedp.Run(ctx, chromedp.Navigate(fmt.Sprintf("http://localhost:%d/dev", webServerOptions.Port))); err != nil {
		log.Fatal(err)
	}
	utils.BoxGreen("Navigating to the development site...")

	stop := build.WithWatch(func (result api.BuildResult) {
		utils.Info("Build done")
		buildDone(result)
		chromedp.Run(ctx, chromedp.Reload())
		utils.Info("Reloading...")
	}).Run().Stop

	// Setup the Ctrl+C handler
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func(){
		<- signals
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		acancel()
		cancel()
		close(doneCh)
		stop()
		os.Exit(0)
	}()

	// Block until something happens
	<- doneCh

	return nil
}

func buildDone(result api.BuildResult) {
	if len(result.Errors) > 0 {
		utils.Error("Some errors were encountered while building...")
		for _, err := range result.Errors {
			fmt.Printf("Location: at %s on line %d\n", err.Location.File, err.Location.Line)
			fmt.Printf("Reason  : %s\n", err.Text)
		}
		return
	}
	if len(result.OutputFiles) > 0 {
		D.SetScript(string(result.OutputFiles[0].Contents))
	}
	if len(result.OutputFiles) > 1 {
		D.SetStyles(string(result.OutputFiles[1].Contents))
	}
}
