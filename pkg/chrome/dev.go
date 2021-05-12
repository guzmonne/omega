package chrome

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"sync"
	"syscall"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/evanw/esbuild/pkg/api"
	"gux.codes/omega/pkg/utils"
)

type DevelopmentEnvironment interface {
	// SetScript sets the current development script
	SetScript(script string)
	// GetScript gets the current development script
	GetScript() string
}

type Dev struct {
	// mu mutex protects the rest of the properties
	mu sync.Mutex
	// Script stores the current built script
	Script string
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

var D DevelopmentEnvironment

func init() {
	D = &Dev{Script: `console.log("Hello, World!")`}
}

func NewDev(entryPoint string) error {
	// Clear the screen
	// shell.Clear()
	utils.Info("Starting the development environment...")

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
    chromedp.Flag("headless", false),
    chromedp.Flag("disable-gpu", false),
    chromedp.Flag("enable-automation", false),
    chromedp.Flag("disable-extensions", false),
	)

	// Start the web server on a different goroutine
	webServerOptions := NewWebServerOptions()
	go Serve(webServerOptions)
	utils.Info("Starting the webserver...")

	// Create the done channel
	doneCh := make(chan bool, 1)
	defer close(doneCh)

	// Start a goroutine to rebuild the files
	changeCh, reloadCh, errorsCh := build(doneCh, entryPoint)

	// Create custom allocator
	actx, acancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer acancel()

	// Create context
	ctx, cancel := chromedp.NewContext(actx, chromedp.WithLogf(log.Printf))
	defer cancel()
	utils.Info("Starting the browser...")

	// Setup a Ctrl+C handler
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func(){
		<- signals
		doneCh <- true
		close(doneCh)
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		acancel()
		cancel()
		os.Exit(0)
	}()

	// Start the listener on the reload channel
	go reloadListener(ctx, doneCh, reloadCh)

	// Start the listener on the errors channel
	go errorsListener(doneCh, errorsCh)

	// Start the listener on the change channel
	go changesListener(doneCh, entryPoint, changeCh)

	// Redirect to the dev site
	if err := chromedp.Run(ctx, chromedp.Navigate(fmt.Sprintf("http://localhost:%d/dev", webServerOptions.Port))); err != nil {
		log.Fatal(err)
	}
	utils.BoxGreen("Navigating to the development site...")

	// Block until something happens
	<- doneCh
	close(doneCh)

	return nil
}

func errorsListener(doneCh <-chan bool, errorsCh <-chan []api.Message) {
	for {
		select {
		case <- doneCh:
			return
		case errs := <- errorsCh:
			if errs == nil {
				continue
			}
			// shell.Clear()
			utils.Error("Some errors were encountered while building...")
			for index, err := range errs {
				utils.BoxRed(fmt.Sprintf("Error #%d", index))
				fmt.Printf( "Location: %s\n", err.Location.File)
				fmt.Printf( "Line Nr : %d\n", err.Location.Line)
				fmt.Println("Detail  :")
				fmt.Println(err.Detail)
				fmt.Println()
			}
		}
	}
}

func changesListener(doneCh <-chan bool, entryPoint string, changeCh chan<- bool) {
	var paths map[string]time.Time

	for {
		select {
		case <- doneCh:
			return
		case <- time.After(2 * time.Second):
			newPaths := make(map[string]time.Time)
			cwd, err := os.Getwd()
			if err != nil {
				log.Println(err)
			}
			folder := filepath.Join(cwd, entryPoint)
			err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				newPaths[path] = info.ModTime()
				return nil
			})
			if err != nil {
				log.Println(err)
			}
			if paths != nil && !reflect.DeepEqual(newPaths, paths) {
				utils.Info("A new update was found")
				changeCh <- true
			}
			paths = newPaths
		}
	}
}

func reloadListener(ctx context.Context, doneCh <-chan bool, reloadCh <-chan bool) {
	for {
		select  {
		case <-doneCh:
			return
		case <-reloadCh:
			chromedp.Run(ctx, chromedp.Reload())
			utils.Info("Reloading...")
		}
	}
}

func build(doneCh <-chan bool, entryPoint string) (chan<- bool, <-chan bool, <-chan []api.Message) {
	changeCh := make(chan bool)
	reloadCh := make(chan bool)
	errorsCh := make(chan []api.Message)

	result := api.Build(api.BuildOptions{
		EntryPoints      : []string{entryPoint},
		Bundle           : true,
		MinifyWhitespace : false,
		MinifyIdentifiers: false,
		MinifySyntax     : false,
		Engines          : []api.Engine{{Name: api.EngineChrome, Version: "91"}},
		Write            : false,
	})

	if len(result.Errors) > 0 {
		log.Fatalf("%+v", result.Errors)
	} else {
		D.SetScript(string(result.OutputFiles[0].Contents))
	}

	go func() {
		for {
			select {
			case <- doneCh:
				close(changeCh)
				close(reloadCh)
				close(errorsCh)
				return
			case <- changeCh:
				result := api.Build(api.BuildOptions{
					EntryPoints      : []string{entryPoint},
					Bundle           : true,
					MinifyWhitespace : false,
					MinifyIdentifiers: false,
					MinifySyntax     : false,
					Engines          : []api.Engine{{Name: api.EngineChrome, Version: "91"}},
					Write            : false,
				})

				utils.Info("Compiling new version")

				if len(result.Errors) > 0 {
					errorsCh <- result.Errors
				} else {
					D.SetScript(string(result.OutputFiles[0].Contents))
					// shell.Clear()
					utils.Success("Compilation is done!")
					reloadCh <- true
				}
			}
		}
	}()

	return changeCh, reloadCh, errorsCh
}