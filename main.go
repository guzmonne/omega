package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	var configPath string
	var outputPath string
	var ulid = ULID()

	app := &cli.App{
		Name: "Omega",
		Usage: "CLI Recorder",
		Action: func(c *cli.Context) error {
			fmt.Println("Command not found. Try the -h or --help flags for more information.")
			return nil
		},
		Commands: []*cli.Command{
			{
				Name: "record",
				Aliases: []string{"r"},
				Usage: "record a cli session",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "config",
						Aliases: []string{"c"},
						Value: "./config.yml",
						Usage: "configuration file",
						Destination: &configPath,
						EnvVars: []string{"OMEGA_CONFIG"},
						Required: true,
					},
					&cli.StringFlag{
						Name: "output",
						Aliases: []string{"o"},
						Value: "./" + ulid + ".yml",
						DefaultText: "./{random}.yml",
						Usage: "configuration file",
						Destination: &outputPath,
						EnvVars: []string{"OMEGA_OUTPUT"},
					},
				},
				Action: func(c *cli.Context) error {
					// Create the configuration struct
					config, err := NewConfig(configPath)
					if err != nil {
						log.Fatal(err)
					}

					if err := RecordShell(outputPath, config); err != nil {
						log.Fatal(err)
					}

					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}