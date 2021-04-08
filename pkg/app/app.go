package app

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"gux.codes/omega/pkg/chrome"
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
		Name: "Ωmega",
		Usage: "CLI Recorder",
		Action: func(c *cli.Context) error {
			fmt.Println("Command not found. Try the -h or --help flags for more information.")
			return nil
		},
		Commands: []*cli.Command{
			// Play
			{
				Name: "play",
				Usage: "reproduce a recording file",
				Aliases: []string{"p"},
				UsageText: "omega play [command options] RECORDING",
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
					&cli.Float64Flag{
						Name: "speedFactor",
						Aliases: []string{"f"},
						Value: 1.0,
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
						SpeedFactor: c.Float64("speedFactor"),
					}
					player.Play(recordingPath, *options)

					return nil
				},
			},
			// Init
			{
				Name: "init",
				Aliases: []string{"i"},
				Usage: "initialize the app",
				UsageText: "omega init [command options]",
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
			// Screenshot
			{
				Name: "screenshot",
				Aliases: []string{"s"},
				Usage: "take a screenshot of your recording",
				UsageText: "omega screenshot [command options]",
				Action: func(c *cli.Context) error {
					if err := chrome.ScreenShot(); err != nil {
						return err
					}

					return nil
				},
			},
			// Record
			{
				Name: "record",
				Aliases: []string{"r"},
				Usage: "record a shell session or chrome animation",
				UsageText: "omega record SUBCOMMAND [subcommand options]",
				Subcommands: []*cli.Command{
					// Shell
					{
						Name: "shell",
						Aliases: []string{"s"},
						Usage: "records a shell session",
						UsageText: "omega record shell [command options]",
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
							config, err := configure.ReadConfig(configPath)
							if err != nil {
								log.Fatal(err)
							}

							if err := record.Shell(outputPath, config); err != nil {
								log.Fatal(err)
							}

							return nil
						},
					},
					// Chrome
					{
						Name: "chrome",
						Aliases: []string{"c"},
						Usage: "record a chrome animation",
						UsageText: "omega record chrome [command options]",
						Action: func(c *cli.Context) error {
							if err := record.Chrome(); err != nil {
								return err
							}
							return nil
						},
					},
				},
			},
		},
	}

	return *app
}