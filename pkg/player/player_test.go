package player

import (
	"testing"
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