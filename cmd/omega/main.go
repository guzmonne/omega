package main

import (
	"log"
	"os"

	"gux.codes/omega/pkg/app"
)

func main() {
	app := app.CreateApp()

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}