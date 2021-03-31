package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

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
func TestNewConfig (t *testing.T) {
	_, err := DefaultConfig()
	if err != nil {
		t.Errorf("DefaultConfig() should't throw an error")
	}
}

// TestWriteConfig tests if the function can successfully write a `Config`
// struct back as a valid YAML file.
func TestWriteConfig(t *testing.T) {
	config, err := NewConfig("./config.yml")
	if err != nil {
		t.Errorf("NewConfig() should't throw an error")
	}
	expected := "/tmp"
	if config.Cwd != expected {
		t.Errorf("config.Cwd = %s; expected %s", config.Cwd, expected)
	}
}
// TestValidateConfigPath test if the function ValidateConfigPath
// validates the provided path correctly.
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