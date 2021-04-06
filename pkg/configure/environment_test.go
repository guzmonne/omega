package configure

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func splitLink(link string, separator string) (string, string) {
	result := strings.Split(link, separator)
	return result[0], result[1]
}

func TestYAMLUnmarshal(t *testing.T) {
	type Test struct {
		Env Environment `yaml:"env"`
	}
	var actual1 Test
	var envs = []string{
		"example=test",
		"something=else",
	}
	var environment = &Environment{Values: envs}
	var expected = &Test{*environment}
	// Should support decoding a map of strings
	var encoded = "env:"
	for _, env := range envs {
		key, value := splitLink(env, "=")
		encoded = encoded + fmt.Sprintf("\n  %s: %s", key, value)
	}
	fmt.Println(encoded)
	if err := yaml.Unmarshal([]byte(encoded), &actual1); err != nil {
		t.Fatalf(err.Error())
	}
	if reflect.DeepEqual(expected, &actual1) == false {
		t.Errorf("actual:\n%s\nexpected:\n%s", &actual1, expected)
	}
	// Should support a slice of strings
	var actual2 Test
	encoded = "env:"
	for _, env := range envs {
		encoded = encoded + fmt.Sprintf("\n  - %s", env)
	}
	fmt.Println(encoded)
	if err := yaml.Unmarshal([]byte(encoded), &actual2); err != nil {
		t.Fatalf(err.Error())
	}
	if reflect.DeepEqual(expected, &actual2) == false {
		t.Errorf("actual:\n%s\nexpected:\n%s", &actual2, expected)
	}
	// If the values are a slice of strings it should only keep those that comply
	// to an environment variable definition.
	expected = &Test{Environment{[]string{"example=test"}}}
	var actual3 Test
	envs = []string{
		"example=test",   // Valid
		"something|else", // Invalid
	}
	encoded = "env:"
	for _, env := range envs {
		encoded = encoded + fmt.Sprintf("\n  - %s", env)
	}
	fmt.Println(encoded)
	err := yaml.Unmarshal([]byte(encoded), &actual3)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if reflect.DeepEqual(expected, &actual3) == false {
		t.Errorf("actual:\n%s\nexpected:\n%s", &actual3, expected)
	}
}