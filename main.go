package main

import (
	"flag"
	"log"
)

// ParseFlags will parse de CLI flags and apply defaults to its values.
func ParseFlags() (string, string, error) {
	var configPath string
	var outputPath string
	var ulid = ULID()

	// Setup flags
	flag.StringVar(&configPath, "config", "./config.yml", "Path to the CLI configuration file.")
	flag.StringVar(&outputPath, "output", "./" + ulid + ".yml", "Path to store the recordin output.")

	// Parse the flags
	flag.Parse()

	// Validate the paths
	if err := ValidateConfigPath(configPath); err != nil {
		return "", "", err
	}

	return configPath, outputPath, nil
}

func main() {
	// Parse the cli flags
	configPath, outputPath, err := ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	// Create the configuration struct
	config, err := NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Run the tty shell and record it.
	if err := RecordShell(outputPath, config); err != nil {
		log.Fatal(err)
	}
}