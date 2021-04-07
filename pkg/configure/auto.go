package configure

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Auto type indicates a value if its bigger or equal to zero
// and -1 when the value is not an int.
type Auto int
// UnmarshallYAML tells the YAML Unmarshal function how to decode the struct.
func (auto *Auto) UnmarshalYAML(value *yaml.Node) error {
	var result int

	if err := value.Decode(&result); err != nil {
		*auto = Auto(-1)
	} else {
		*auto = Auto(result)
	}

	return nil
}

// MarshalYAML tells the YAML Marshal function how to encode the struct.
func (auto Auto) MarshalYAML() (interface{}, error) {
	if auto == Auto(-1) {
		return auto.String(), nil
	}

	return int(auto), nil
}

// String returns Auto's string representation.
func (auto *Auto) String() string {
	if *auto == Auto(-1) {
		return "auto"
	}
	return fmt.Sprintf("%d", *auto)
}