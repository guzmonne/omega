package app

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"gux.codes/omega/pkg/configure"
	"gux.codes/omega/pkg/player"
	"gux.codes/omega/pkg/record"
	"gux.codes/omega/pkg/utils"
)

func CreateApp() cli.App {
	var configPath string
	var outputPath string
	var projectFolder string

	var ulid = utils.ULID()
	var home = os.Getenv("HOME") + "/.omega"

	app := &cli.App{
		Name: "Omega",
		Usage: "CLI Recorder",
		Action: func(c *cli.Context) error {
			fmt.Println("Command not found. Try the -h or --help flags for more information.")
			return nil
		},
		Commands: []*cli.Command{
			{
				Name: "play",
				Usage: "reproduce a recording file",
				Aliases: []string{"p"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name: "realTiming",
						Aliases: []string{"r"},
						Value: false,
						Usage: "use the original timing between records",
						EnvVars: []string{"OMEGA_PLAY_REAL_TIMING"},
					},
					&cli.BoolFlag{
						Name: "silent",
						Aliases: []string{"s"},
						Value: false,
						Usage: "silence the message before starting the recording",
						EnvVars: []string{"OMEGA_PLAY_SILENT"},
					},
					&cli.IntFlag{
						Name: "speedFactor",
						Aliases: []string{"f"},
						Value: 1,
						Usage: "applies a multiplier to each delay",
						EnvVars: []string{"OMEGA_PLAY_SILENT"},
					},
				},
				Action: func(c *cli.Context) error {
					var recordingPath string
					if c.NArg() == 0 {
						return errors.New("No recording file was supplied")
					}
					recordingPath = c.Args().Get(0)
					options := &player.PlayOptions{
						RealTiming: c.Bool("realTiming"),
						Silent: c.Bool("silent"),
						SpeedFactor: c.Int("speedFactor"),
					}
					player.Play(recordingPath, *options)

					return nil
				},
			},
			{
				Name: "init",
				Aliases: []string{"i"},
				Usage: "initialize the app",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "path",
						Aliases: []string{"p"},
						Value: home,
						Usage: "project folder path",
						Destination: &projectFolder,
						EnvVars: []string{"OMEGA_PROJECT_FOLDER"},
					},
					&cli.BoolFlag{
						Name: "force",
						Aliases: []string{"f"},
						Value: false,
						Usage: "overwrites the project folder if defined",
						EnvVars: []string{"OMEGA_INIT_FORCE"},
					},
				},
				Action: func(c *cli.Context) error {
					if c.Bool("force") {
						if err := os.RemoveAll(projectFolder); err != nil {
							log.Fatal(err)
						} else {
							color.Red("Folder %s destroyed", projectFolder)
							fmt.Println("---")
						}
					}

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
						Value: home + "/config.yml",
						Usage: "configuration file",
						Destination: &configPath,
						EnvVars: []string{"OMEGA_RECORD_CONFIG"},
					},
					&cli.StringFlag{
						Name: "output",
						Aliases: []string{"o"},
						Value: "./" + ulid + ".yml",
						DefaultText: "./{random}.yml",
						Usage: "configuration file",
						Destination: &outputPath,
						EnvVars: []string{"OMEGA_RECORD_OUTPUT"},
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