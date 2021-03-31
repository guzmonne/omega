package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Auto type indicates a value if its bigger or equal to zero
// and -1 when the value is not an int.
type Auto int
// Environment variables type
type Environment struct {
	Values []string
}
// Unmarshall function that will parse any value that is not an
// integer as the value "auto".
func (auto *Auto) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value int

	// Try to unmarshall the value as if it was an int
	err := unmarshal(&value)

	// If an error ocurred then set the value as -1 else store the value.
	if err != nil {
		*auto = -1
	} else {
		*auto = Auto(value)
	}

	return nil
}
// Unmarshalls the Environment type from a `map[string]string` into `[]string`.
func (env *Environment) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Variable to store the environment variables with format key=value
	var values []string
	// Environment unmarshaled map
	environments := make(map[string]string)

	// Trt to unmarshall the value as if it was a []string
	err := unmarshal(&values)
	if err == nil {
		env.Values = values
		return nil
	}

	// Try to unmarshall the value.
	err = unmarshal(&environments)
	if err != nil {
		// If its empty return an empty map.
		env.Values = make([]string, 0)
	} else {
		// If keys are defined convert them into a slice of `key=value` strings
		for key, value := range environments {
			values = append(values, key + "=" + value)
		}
		env.Values = values
	}

	return nil
}
//
// CLI Configuration object
//
type Config struct {
  //
  // Command to execute on the pty interface.
  //
  Command string `yaml:"command"`
  //
  // Current working directory of the pty interface.
  //
  Cwd string `yaml:"cwd"`
  //
  // Environment variables to populate the pty interface.
  // They will override the default environment variables.
  //
  Env Environment `yaml:"env"`
  //
  // The number of columns to display on the pty interface.
  //
  Cols Auto `yaml:"cols"`
  //
  // The number of rows to display on the pty interface.
  //
  Rows Auto `yaml:"rows"`
  //
  // Ammount of times the animation should repeat when
  // rendering the recording as a GIF:
  // - `-1`: Play once.
  // - `0`: Loop indefinitely.
  // - `n`: Loop `n` times.
  //
  Repeat int `yaml:"repeat"`
  //
  // Quality of the GIF.
  //
  Quality int `yaml:"quality"`
  //
  // Delay between frames in ns.
  // If the value is "auto" use the actual recording delay.
  //
  FrameDelay Auto `yaml:"frameDelay"`
  //
  // Maximum delay between frames in ns.
  // Ignored if the `frameDelay` option is set to `auto`.
  // Set to `auto` to prevent limiting the max idle time.
  //
  MaxIdleTime Auto `yaml:"maxIdleTimeout"`
  //
  // Style of the cursor on the animation.
  //
  CursorStyle string `yaml:"cursorStyle"`
  //
  // Font to be used on the animation. It can be any
  // font installed on your machine.
  //
  FontFamily string `yaml:"fontFamily"`
  //
  // The size of the font.
  //
  FontSize int `yaml:"fontSize"`
  //
  // The height of the lines.
  //
  LineHeight int `yaml:"lineHeight"`
  //
  // The spacing between letters
  //
  LetterSpacing int `yaml:"letterSpacing"`
}

// DefaultConfig returns a default Config struct
func DefaultConfig() (*Config, error) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Create a default config struct
	config := &Config{
		Command: "/bin/bash",
		Cwd: cwd,
		Env: Environment{make([]string, 0)},
		Cols: Auto(-1),
		Rows: Auto(-1),
		Repeat: 0,
		Quality: 100,
		FrameDelay: Auto(-1),
		MaxIdleTime: Auto(-1),
		CursorStyle: "block",
		FontFamily: "Monaco, Lucida Console, Ubuntu Mono, Monospace",
		FontSize: 12,
		LineHeight: 1,
		LetterSpacing: 0,
	}
	return config, nil
}

// NewConfig returns a new CLI configuration with its default values.
func NewConfig(configPath string) (*Config, error) {
	config, err := DefaultConfig()

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decoder
	decoder := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	// Return the configuration object
	return config, nil
}

// ValidateConfigPath makes sure the path is valid.
func ValidateConfigPath(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a file.", path)
	}
	return nil
}
