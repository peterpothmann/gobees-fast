//go:build (ui || (darwin && cgo)) && !windows

package ui

import "time"

var clockStart = time.Now()

// now returns a point in time for measuring durations (the monotonic clock, which resolves nanoseconds on Linux and
// macOS).
func now() time.Duration {
	return time.Since(clockStart)
}
