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
	var config = configure.NewConfig()
	var records = []Record {
		{0, "something"},
		{1000, "else"},
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

func TestNewFrameDelayOptions(t *testing.T) {
	var maxIdleTime = configure.Auto(-1)
	var frameDelay = configure.Auto(-1)
	var speedFactor = 1.0
	// Should return a new PlayerOptions stuct with correct defaults
	var actual = NewFrameDelayOptions()
	if actual.MaxIdleTime != maxIdleTime {
		t.Errorf("actual = %d; expected = %d", actual.MaxIdleTime, maxIdleTime)
	}
	if actual.FrameDelay != frameDelay {
		t.Errorf("actual = %v; expected = %v", actual.FrameDelay, frameDelay)
	}
	if actual.SpeedFactor != speedFactor {
		t.Errorf("actual = %v; expected = %v", actual.SpeedFactor, speedFactor)
	}
}

func TestAdjustFrameDelays(t *testing.T) {
	t.Run("should do nothing if the default options are provided", func(t *testing.T) {
		options := NewFrameDelayOptions()
		config := configure.NewConfig()
		records := []Record{{Delay: 1, Content: "1"}, {Delay: 2, Content: "2"}}
		recording := &Recording{Config: *config, Records: records}
		control := make([]Record, 2)
		_ = copy(control, records)

		// Should do nothing if the default options are provided
		recording.AdjustFrameDelays(options)
		if !reflect.DeepEqual(records, control) {
			t.Errorf("Records mismatch.\nactual:\n%v\nexpected:\n%v", records, control)
		}
	})

	t.Run("should apply the frameDelay provided by the options if it's not -1", func(t *testing.T) {
		var frameDelay = 100
		config := configure.NewConfig()
		options := NewFrameDelayOptions()
		options.FrameDelay = configure.Auto(frameDelay)
		records := []Record{{Delay: 1, Content: "1"}, {Delay: 2, Content: "2"}}
		recording := &Recording{Config: *config, Records: records}
		control := make([]Record, 2)
		_ = copy(control, records)
		for i := range control {
			control[i].Delay = frameDelay
		}
		recording.AdjustFrameDelays(options)
		if !reflect.DeepEqual(records, control) {
			t.Errorf("Records mismatch.\nactual:\n%v\nexpected:\n%v", records, control)
		}
	})

	t.Run("should max the delay if maxIdleDelay is not -1", func(t *testing.T) {
		var maxIdleDelay = 50
		config := configure.NewConfig()
		options := NewFrameDelayOptions()
		options.MaxIdleTime = configure.Auto(maxIdleDelay)
		records := []Record{{Delay: 100, Content: "1"}, {Delay: 200, Content: "2"}}
		recording := &Recording{Config: *config, Records: records}
		control := make([]Record, 2)
		_ = copy(control, records)
		for i := range control {
			control[i].Delay = maxIdleDelay
		}
		recording.AdjustFrameDelays(options)
		if !reflect.DeepEqual(records, control) {
			t.Errorf("Records mismatch.\nactual:\n%v\nexpected:\n%v", records, control)
		}
	})

	t.Run("should multiply the delay by the speed factor", func(t *testing.T) {
		var speedFactor = 2.0
		config := configure.NewConfig()
		options := NewFrameDelayOptions()
		options.SpeedFactor = speedFactor
		records := []Record{{Delay: 1, Content: "1"}, {Delay: 2, Content: "2"}}
		recording := &Recording{Config: *config, Records: records}
		control := make([]Record, 2)
		_ = copy(control, records)
		for i := range control {
			control[i].Delay = int(float64(control[i].Delay) * speedFactor)
		}
		recording.AdjustFrameDelays(options)
		if !reflect.DeepEqual(records, control) {
			t.Errorf("Records mismatch.\nactual:\n%v\nexpected:\n%v", records, control)
		}
	})

	t.Run("should multiply the delay by the speed factor", func(t *testing.T) {
		var speedFactor = 0.5
		config := configure.NewConfig()
		options := NewFrameDelayOptions()
		options.SpeedFactor = speedFactor
		records := []Record{{Delay: 10, Content: "1"}, {Delay: 100, Content: "2"}}
		recording := &Recording{Config: *config, Records: records}
		control := make([]Record, 2)
		_ = copy(control, records)
		for i := range control {
			control[i].Delay = int(float64(control[i].Delay) * speedFactor)
		}
		recording.AdjustFrameDelays(options)
		if !reflect.DeepEqual(records, control) {
			t.Errorf("Records mismatch.\nactual:\n%v\nexpected:\n%v", records, control)
		}
	})
}