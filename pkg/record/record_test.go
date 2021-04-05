package record

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
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