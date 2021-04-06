package configure

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
)

// Removes anything on the path.
func cleanup(path string) error {
	// Delete the temp folder
	if err := os.RemoveAll(path); err != nil {
		return err
	}

	return nil
}

// TestDefaultConfig tests if the function can create a new Config struct.
func TestDefaultConfig(t *testing.T) {
	_, err := DefaultConfig()
	if err != nil {
		t.Errorf("DefaultConfig() should't throw an error")
	}
}

func TestNewConfig(t *testing.T) {
	config, err := NewConfig("./config.yml")
	if err != nil {
		t.Errorf("NewConfig() should't throw an error")
	}
	expected := "/tmp"
	if config.Cwd != expected {
		t.Errorf("config.Cwd = %s; expected %s", config.Cwd, expected)
	}
}

func TestValidateConfigPath(t *testing.T) {
	var path string
	var expected string
	var err error
	// Should fail if the path is a directory
	path = "/tmp"
	expected = "'/tmp' is a directory, not a file."
	err = ValidateConfigPath(path)
	if err.Error() != expected {
		t.Errorf("err.Error() = %s; expected %s", err.Error(), expected)
	}
	// Should fail if the path is not valid
	path = ""
	expected = "stat : no such file or directory"
	err = ValidateConfigPath(path)
	if err.Error() != expected {
		t.Errorf("err.Error() = %s; expected %s", err.Error(), expected)
	}
}

func TestWriteConfig(t *testing.T) {
	configPath := "/tmp/config.yml"
	// Make sure there is nothing on the test config path
	if err := cleanup(configPath); err != nil {
		t.Fatalf(err.Error())
	}
	// Cleanup
	defer cleanup(configPath)

	// Create a test Config
	config := &Config{
		Command: "/example",
		Cwd: "/example",
		Env: *&Environment{
			Values: []string{"environment=variable"},
		},
	}

	// Should not throw an error while writing
	if err := WriteConfig(configPath, *config); err != nil {
		t.Fatalf(err.Error())
	}

	// Should have the right content written to the file
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Fatalf(err.Error())
	}
	expected := `command: /example
cwd: /example
env:
  values:
    - environment=variable
cols: 0
rows: 0
repeat: 0
quality: 0
frameDelay: 0
maxIdleTimeout: 0
cursorStyle: ""
fontFamily: ""
fontSize: 0
lineHeight: 0
letterSpacing: 0`
	if a, e := strings.TrimSpace(string(content)), strings.TrimSpace(expected); a != e {
		t.Errorf("Actual is different than expected:\n%v", diff.LineDiff(e, a))
	}
}

func TestReadConfig(t *testing.T) {
	configPath := "/tmp/test_read_config.yml"
	config := &Config{
		Command: "/example",
		Cwd: "/example",
		Env: *&Environment{
			Values: []string{"environment=variable"},
		},
	}
	// Make sure there are no existing files
	if err := cleanup(configPath); err != nil {
		t.Fatalf(err.Error())
	}
	// Clean everything after the test stops
	//defer cleanup(configPath)

	// Write a sample config to the test file
	if err := WriteConfig(configPath, *config); err != nil {
		t.Fatalf(err.Error())
	}

	// Read the configuration and check that the contents match
	conf, err := ReadConfig(configPath)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if conf != config {
		t.Errorf("\nactual:\n\t%v\nexpected:\n\t%v", conf, config)
	}
}

const configFolder = "/tmp/test"
const configPath = configFolder + "/" + CONFIG_FILENAME
const dummyContent = "test"

func createDefaultConfig() string {
	cwd := os.Getenv("HOME") + "/.omega"
	return `command: /bin/bash
cwd: ` + cwd + `
env:
  values: []
cols: -1
rows: -1
repeat: 0
quality: 100
frameDelay: -1
maxIdleTimeout: -1
cursorStyle: block
fontFamily: Monaco, Lucida Console, Ubuntu Mono, Monospace
fontSize: 12
lineHeight: 1
letterSpacing: 0
`
}

func readConfigFile() (string, error) {
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func TestInit(t *testing.T) {
	// Make sure the files don't exist
	cleanup(configFolder)
	// Run after the test complete
	defer cleanup(configFolder)

	// The config file shouldn't exist
	_, err := os.Stat(configPath);
	if os.IsNotExist(err) == false {
		t.Fatalf("should not have thrown error:\n%s", err.Error())
	}

	// Should create the config file
	if err := Init(configFolder); err != nil {
		t.Fatalf("should not have thrown error:\n%s", err.Error())
	}
	content, err := readConfigFile()
	if err != nil {
		t.Fatalf(err.Error())
	}
	// Should write the default config if the file is not present
	defaultConfig := createDefaultConfig()
	if a, e := strings.TrimSpace(string(content)), strings.TrimSpace(defaultConfig); a != e {
		t.Errorf("Actual is different than expected:\n%v", diff.LineDiff(e, a))
	}

	// Clean the file to check if Init overwrites it
	if file, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755); err != nil {
		t.Fatalf(err.Error())
	} else {
		if _, err := file.WriteString(dummyContent); err != nil {
			t.Fatalf(err.Error())
		}
		file.Close()
	}
	// Should contain the dummy content
	content, err = readConfigFile()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if a, e := strings.TrimSpace(string(content)), strings.TrimSpace(dummyContent); a != e {
		t.Errorf("Actual is different than expected:\n%v", diff.LineDiff(e, a))
	}

	// Should do nothing if the file exists
	if err := Init(configFolder); err != nil {
		t.Fatalf("should not have thrown error:\n%s", err.Error())
	}
	content, err = readConfigFile()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if a, e := strings.TrimSpace(string(content)), strings.TrimSpace(dummyContent); a != e {
		t.Errorf("Actual is different than expected:\n%v", diff.LineDiff(e, a))
	}
}