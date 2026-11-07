//go:build windows

package ui

import (
	"syscall"
	"time"
	"unsafe"
)

// On Windows the monotonic clock of Go (time.Now) advances in steps of about 0.4 ms, too coarse for parts that take
// microseconds. QueryPerformanceCounter resolves below one microsecond.
var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	qpc      = kernel32.NewProc("QueryPerformanceCounter")
	qpcFreq  = func() int64 {
		var f int64
		kernel32.NewProc("QueryPerformanceFrequency").Call(uintptr(unsafe.Pointer(&f)))
		return f
	}()
)

// now returns a point in time for measuring durations (the performance counter).
func now() time.Duration {
	var c int64
	qpc.Call(uintptr(unsafe.Pointer(&c)))
	return time.Duration(c/qpcFreq)*time.Second + time.Duration(c%qpcFreq*int64(time.Second)/qpcFreq)
}
