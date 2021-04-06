package player

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// Map that stores clear functions
var clear map[string]func()

func init() {
	// Initialize the clear map
	clear = make(map[string]func())
	// Clear command for Linux OS
	clear["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	// Clear command for windos
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	// Clear command for macOS (darwin)
	clear["darwin"] = func() {
		fmt.Printf("\033[2J")
	}
}

// Clear cleans the terminal window. Should work in Linux, macOs, and Windos.
func Clear() {
	// Get the appropiate function for the current OS
	f, ok := clear[runtime.GOOS]
	if ok {
		// The function is valid so we can call it.
		f()
		fmt.Printf("\033[0;0H")
	} else {
		fmt.Println("Cleaning the screen is not supported for your OS:", runtime.GOOS)
	}
}
