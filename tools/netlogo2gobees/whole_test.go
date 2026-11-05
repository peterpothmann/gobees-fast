package main

import (
	"reflect"
	"testing"
)

// TestSetFieldWhole checks that an int field takes a whole number and rejects a fraction (step F8). NetLogo writes
// every number as a float, so 10000 and 10000.0 are both whole.
func TestSetFieldWhole(t *testing.T) {
	var n int
	v := reflect.ValueOf(&n).Elem()
	for _, literal := range []string{"10000", "10000.0"} {
		if err := setField(v, literal); err != nil || n != 10000 {
			t.Errorf("%s: got %d, %v", literal, n, err)
		}
	}
	if err := setField(v, "10.5"); err == nil {
		t.Error("10.5: want an error")
	}
}
