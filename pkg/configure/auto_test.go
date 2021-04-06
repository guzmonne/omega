package configure

import (
	"fmt"
	"math/rand"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAutoUnmarshall(t *testing.T) {
	type Test struct {
		Auto Auto `yaml:"auto"`
	}
	var actual Test
	var value = rand.Int()
	var expected = &Test{Auto(value)}
	var encoded = fmt.Sprintf("auto: %d", value)
	// Should be decoded as is if it's a number
	if err := yaml.Unmarshal([]byte(encoded), &actual); err != nil {
		t.Fatalf(err.Error())
	}
	if *expected != actual {
		t.Errorf("actual = %d; expected = %d", actual, expected)
	}
	// Should be decoded as -1 if it's anything else
	anything := []string{
		"[]",
		"{}",
		"'auto'",
		"True",
	}
	expected = &Test{Auto(-1)}
	for _, thing := range anything {
		encoded = fmt.Sprintf("auto: %s", thing)
		if err := yaml.Unmarshal([]byte(encoded), &actual); err != nil {
			t.Fatalf(err.Error())
		}
		if *expected != actual {
			t.Errorf("actual = %d; expected = %d", actual, expected)
		}
	}
}

func TestAutoMarshall(t *testing.T) {
	type Test struct {
		Auto Auto `yaml:"auto"`
	}
	// Should be encoded into a number if it is an int
	var value = rand.Int()
	var expected = fmt.Sprintf("auto: %d\n", value)
	var decoded = &Test{Auto(value)}
	if actual, err := yaml.Marshal(decoded); err != nil {
		t.Fatalf(err.Error())
	} else {
		if expected != string(actual) {
			t.Errorf("actual:\n%s\nexpected:\n%s", string(actual), expected)
		}
	}
	// Should be encoed into the string "auto" if it is -1
	expected = "auto: auto\n"
	decoded = &Test{Auto(-1)}
	actual, err := yaml.Marshal(decoded)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if expected != string(actual) {
		t.Errorf("actual:\n%s\nexpected:\n%s", string(actual), expected)
	}
}
