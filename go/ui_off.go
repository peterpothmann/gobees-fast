//go:build !ui && !windows && !(darwin && cgo)

package main

import "errors"

// errNoUI is the error of --ui in a build without the window. The window needs Ebitengine. On Windows it is always
// built in, on macOS with cgo (the Xcode command line tools). On Linux Ebitengine needs cgo and the X11 and OpenGL
// headers, which a cluster often lacks, so there the window is built only with the tag ui.
var errNoUI = errors.New("--ui: this gobees is built without the window. On Linux run it with go run -tags ui . --ui " +
	"(needs cgo and the X11 and OpenGL headers), on macOS cgo must be on (Xcode command line tools)")

func watcher(title, screenshot string, shotTurns int) (watchFunc, error) { return nil, errNoUI }

func explainWindow() error { return errNoUI }
