package chrome

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/gobs/args"
)

func runCommand(commandString string) error {
	parts := args.GetArgs(commandString)
	cmd := exec.Command(parts[0], parts[1:]...)
	return cmd.Start()
}

// Starts a headless Chrome process
func Start() error {
	chromeapp := os.Getenv("OMEGA_CHROMEAPP")

	if chromeapp == "" {
		switch runtime.GOOS {
		case "darwin":
			for _, c := range []string{
				"/Applications/Google Chrome Canary.app",
				"/Applications/Google Chrome.app",
			} {
				// MacOS apps are actually folders
				if info, err := os.Stat(c); err == nil && info.IsDir() {
					chromeapp = fmt.Sprintf("open %q --args", c)
					break
				}
			}

		case "linux":
			for _, c := range []string{
				"headless_shell",
				"chromium",
				"google-chrome-beta",
				"google-chrome-unstable",
				"google-chrome-stable"} {
				if _, err := exec.LookPath(c); err == nil {
					chromeapp = c
					break
				}
			}

		case "windows":
			// TODO
		}
	}

	if chromeapp == "" {
		return errors.New("chromeapp not found")
	}

	if chromeapp == "headless_shell" {
		chromeapp += " --no-sandbox"
	} else {
		chromeapp += " --headless"
	}

	chromeapp += " --remote-debugging-port=9222 --hide-scrollbars --bwsi --disable-extensions about:blank"

	if err := runCommand(chromeapp); err != nil {
		return err
	}

	return nil
}