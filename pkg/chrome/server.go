package chrome

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// WebServerOptions are used to configure the web server.
type WebServerOptions struct {
	// Port from which to run the server
	Port int
}

// NewWebServerOptions creates a default WebServerOptions struct.
func NewWebServerOptions() WebServerOptions {
	return WebServerOptions{
		Port: 38080,
	}
}

// start the server from which the handler function is served.
func Serve(options WebServerOptions) {
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
		c.HTML(http.StatusOK, "three.html.tmpl", nil)
	})
	router.GET("/dev", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dev.html.tmpl", nil)
	})
	router.GET("/dev/script.js", func(c *gin.Context) {
		c.String(http.StatusOK, D.GetScript())
	})
	// Run the server
	router.Run(fmt.Sprintf(":%d", options.Port))
}

