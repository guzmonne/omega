package player

import (
	"testing"
)

func TestNewPlayOptions(t *testing.T) {
	var speedFactor = 1.0
	var realTiming = true
	var silent = false
	// Should return a new PlayerOptions stuct with correct defaults
	var actual = NewPlayOptions()
	if actual.SpeedFactor != speedFactor {
		t.Errorf("actual = %.2f; expected = %.2f", actual.SpeedFactor, speedFactor)
	}
	if actual.RealTiming != realTiming {
		t.Errorf("actual = %v; expected = %v", actual.RealTiming, realTiming)
	}
	if actual.Silent != silent {
		t.Errorf("actual = %v; expected = %v", actual.Silent, silent)
	}
}