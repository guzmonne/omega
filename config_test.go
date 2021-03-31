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
// TestYAMLUnmarshal test if the Unmarshall overrides for YAML
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
}