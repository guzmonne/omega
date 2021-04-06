package player

import (
	"reflect"
	"testing"

	"gux.codes/omega/pkg/configure"
	"gux.codes/omega/pkg/record"
)

func TestNewPlayOptions(t *testing.T) {
	var speedFactor = 1
	var realTiming = true
	var silent = false
	// Should return a new PlayerOptions stuct with correct defaults
	var actual = NewPlayOptions()
	if actual.speedFactor != speedFactor {
		t.Errorf("actual = %d; expected = %d", actual.speedFactor, speedFactor)
	}
	if actual.realTiming != realTiming {
		t.Errorf("actual = %v; expected = %v", actual.realTiming, realTiming)
	}
	if actual.silent != silent {
		t.Errorf("actual = %v; expected = %v", actual.silent, silent)
	}
}

func TestNewFrameDelayOptions(t *testing.T) {
	var maxIdleTime = configure.Auto(-1)
	var frameDelay = configure.Auto(-1)
	var speedFactor = 1
	// Should return a new PlayerOptions stuct with correct defaults
	var actual = NewFrameDelayOptions()
	if actual.maxIdleTime != maxIdleTime {
		t.Errorf("actual = %d; expected = %d", actual.maxIdleTime, maxIdleTime)
	}
	if actual.frameDelay != frameDelay {
		t.Errorf("actual = %v; expected = %v", actual.frameDelay, frameDelay)
	}
	if actual.speedFactor != speedFactor {
		t.Errorf("actual = %v; expected = %v", actual.speedFactor, speedFactor)
	}
}

func TestAdjustFrameDelays(t *testing.T) {
	options := NewFrameDelayOptions()
	records := []record.Record{{Delay: 1, Content: "1"}, {Delay: 2, Content: "2"}}
	control := make([]record.Record, 2)
	_ = copy(control, records)

	// Should do nothing if the default options are provided
	adjustFrameDelays(&records, options)
	if !reflect.DeepEqual(records, control) {
		t.Errorf("Records mismatch.\nactual:\n%v\nexpected:\n%v", records, control)
	}

	// Should apply the frameDelay provided by the options if is not -1.
	var frameDelay = 100
	options.frameDelay = configure.Auto(frameDelay)
	for i := range control {
		control[i].Delay = frameDelay
	}
	adjustFrameDelays(&records, options)
	if !reflect.DeepEqual(records, control) {
		t.Errorf("Records mismatch.\nactual:\n%v\nexpected:\n%v", records, control)
	}

	// Should max the delay if maxIdleDelay is not -1
	var maxIdleDelay = 50
	options.frameDelay = configure.Auto(-1)
	options.maxIdleTime = configure.Auto(maxIdleDelay)
	for i := range control {
		control[i].Delay = maxIdleDelay
	}
	adjustFrameDelays(&records, options)
	if !reflect.DeepEqual(records, control) {
		t.Errorf("Records mismatch.\nactual:\n%v\nexpected:\n%v", records, control)
	}
	// Should multiply the delay by the speed factor
	var speedFactor = 2
	options.speedFactor = speedFactor
	for i := range control {
		control[i].Delay = maxIdleDelay * speedFactor
	}
	adjustFrameDelays(&records, options)
	if !reflect.DeepEqual(records, control) {
		t.Errorf("Records mismatch.\nactual:\n%v\nexpected:\n%v", records, control)
	}
}