package record

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gux.codes/omega/pkg/chrome"
	"gux.codes/omega/pkg/utils"
)

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
	c := chrome.New()

	err := c.Open(chrome.NewOpenOptions())
	if err != nil {
		return err
	}
	defer c.Close()

	// Start the HTTP server
	go startServer(NewAnimationOptions())

	// Create hold channel
	close := make(chan bool, 1)

	// Setup the message handler
	go func() {
		for message := range c.AnimationMessages {
			switch message.Type {
			case "command":
				switch message.Action {
				case "start":
					c.StartScreencast()
				case "close":
					c.StopScreencast()
					close <- true
				}
				fmt.Printf("%s %s\n", utils.BoxRed("command"), message.Message)
			case "message":
				fmt.Printf("%s %s\n", utils.BoxGreen("message"), message.Message)
			}
		}
	}()

	// Handle screencastFrames
	screencastFrames := make([][]byte, 0)
	go func() {
		for screencastFrame := range c.ScreencastFrames {
			fmt.Printf("%s %s\n", utils.BoxGreen("message"), screencastFrame.Metadata.Timestamp)
			screencastFrames = append(screencastFrames, screencastFrame.Data)
		}
	}()

	// Navigate to handler URL
	options := NewAnimationOptions()
	urlstr := fmt.Sprintf("http://localhost:%d/handler", options.Port)
	timeout := 1 * time.Second
	if err := c.Navigate(urlstr, timeout); err != nil {
		return err
	}

	<- close


	for index, value := range screencastFrames {
		err = ioutil.WriteFile(fmt.Sprintf("/tmp/%d.png", index), value, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}



func startServer(options *AnimationOptions) {
	// Force logs in color
	gin.ForceConsoleColor()
	// Set the "release" mode
	gin.SetMode(gin.ReleaseMode)
	// Create the default router
	router := gin.Default()
	// Load the templates
	router.LoadHTMLGlob("templates/*")
	// Serve static assets
	router.Static("/assets", "./assets")
	// Create the routes
	router.GET("/handler", func(c *gin.Context) {
		c.HTML(http.StatusOK, "handler.html.tmpl", options)
	})
	// Run the server
	router.Run(fmt.Sprintf(":%d", options.Port))
}
