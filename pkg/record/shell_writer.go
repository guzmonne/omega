package record

import (
	"bytes"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v3"
)

// RecordWriter writes inputs as Records
type ShellWriter struct {
	specification *ShellSpecification
	timestamp time.Time
	records []Record
}

// NewShellWriter creates a new default ShellWriter.
func NewShellWriter(specification *ShellSpecification) *ShellWriter {
	return &ShellWriter{
		specification: specification,
		timestamp: time.Now(),
		records: make([]Record, 0),
	}
}

// now updates the timestamp to the current time.
func (writer *ShellWriter) now() {
	writer.timestamp = time.Now()
}

// Write the input as a new Record. If the time since the last Record is less
// than MIN_DELAY, then it modifies the last record appending the new bytes.
func (writer *ShellWriter) Write(input []byte) (int, error) {
	// Calculate the delay since the last record
	var delay int

	defer writer.now()

	if len(writer.records) == 0 {
		record := &Record{0, string(input)}
		writer.records = append(writer.records, *record)
		return len(input), nil
	}

	// If the delay is less than MIN_DELAY then we get the previous record
	// and update it. Else we create a new one.
	delay = int(time.Since(writer.timestamp) / 1000 / 1000)
	if delay < writer.specification.MinDelay {
		previous := &writer.records[len(writer.records) - 1]
		previous.Content = previous.Content + string(input)
	} else {
		record := &Record{delay, string(input)}
		writer.records = append(writer.records, *record)
	}

	// Comply with the Writer interface
	return len(input), nil
}

// Dump writes the Recording to a YAML file on the path provided
// by the shell specification.
func (writer ShellWriter) Dump() error {
	var file bytes.Buffer

	// Create a custom YAML encoder
	encoder := yaml.NewEncoder(&file)
	encoder.SetIndent(2)

	// Marshall to YAML the Recording struct
	err := encoder.Encode(writer.records)
	if err != nil {
		return err
	}

	// Write the recording file
	err = ioutil.WriteFile(writer.specification.OutputPath, file.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}