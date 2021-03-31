package main

import (
	"flag"
	"fmt"
	"log"
)

// ParseFlags will parse de CLI flags and apply defaults to its values.
func ParseFlags() (string, error) {
	var configPath string

	// Setup a CLI flag called `-config-path`.
	flag.StringVar(&configPath, "config-path", "./config.yml", "Path to the CLI configuration file.")

	// Parse the flags
	flag.Parse()

	// Validate the path
	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}

	return configPath, nil
}

func main() {
	// Generate the config struct
	configPath, err := ParseFlags()
	if err != nil {
		log.Fatal(err)
	}
	config, err := NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("--- config:\n%v\n\n", config)

	if err := RecordShell(config); err != nil {
		log.Fatal(err)
	}
}