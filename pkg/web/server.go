package web

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gux.codes/omega/pkg/utils"
)

// WebServerOptions are used to configure the web server.
type WebServerOptions struct {
	// Width of the page
	Width int
	// Height of the page
	Height int
	// Port from which to run the server
	Port int
}

// NewWebServerOptions creates a default WebServerOptions struct.
func NewWebServerOptions() *WebServerOptions {
	return &WebServerOptions{
		Width: 1920,
		Height: 1080,
		Port: 8080,
	}
}

// start the server from which the handler function is served.
func startWebServer(options WebServerOptions) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Web server error in: ", r)
		}
	}()
	// Force logs in color
	gin.ForceConsoleColor()
	// Set the "release" mode
	gin.SetMode(gin.ReleaseMode)
	// Create the default router
	router := gin.New()
	// Change the way logs are outputed.
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s %s %d\n", utils.BoxBlue(param.Method), param.Path, param.StatusCode)
	}))
	router.Use(gin.Recovery())
	// Load the templates
	templates := os.Getenv("OMEGA_SERVER_TEMPLATES")
	if templates == "" {
		templates = "./templates/*"
	}
	router.LoadHTMLGlob(templates)
	// Serve static assets
	assets := os.Getenv("OMEGA_SERVER_ASSETS")
	if assets == "" {
		assets = "./assets"
	}
	router.Static("/assets", assets)
	// Create the routes
	router.GET("/handler", func(c *gin.Context) {
		c.HTML(http.StatusOK, "handler.html.tmpl", gin.H{"width": options.Width, "height": options.Height})
	})
	// Run the server
	router.Run(fmt.Sprintf(":%d", options.Port))
}

