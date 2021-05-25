package chrome

import (
	"fmt"
	"net/http"
	"os"
	"regexp"

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

var cssRe = regexp.MustCompile(`\.css$`)

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
	router.GET("/dev/:asset", func(c *gin.Context) {
		asset   := c.Param("asset")
		content := string(build.Get(asset))
		if cssRe.MatchString(asset) {
			c.Header("Content-Type", "text/css")
		}
		c.String(http.StatusOK, content)
	})
	// Run the server
	router.Run(fmt.Sprintf(":%d", options.Port))
}

