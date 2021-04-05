package configure

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"gopkg.in/yaml.v3"
)

// Struct to test the Auto Unmarshall and Marshall overrides.
type TestAuto struct {
	Test Auto `yaml:"test"`
}
// Struct to test the Environment Unmarshall and Marshall overrides.
type TestEnvironment struct {
	Env Environment `yaml:"env"`
}
// TestYAMLUnmarshal tests if the Unmarshall overrides for YAML
// encoded strings works.
func TestYAMLUnmarshal(t *testing.T) {
	var expected = Auto(rand.Int())
	var a TestAuto
	// If the value of Auto is a number it should leave it as is.
	yaml.Unmarshal([]byte("test: " + fmt.Sprint(expected)), &a)
	if a.Test != expected {
		t.Errorf("a.Test = %d; expected %d", a.Test, expected)
	}
	// If the value of Auto is anything else is should be -1.
	yaml.Unmarshal([]byte("test: auto"), &a)
	if a.Test != -1 {
		t.Errorf("a.Test = %d; expected %d", a.Test, -1)
	}
	yaml.Unmarshal([]byte("test: yes"), &a)
	if a.Test != -1 {
		t.Errorf("a.Test = %d; expected %d", a.Test, -1)
	}
	yaml.Unmarshal([]byte("test: {\"something\": \"cool\"}"), &a)
	if a.Test != -1 {
		t.Errorf("a.Test = %d; expected %d", a.Test, -1)
	}
	yaml.Unmarshal([]byte("test: [1, 2, 3]"), &a)
	if a.Test != -1 {
		t.Errorf("a.Test = %d; expected %d", a.Test, -1)
	}
	// If the value of environment is undefined is should return an empty `[]string`
	var e TestEnvironment
	yaml.Unmarshal([]byte("something: else"), &e)
	if len(e.Env.Values) != 0 {
		t.Errorf("len(e.Env.Values) = %d; expected 0", len(e.Env.Values))
	}
	// If the value is not a `map[string]string` it sould return an empty `[]string`
	yaml.Unmarshal([]byte("env: else"), &e)
	if len(e.Env.Values) != 0 {
		t.Errorf("len(e.Env.Values) = %d; expected 0", len(e.Env.Values))
	}
	// If the value is a `map[string]string` it sould return a `[]string` filled with string of type `key=value`
	environments := []string{"something=awesome", "test=example"}
	yaml.Unmarshal([]byte("env:\n  something: awesome\n  test: example"), &e)
	if reflect.DeepEqual(e.Env.Values, environments) == false {
		t.Errorf("e.Env.Values = %s; expected %s", e.Env.Values, environments)
	}
	// If the value is a `[]string` it sould return a `[]string`
	environments = []string{"something=awesome", "test=example"}
	yaml.Unmarshal([]byte("env:\n  - something=awesome\n  - test=example"), &e)
	if reflect.DeepEqual(e.Env.Values, environments) == false {
		t.Errorf("e.Env.Values = %s; expected %s", e.Env.Values, environments)
	}
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

	// Create a random Config
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
	// TODO
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

func cleanup(path string) error {
	// Delete the temp folder
	if err := os.RemoveAll(path); err != nil {
		return err
	}

	return nil
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