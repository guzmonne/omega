package record

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gux.codes/omega/pkg/utils"
)

// startServers starts the server from which the handler function is served.
func startServer(method string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("error in: ", r)
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
	router.LoadHTMLGlob("./templates/*")
	// Serve static assets
	router.Static("/assets", "./assets")
	// Create the routes
	router.GET("/handler", func(c *gin.Context) {
		c.HTML(http.StatusOK, "handler.html.tmpl", gin.H{"width": 1920, "height": 1080, "method": method})
	})
	// Run the server
	router.Run(fmt.Sprintf(":%d", port))
}

