package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"gux.codes/omega/pkg/chrome"
	"gux.codes/omega/pkg/shell"
	"gux.codes/omega/pkg/utils"
)

func main() {
	var outputPath string

	var ulid = utils.ULID()

	app := &cli.App{
		Name: "Ωmega",
		Usage: "CLI and Chrome Recorder",
		UsageText: "omega COMMAND [COMMAND OPTIONS] SUBCOMMAND [SUBCOMMAND OPTIONS] ARGUMENTS",
		Action: func(c *cli.Context) error {
			fmt.Println("Command not found. Try the -h or --help flags for more information.")
			return nil
		},
		Commands: []*cli.Command{
			// Shell
			{
				Name: "shell",
				Usage: "record or play a shell session recording",
				Aliases: []string{"r"},
				UsageText: "omega shell SUBCOMMAND [SUBCOMMAND OPTIONS]",
				Subcommands: []*cli.Command{
					// Play
					{
						Name: "play",
						Usage: "play a recording file",
						Aliases: []string{"p"},
						UsageText: "omega shell play [OPTIONS] RECORDING",
						Flags: []cli.Flag{
							&cli.IntFlag{
								Name: "maxIdleTime",
								Value: -1,
								Usage: "sets the maximum delay between frames in ms",
								EnvVars: []string{"OMEGA_SHELL_PLAY_MAXIDLETIME"},
							},
							&cli.IntFlag{
								Name: "frameDelay",
								Value: -1,
								Usage: "sets a fixed delay between records in ms.",
								EnvVars: []string{"OMEGA_SHELL_PLAY_FRAMEDELAY"},
							},
							&cli.BoolFlag{
								Name: "silent",
								Value: false,
								Usage: "silence the message before starting the recording",
								EnvVars: []string{"OMEGA_SHELL_PLAY_SILENT"},
							},
							&cli.Float64Flag{
								Name: "speedFactor",
								Value: 1.0,
								Usage: "applies a multiplier to each delay",
								EnvVars: []string{"OMEGA_SHELL_PLAY_SPEEDFACTOR"},
							},
						},
						Action: func(c *cli.Context) error {
							// Check if a recording file was supplied
							if c.NArg() == 0 {
								return errors.New("no recording file was supplied")
							}
							recordingPath := c.Args().Get(0)

							// Create the PlayOptions object
							options := &shell.PlayOptions{
								MaxIdleTime: c.Int("maxIdleTime"),
								FrameDelay: c.Int("frameDelay"),
								Silent: c.Bool("silent"),
								SpeedFactor: c.Float64("speedFactor"),
							}

							// Play the animation
							shell.Play(recordingPath, *options)

							return nil
						},
					},
					// Record
					{
						Name: "record",
						Aliases: []string{"r"},
						Usage: "records a shell session",
						UsageText: "omega shell record [OPTIONS]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name: "command",
								Usage: "command to run on the pty",
								EnvVars: []string{"OMEGA_SHELL_RECORD_COMMAND"},
							},
							&cli.StringFlag{
								Name: "cwd",
								Usage: "current working directory",
								EnvVars: []string{"OMEGA_SHELL_RECORD_CWD"},
							},
							&cli.StringSliceFlag{
								Name: "env",
								Usage: "map of environment variables",
								EnvVars: []string{"OMEGA_SHELL_RECORD_ENV"},
							},
							&cli.IntFlag{
								Name: "minDelay",
								Usage: "minimum delay in ms between two records",
								EnvVars: []string{"OMEGA_SHELL_RECORD_MINDELAY"},
							},
							&cli.IntFlag{
								Name: "cols",
								Usage: "number of columns to display on the pty interface",
								EnvVars: []string{"OMEGA_SHELL_RECORD_COLS"},
							},
							&cli.IntFlag{
								Name: "rows",
								Usage: "number of rows to display on the pty interface",
								EnvVars: []string{"OMEGA_SHELL_RECORD_ROWS"},
							},
							&cli.StringFlag{
								Name: "outputPath",
								Aliases: []string{"o"},
								Value: "./" + ulid + ".yml",
								DefaultText: "./{{ ulid }}.yml",
								Usage: "recording output path",
								Destination: &outputPath,
								EnvVars: []string{"OMEGA_SHELL_RECORD_OUTPUTPATH"},
							},
						},
						Action: func(c *cli.Context) error {
							specification := shell.NewShellSpecification()

							// Overwrite default specification options
							if command := c.String("command"); command != "" {
								specification.Command = command
							}
							if cwd := c.String("cwd"); cwd != "" {
								specification.Cwd = cwd
							}
							if env := c.StringSlice("env"); env != nil {
								specification.Env = env
							}
							if minDelay := c.Int("minDelay"); minDelay != 0 {
								specification.MinDelay = c.Int("minDelay")
							}
							if cols := c.Int("cols"); cols != 0 {
								specification.Cols = cols
							}
							if rows := c.Int("rows"); rows != 0 {
								specification.Rows = rows
							}
							if outputPath := c.String("outputPath"); outputPath != "" {
								specification.OutputPath = outputPath
							}

							// Start recording the shell
							if err := shell.Shell(*specification); err != nil {
								log.Fatal(err)
							}

							return nil
						},
					},
				},
			},
			// Chrome
			{
				Name: "chrome",
				Aliases: []string{"c"},
				Usage: "chrome animation",
				UsageText: "omega chrome SUBCOMMAND [SUBCOMMAND OPTIONS]",
				Subcommands: []*cli.Command{
					{
						Name: "record",
						Usage: "record chrome animation",
						UsageText: "omega chrome record [OPTIONS]",
						Flags: []cli.Flag{},
						Action: func(c *cli.Context) error {
							if err := chrome.Record(); err != nil {
								return err
							}

							return nil
						},
					},
					{
						Name: "serve",
						Usage: "serve the handler web server",
						UsageText: "omega chrome serve [OPTIONS]",
						Flags: []cli.Flag{},
						Action: func(c *cli.Context) error {
							webServerOptions := chrome.NewWebServerOptions()
							chrome.Serve(webServerOptions)

							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}