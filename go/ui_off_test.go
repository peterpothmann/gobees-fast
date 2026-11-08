//go:build !ui && !windows && !(darwin && cgo)

package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// TestUIWithoutTag: --ui in a build without the tag ui is an error before anything is written.
func TestUIWithoutTag(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out")
	c := cli{UI: true, Output: out}
	if err := c.run(io.Discard); !errors.Is(err, errNoUI) {
		t.Fatalf("error %v, want %v", err, errNoUI)
	}
	if _, err := os.Stat(out); err == nil {
		t.Error("output written")
	}
	c = cli{UIExplain: true}
	if err := c.run(io.Discard); !errors.Is(err, errNoUI) {
		t.Errorf("--ui-explain: error %v, want %v", err, errNoUI)
	}
}
