package record

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"github.com/gobs/simplejson"
	"github.com/raff/godet"
	"gux.codes/omega/pkg/chrome"
	"gux.codes/omega/pkg/utils"
)

type AnimationMessage struct {
	Session int `json:"session"`
	Type string `json:"type"`
	Message string `json:"message"`
	Action string `json:"action"`
}

type AnimationOptions struct {
	Session int
	Port int
}

func NewAnimationOptions() *AnimationOptions {
	return &AnimationOptions{
		Session: rand.Int(),
		Port: 38080,
	}
}

func Chrome() error {
	var remote *godet.RemoteDebugger
	var err error

	err = chrome.Start()
	if err != nil {
		return err
	}

	// Start the HTTP server
	options := NewAnimationOptions()
	go startServer(options)

	for i := 0; i < 30; i++ {
		if i > 0 {
			time.Sleep(500 * time.Millisecond)
		}
		remote, err = godet.Connect("localhost:9222", false)
		if err == nil {
			break
		}
	}
	if err != nil {
		return err
	}
	defer remote.Close()

	// Get list of open tabs
	tabs, err := remote.TabList("")
	if err != nil {
		return err
	}

	// navigate in existing tab
	_ = remote.ActivateTab(tabs[0])

	// Enable event processing
	remote.PageEvents(true)
	remote.AllEvents(true)

	exception := make(chan bool, 1)
	close := make(chan bool, 1)

	remote.CallbackEvent("Page.screencastFrame", func (params godet.Params) {
		message, err := parseParams(params)
		if err != nil {
			fmt.Printf("%s parseParams\n%s\n", utils.BoxRed("Error:"), err.Error())
			exception <- true
		}
		if message.Type != "command" {
			fmt.Printf("%s %s\n", utils.BoxGreen(message.Type), message.Message)
			go takeScreenshot(remote)
			return
		}
		fmt.Printf("%s %s\n", utils.BoxRed(message.Type), message.Message)
		switch message.Action {
		case "close":
			close <- true
		}
	})

	urlstr := fmt.Sprintf("http://localhost:%d/handler", options.Port)
	_, err = remote.Navigate(urlstr)
	if err != nil {
		return err
	}

	p, err := remote.SendRequest("Page.startScreenCast", godet.Params{})
	if err != nil {
		return err
	}

	fmt.Println(p)

	select {
	case <- exception:
	case <- close:
	}

	return nil
}

func documentNode(remote *godet.RemoteDebugger) (int, error) {
	res, err := remote.GetDocument()
	if err != nil {
		return -1, err
	}

	doc := simplejson.AsJson(res)
	return doc.GetPath("root", "nodeId").MustInt(-1), nil
}


func takeScreenshot(remote *godet.RemoteDebugger) {
	remote.SaveScreenshot(utils.ULID() + ".png", 0644, 0, true)
}

func startScreencast(options *AnimationOptions) chromedp.Tasks {
	urlstr := fmt.Sprintf("http://localhost:%d/handler", options.Port)
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			page.StartScreencast().Do(ctx)

			chromedp.ListenTarget(ctx, func(ev interface{}) {
				switch ev := ev.(type) {
				case *runtime.EventConsoleAPICalled:
					go handleEventConsoleApiCalled(ev)
				case *page.EventScreencastVisibilityChanged:
				case *page.EventScreencastFrame:
					go handleEventScreencastFrame(ev)
					page.ScreencastFrameAck(ev.SessionID).Do(ctx)
				}

			})

			return nil
		}),
	}
}

func handleEventScreencastFrame(ev *page.EventScreencastFrame) {
	unbased, err := base64.StdEncoding.DecodeString(ev.Data)
	if err != nil {
		fmt.Printf("%s base64.StdEncoding.DecodeString\n%s", utils.BoxRed("Error:"), err.Error())
		//exception <- true
	}

	if err := ioutil.WriteFile(utils.ULID() + ".png", unbased, 0o644); err != nil {
		fmt.Printf("%s ioutil.WriteFile\n%s", utils.BoxRed("Error:"), err.Error())
		//exception <- true
	}
}

func parseParams(params godet.Params) (AnimationMessage, error) {
	message := AnimationMessage{}
	l := []string{}
	for _, a := range params["args"].([]interface{}) {
		arg := a.(map[string]interface{})
		if arg["value"] != nil {
			l = append(l, arg["value"].(string))
		} else {
			l = append(l, arg["type"].(string))
		}
	}
	value := l[0]
	if err := json.Unmarshal([]byte(value), &message); err != nil {
		return message, err
	}
	return message, nil
}

func handleEventConsoleApiCalled(ev *runtime.EventConsoleAPICalled) {
	if ev.Type != "info" || len(ev.Args) != 1 {
		return
	}
	message := AnimationMessage{}
	value, err := strconv.Unquote(string(ev.Args[0].Value))
	if err != nil {
		fmt.Printf("%s strconv.Unquote\n%s", utils.BoxRed("Error:"), err.Error())
		//exception <- true
	}
	if err := json.Unmarshal([]byte(value), &message); err != nil {
		fmt.Printf("%s json.Unmarshal\n%s", utils.BoxRed("Error:"), err.Error())
		//exception <- true
	}
	if message.Type != "command" {
		fmt.Printf("%s %s\n", utils.BoxGreen(message.Type), message.Message)
		return
	}
	fmt.Printf("%s %s\n", utils.BoxRed(message.Type), message.Message)
	switch message.Action {
	case "close":
		//close <- true
	}
}

/*
func Chrome() error {
	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)
	defer cancel()

	// Start the HTTP server
	options := NewAnimationOptions()
	go startServer(options)

	// Create signal channels to stop the execution when an Exception
	// is thrown or when the animation is done.
	exception := make(chan bool, 1)
	close := make(chan bool, 1)

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			if ev.Type != "info" || len(ev.Args) != 1 {
				break
			}
			message := AnimationMessage{}
			value, err := strconv.Unquote(string(ev.Args[0].Value))
			if err != nil {
				fmt.Printf("%s strconv.Unquote\n%s", utils.BoxRed("Error:"), err.Error())
				exception <- true
				break
			}
			if err := json.Unmarshal([]byte(value), &message); err != nil {
				fmt.Printf("%s json.Unmarshal\n%s", utils.BoxRed("Error:"), err.Error())
				exception <- true
				break
			}
			if message.Type == "command" {
				fmt.Printf("%s %s\n", utils.BoxRed(message.Type), message.Message)
				switch message.Action {
				case "start":
					page.StartScreencast()
				case "close":
					page.StopScreencast()
					close <- true
				}
			} else {
				fmt.Printf("%s %s\n", utils.BoxGreen(message.Type), message.Message)
			}
		case *runtime.EventExceptionThrown:
			s := ev.ExceptionDetails.Error()
			fmt.Printf("* %s \n", s)
			exception <- true
		case *page.EventScreencastFrame:
			go func() {
				unbased, err := base64.StdEncoding.DecodeString(ev.Data)
				if err != nil {
					fmt.Printf("%s base64.StdEncoding.DecodeString\n%s", utils.BoxRed("Error:"), err.Error())
					exception <- true
				}

				if err := ioutil.WriteFile(utils.ULID() + ".png", unbased, 0o644); err != nil {
					fmt.Printf("%s ioutil.WriteFile\n%s", utils.BoxRed("Error:"), err.Error())
					exception <- true
				}
			}()
			page.ScreencastFrameAck(ev.SessionID).Do(ctx)
		}
	})

	target.

	urlstr := fmt.Sprintf("http://localhost:%d/handler", options.Port)
	if err := chromedp.Run(ctx,
		chromedp.Navigate(urlstr),
	); err != nil {
		return err
	}

	select {
	case <- exception:
	case <- close:
	}

	ctx.Done()

	return nil
}

*/
func startServer(options *AnimationOptions) {
	// Force logs in color
	gin.ForceConsoleColor()
	// Set the "release" mode
	gin.SetMode(gin.ReleaseMode)
	// Create the default router
	router := gin.Default()
	// Load the templates
	router.LoadHTMLGlob("templates/*")
	// Create the routes
	router.GET("/handler", func(c *gin.Context) {
		c.HTML(http.StatusOK, "handler.tmpl", options)
	})
	// Run the server
	router.Run(fmt.Sprintf(":%d", options.Port))
}