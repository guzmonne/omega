package configure

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

// Environment variables type
type Environment struct {
	Values []string
}
// Unmarshalls the Environment type from a `map[string]string` into `[]string`.
func (env *Environment) UnmarshalYAML(value *yaml.Node) error {
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
  // Delay between frames in ms.
  // If the value is "auto" use the actual recording delay.
  //
  FrameDelay Auto `yaml:"frameDelay"`
  //
  // Maximum delay between frames in ms.
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
	// Create a default config struct
	config := &Config{
		Command: "/bin/bash",
		Cwd: os.Getenv("HOME") + "/" + ".omega",
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

// ReadConfig reads a config file and return its unmarshalled contents
func ReadConfig(configPath string) (*Config, error) {
	// Check if the config exists at `configPath`
	if _, err := os.Stat(configPath); err != nil {
		return &Config{}, errors.New("Can't find a file at: " + configPath)
	}
	// Open the configuration file
	configFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return &Config{}, err
	}
	// Unmarshall the configuration file
	var config Config
	if err := yaml.Unmarshal(configFile, &config); err != nil {
		return &Config{}, err
	}

	return &config, nil
}

// WriteConfig writes a config object to a YAML file
func WriteConfig(configPath string, config Config) error {
	// Create YAML encoder
	var file bytes.Buffer
	encoder := yaml.NewEncoder(&file)
	encoder.SetIndent(2)
	// Encode the default config
	if err := encoder.Encode(config); err != nil {
		return err
	}
	// Write the recording file
	if err := ioutil.WriteFile(configPath, file.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

// Merge mergess two Config structs
func Merge(config1 *Config, config2 *Config) *Config {
	// Merge `cwd`
	config1.Cwd = config2.Cwd

	return config1
}

// UpdateConfig updates the cli project configuration.
func UpdateConfig(configPath string, updates *Config) error {
	// Get the configuration from the file
	config, err := ReadConfig(configPath)
	if err != nil {
		return err
	}
	// Mege both configuration structs
	config = Merge(config, updates)
	if err != nil {
		return err
	}
	// Store the updated configuration to the file
	if err := WriteConfig(configPath, *config); err != nil {
		return err
	}

	return nil
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
const CONFIG_FILENAME = "config.yml"

func writeDefaultConfig(configPath string) error {
	if _, err := os.Stat(configPath); err != nil {
		// File doesn't exists
		config, err := DefaultConfig()
		if err != nil {
			return err
		}
		WriteConfig(configPath, *config)
		color.Green("Configuration file created at: %s", configPath)
	} else {
		// File exists
		color.Blue("Configuration file already exists: %s", configPath)
	}

	return nil
}
// Init instantiates the cli project folder with its default configuration.
func Init(projectFolder string) error {
	stat, err := os.Stat(projectFolder)
	if err != nil {
		// Folder doesn't exists
		if err := os.MkdirAll(projectFolder, 0755); err != nil {
			return err
		}
		color.Green("Folder %s created.", projectFolder)
	} else {
		if stat.IsDir() {
			color.Blue("Folder %s already exists.", projectFolder)
		}
	}

	writeDefaultConfig(projectFolder + "/config.yml")

	return nil
}
