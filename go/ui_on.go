//go:build ui || windows || (darwin && cgo)

package main

import (
	"gobees/obs"
	"gobees/ui"

	"github.com/mlange-42/ark-tools/app"
)

// watcher returns the function that runs a job in the window (--ui), titled with the run file.
func watcher(title, screenshot string) (watchFunc, error) {
	return func(a *app.App, counters *obs.Counters) error {
		return ui.Watch(a, counters, ui.Config{Title: title, Screenshot: screenshot})
	}, nil
}

// explainWindow opens only the window that explains the model (--ui-explain).
func explainWindow() error {
	return ui.RunExplain(0, 0, "")
}
