package input

import (
	"strings"
	"testing"
)

// TestCheckWhole checks that a whole number written as a float is read, and that the error of a fraction names the
// key. yaml.v3 alone would read 12.5 as 12.
func TestCheckWhole(t *testing.T) {
	c, err := ReadConfig(strings.NewReader("species:\n  - batchSize: 12.0\n    seasonStop: 305\n"))
	if err != nil || c.Species[0].BatchSize != 12 || c.Species[0].SeasonStop != 305 {
		t.Errorf("12.0 and 305: got %+v, %v", c.Species, err)
	}
	_, err = ReadConfig(strings.NewReader("species:\n  - batchSize: 12.5\n"))
	if err == nil || !strings.Contains(err.Error(), "species[0].batchSize: 12.5 is not a whole number") {
		t.Errorf("12.5: error %v", err)
	}
	// floats stay floats
	if _, err := ReadConfig(strings.NewReader("species:\n  - growthFactor: 1.88\n")); err != nil {
		t.Errorf("growthFactor 1.88: %v", err)
	}
}
