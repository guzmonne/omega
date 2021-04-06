package configure

import (
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Environment handles a list of environment variables.
type Environment struct {
	Values []string
}

func unmarshallMap(env *Environment, value *yaml.Node) error {
	var result map[string]string

	if err := value.Decode(&result); err != nil {
		return err
	}

	for key, value := range result {
		env.Values = append(env.Values, fmt.Sprintf("%s=%s", key, value))
	}

	return nil
}

func unmarshallSlice(env *Environment, value *yaml.Node) error {
	var result []string

	if err := value.Decode(&result); err != nil {
		return err
	}

	for _, value := range result {
		matched, err := regexp.Match(`.*=.*`, []byte(value))
		if err != nil {
			continue
		}
		if matched {
			env.Values = append(env.Values, value)
		}
	}

	return nil
}

// UnmarshallYAML tells the YAML Unmarshal function how to decode the struct.
func (env *Environment) UnmarshalYAML(value *yaml.Node) error {
	// Try to unmarshall as if it is a map of strings
	if err := unmarshallMap(env, value); err == nil {
		return nil
	}

	// Try to unmarshall as if it is a slice of strings
	if err := unmarshallSlice(env, value); err != nil {
		return err
	}

	return nil
}