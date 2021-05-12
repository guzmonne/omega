package utils

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/oklog/ulid"
)

func Touch(filePath string, content string) error {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		// If the file does not exist we create it
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		// Write the provided content
		if _, err := file.Write([]byte(content)); err != nil {
			return err
		}
		defer file.Close()
	} else {
		// We touch the file by updating its access and modification times
		currentTime := time.Now().Local()
		if err := os.Chtimes(filePath, currentTime, currentTime); err != nil {
			return err
		}
	}

	return nil
}
// ULID returns a valid pseudo-random ULID id.
func ULID() string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}

var bgGreen = color.New(color.BgGreen).SprintFunc()
var bgRed = color.New(color.BgRed).SprintFunc()
var bgBlue = color.New(color.BgBlue).SprintFunc()

// BoxGreen returns a string inside a green box
func BoxGreen(content string) string {
	return bgGreen(" " + content + " ")
}

// BoxRed returns a string inside a red box
func BoxRed(content string) string {
	return bgRed(" " + content + " ")
}
// BoxBlue returns a string inside a blue box
func BoxBlue(content string) string {
	return bgBlue(" " + content + " ")
}

// Float64 allocates and returns an *float64
func Float64(x float64) *float64 {
	return &x
}

func Info(s string) {
	fmt.Printf("%s %s\n", BoxBlue("INFO   "), s)
}

func Message(s string) {
	fmt.Printf("%s %s\n", BoxGreen("MESSAGE"), s)
}

func Command(s string) {
	fmt.Printf("%s %s\n", BoxRed("COMMAND"), s)
}

func Error(s string) {
	fmt.Printf("%s %s\n", BoxRed("ERROR  "), s)
}

func Success(s string) {
	fmt.Printf("%s %s\n", BoxGreen("SUCCESS"), s)
}