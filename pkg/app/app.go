package app

import (
	"errors"
	"fmt"
	"log"

	"github.com/urfave/cli/v2"
	"gux.codes/omega/pkg/player"
	"gux.codes/omega/pkg/record"
	"gux.codes/omega/pkg/utils"
)

func CreateApp() cli.App {
	var outputPath string

	var ulid = utils.ULID()

	app := &cli.App{
		Name: "Ωmega",
		Usage: "CLI and Chrome Recorder",
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
					&cli.IntFlag{
						Name: "maxIdleTime",
						Value: -1,
						Usage: "sets the maximum delay between frames in ms",
						EnvVars: []string{"OMEGA_PLAY_MAXIDLETIME"},
					},
					&cli.IntFlag{
						Name: "frameDelay",
						Value: -1,
						Usage: "sets a fixed delay between records in ms.",
						EnvVars: []string{"OMEGA_PLAY_FRAMEDELAY"},
					},
					&cli.BoolFlag{
						Name: "silent",
						Value: false,
						Usage: "silence the message before starting the recording",
						EnvVars: []string{"OMEGA_PLAY_SILENT"},
					},
					&cli.Float64Flag{
						Name: "speedFactor",
						Value: 1.0,
						Usage: "applies a multiplier to each delay",
						EnvVars: []string{"OMEGA_PLAY_SPEEDFACTOR"},
					},
				},
				Action: func(c *cli.Context) error {
					// Check if a recording file was supplied
					if c.NArg() == 0 {
						return errors.New("no recording file was supplied")
					}
					recordingPath := c.Args().Get(0)

					// Create the PlayOptions object
					options := &player.PlayOptions{
						MaxIdleTime: c.Int("maxIdleTime"),
						FrameDelay: c.Int("frameDelay"),
						Silent: c.Bool("silent"),
						SpeedFactor: c.Float64("speedFactor"),
					}

					// Play the animation
					player.Play(recordingPath, *options)

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
								Name: "command",
								Usage: "command to run on the pty",
								EnvVars: []string{"OMEGA_RECORD_SHELL_COMMAND"},
							},
							&cli.StringFlag{
								Name: "cwd",
								Usage: "current working directory",
								EnvVars: []string{"OMEGA_RECORD_SHELL_CWD"},
							},
							&cli.StringSliceFlag{
								Name: "env",
								Usage: "map of environment variables",
								EnvVars: []string{"OMEGA_RECORD_SHELL_ENV"},
							},
							&cli.IntFlag{
								Name: "minDelay",
								Usage: "minimum delay in ms between two records",
								EnvVars: []string{"OMEGA_RECORD_SHELL_MINDELAY"},
							},
							&cli.IntFlag{
								Name: "cols",
								Usage: "number of columns to display on the pty interface",
								EnvVars: []string{"OMEGA_RECORD_SHELL_COLS"},
							},
							&cli.IntFlag{
								Name: "rows",
								Usage: "number of rows to display on the pty interface",
								EnvVars: []string{"OMEGA_RECORD_SHELL_ROWS"},
							},
							&cli.StringFlag{
								Name: "outputPath",
								Aliases: []string{"o"},
								Value: "./" + ulid + ".yml",
								DefaultText: "./{{ ulid }}.yml",
								Usage: "recording output path",
								Destination: &outputPath,
								EnvVars: []string{"OMEGA_RECORD_SHELL_OUTPUTPATH"},
							},
						},
						Action: func(c *cli.Context) error {
							specification := record.NewShellSpecification()

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
							if err := record.Shell(*specification); err != nil {
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
						Flags: []cli.Flag{
							// Record flags
							&cli.StringFlag{
								Name: "method",
								Aliases: []string{"m"},
								Value: "timeweb",
								Usage: "recording method (must be one of 'timeweb' or 'screencast')",
								EnvVars: []string{"OMEGA_RECORD_WIDTH"},
							},
							&cli.IntFlag{
								Name: "width",
								Value: 1920,
								Usage: "width of the Chrome window",
								EnvVars: []string{"OMEGA_RECORD_METHOD"},
							},
							&cli.IntFlag{
								Name: "height",
								Value: 1080,
								Usage: "height of the Chrome window",
								EnvVars: []string{"OMEGA_RECORD_HEIGHT"},
							},
							&cli.IntFlag{
								Name: "virtualTime",
								Aliases: []string{"v"},
								Value: 0,
								Usage: "initial virtual time (only used when method is `timeweb`)",
								EnvVars: []string{"OMEGA_RECORD_VIRTUAL_TIME"},
							},
							&cli.Float64Flag{
								Name: "fps",
								Aliases: []string{"f"},
								Value: 0,
								Usage: "onfigures the rate at which the animation will be recorded (only used when method is `timeweb`)",
								EnvVars: []string{"OMEGA_RECORD_FPS"},
							},
							// Chrome devtools flags
							&cli.IntFlag{
								Name: "devtools.port",
								Value: 9222,
								Usage: "port to be used by the Chrome Devtools Protocol",
								EnvVars: []string{"OMEGA_RECORD_DEVTOOLS_PORT"},
							},
							&cli.BoolFlag{
								Name: "devtools.allow_http_screen_capture",
								Value: true,
								Usage: "allows non-secure origins to use the screen capture API",
								EnvVars: []string{"OMEGA_RECORD_DEVTOOLS_ALLOWHTTPSCREENCAPTURE"},
							},
							&cli.BoolFlag{
								Name: "devtools.allow_insecure_localhost",
								Value: true,
								Usage: "enables TLS/SSL errors on localhost to be ignored",
								EnvVars: []string{"OMEGA_RECORD_DEVTOOLS_ALLOWINSECURELOCALHOST"},
							},
							&cli.BoolFlag{
								Name: "devtools.bwsi",
								Value: true,
								Usage: "indicates that the browser will run a Guest session",
								EnvVars: []string{"OMEGA_RECORD_DEVTOOLS_BWSI"},
							},
							&cli.BoolFlag{
								Name: "devtools.disable_extensions",
								Value: true,
								Usage: "disable the use of browser extensions",
								EnvVars: []string{"OMEGA_RECORD_DEVTOOLS_DISABLEEXTENSIONS"},
							},
							&cli.BoolFlag{
								Name: "devtools.disable_frame_rate_limit",
								Value: true,
								Usage: "disable the use of browser extensions",
								EnvVars: []string{"OMEGA_RECORD_DEVTOOLS_DISABLEFRAMERATELIMIT"},
							},
							&cli.BoolFlag{
								Name: "devtools.disable_gpu",
								Value: true,
								Usage: "disables GPU hardware acceleration",
								EnvVars: []string{"OMEGA_RECORD_DEVTOOLS_DISABLEFRAMERATELIMIT"},
							},
							&cli.BoolFlag{
								Name: "devtools.disable_web_security",
								Value: true,
								Usage: "makes the browser don't enforce the same-origin policy",
								EnvVars: []string{"OMEGA_RECORD_DEVTOOLS_DISABLEWEBSECURITY"},
							},
							&cli.BoolFlag{
								Name: "devtools.enable_accelerated_2d_canvas",
								Value: true,
								Usage: "enables accelerated 2D canvas",
								EnvVars: []string{"OMEGA_RECORD_DEVTOOLS_ENABLEACCELERATED2DCANVAS"},
							},
							&cli.BoolFlag{
								Name: "devtools.hide_scrollbars",
								Value: true,
								Usage: "disables the browser scrollbars",
								EnvVars: []string{"OMEGA_RECORD_DEVTOOLS_HIDESCROLLBARS"},
							},
						},
						Action: func(c *cli.Context) error {
							specification := record.NewChromeRecordingSpecification()
							// Override default chrome record specification options
							specification.Method = c.String("method")
							specification.ChromeFlags.CastInitialScreenWidth = c.Int("width")
							specification.ChromeFlags.CastInitialScreenHeight = c.Int("height")
							specification.VirtualTime = c.Int("virtualTime")
							specification.FPS = c.Float64("fps")
							// Override default devtools options
							if specification.ChromeFlags.DevToolsPort != c.Int("devtools.port") {
								specification.ChromeFlags.DevToolsPort = c.Int("devtools.port")
							}
							if specification.ChromeFlags.AllowHttpScreenCapture != c.Bool("devtools.allow_http_screen_capture") {
								specification.ChromeFlags.AllowHttpScreenCapture = c.Bool("devtools.allow_http_screen_capture")
							}
							if specification.ChromeFlags.AllowInsecuredLocalhost != c.Bool("devtools.allow_insecure_localhost") {
								specification.ChromeFlags.AllowInsecuredLocalhost = c.Bool("devtools.allow_insecure_localhost")
							}
							if specification.ChromeFlags.BWSI != c.Bool("devtools.bwsi") {
								specification.ChromeFlags.BWSI = c.Bool("devtools.bwsi")
							}
							if specification.ChromeFlags.DisableExtensions != c.Bool("devtools.disable_extensions") {
								specification.ChromeFlags.DisableExtensions = c.Bool("devtools.disable_extensions")
							}
							if specification.ChromeFlags.DisableFrameRateLimit != c.Bool("devtools.disable_frame_rate_limit") {
								specification.ChromeFlags.DisableFrameRateLimit = c.Bool("devtools.disable_frame_rate_limit")
							}
							if specification.ChromeFlags.DisableGPU != c.Bool("devtools.disable_gpu") {
								specification.ChromeFlags.DisableGPU = c.Bool("devtools.disable_gpu")
							}
							if specification.ChromeFlags.DisableWebSecurity != c.Bool("devtools.disable_web_security") {
								specification.ChromeFlags.DisableWebSecurity = c.Bool("devtools.disable_web_security")
							}
							if specification.ChromeFlags.EnableAccelerated2dCanvas != c.Bool("devtools.enable_accelerated_2d_canvas") {
								specification.ChromeFlags.EnableAccelerated2dCanvas = c.Bool("devtools.enable_accelerated_2d_canvas")
							}
							if specification.ChromeFlags.HideScrollbars != c.Bool("devtools.hide_scrollbars") {
								specification.ChromeFlags.HideScrollbars = c.Bool("devtools.hide_scrollbars")
							}
							// Open Chrome to record the animation
							if err := record.Chrome(specification); err != nil {
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