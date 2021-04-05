package app

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"gux.codes/omega/pkg/configure"
	"gux.codes/omega/pkg/record"
	"gux.codes/omega/pkg/utils"
)

func CreateApp() cli.App {
	var configPath string
	var outputPath string
	var projectFolder string
	var ulid = utils.ULID()

	app := &cli.App{
		Name: "Omega",
		Usage: "CLI Recorder",
		Action: func(c *cli.Context) error {
			fmt.Println("Command not found. Try the -h or --help flags for more information.")
			return nil
		},
		Commands: []*cli.Command{
			{
				Name: "init",
				Aliases: []string{"i"},
				Usage: "initialize the app",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "folder",
						Aliases: []string{"f"},
						Value: os.Getenv("HOME") + "/.omega",
						Usage: "project folder",
						Destination: &projectFolder,
						EnvVars: []string{"OMEGA_PROJECT_FOLDER"},
					},
					&cli.StringFlag{
						Name: "cwd",
						Aliases: []string{"f"},
						Value: os.Getenv("HOME") + "/.omega",
						Usage: "change the default config 'cwd' value",
						EnvVars: []string{"OMEGA_PROJECT_CWD"},
					},
				},
				Action: func(c *cli.Context) error {
					if err := configure.Init(projectFolder); err != nil {
						log.Fatal(err)
					}

					return nil
				},
			},
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
					config, err := configure.NewConfig(configPath)
					if err != nil {
						log.Fatal(err)
					}

					if err := record.RecordShell(outputPath, config); err != nil {
						log.Fatal(err)
					}

					return nil
				},
			},
		},
	}

	return *app
}