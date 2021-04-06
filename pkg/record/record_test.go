package record

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"gux.codes/omega/pkg/configure"
)

// WriteRecording function tests
func TestWriteRecording(t *testing.T) {
	var recordingPath = "/tmp/" + fmt.Sprint(rand.Int())
	var config, err = configure.DefaultConfig()
	var records = []Record {
		{0, "something"},
		{1000, "else"},
	}
	if err != nil {
		t.Fatalf(err.Error())
	}
	// Change to a static value of Cwd so the test doesn't fail if
	// run from another folder.
	config.Cwd = "/tmp"
	// Should not thow an error while writing the recording.
	if err := WriteRecording(recordingPath, config, records); err != nil {
		t.Fatalf("Expected WriteRecording not to throw:\n%s", err.Error())
	}
	// Should write the recording to the correct location
	if _, err := os.Stat(recordingPath); err != nil {
		t.Fatalf("Expected file to exist")
	}
	// Should be a valid YAML file
	var expected = `config:
  command: /bin/bash
  cwd: /tmp
  env: []
  cols: auto
  rows: auto
  repeat: 0
  quality: 100
  frameDelay: auto
  maxIdleTimeout: auto
  cursorStyle: block
  fontFamily: Monaco, Lucida Console, Ubuntu Mono, Monospace
  fontSize: 12
  lineHeight: 1
  letterSpacing: 0
records:
  - delay: 0
    content: something
  - delay: 1000
    content: else
`
	actual, err := ioutil.ReadFile(recordingPath)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if a, e := strings.TrimSpace(string(actual)), strings.TrimSpace(expected); a != e {
		t.Errorf("Actual is different than expected:\n%v", diff.LineDiff(e, a))
	}
}

func TestReadRecording(t *testing.T) {
	recordingPath := "/tmp/recording.yml"
	config := &configure.Config{
		Command: "/example",
		Cwd: "/example",
		Env: configure.Environment{
			Values: []string{"environment=variable"},
		},
	}
	records := []Record{{0, "0"}, {1, "1"}}
	recording := &Recording{*config, records}
	// Make sure there are no existing files
	if err := os.RemoveAll(recordingPath); err != nil {
		t.Fatalf(err.Error())
	}
	// Clean everything after the test stops
	defer func() {
		if err := os.RemoveAll(recordingPath); err != nil {
			panic(err)
		}
	}()

	// Write a sample config to the test file
	if err := WriteRecording(recordingPath, config, records); err != nil {
		t.Fatalf(err.Error())
	}

	// Read the recording and check that the contents match
	content, err := ReadRecording(recordingPath)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !reflect.DeepEqual(content, recording) {
		t.Errorf("\nactual:\n\t%v\nexpected:\n\t%v", content, recording)
	}
}