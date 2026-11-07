// Package day holds the types for days. NetLogo counts days with floating-point numbers, gobees-fast with whole
// numbers. Two types keep the absolute day of the run apart from the day of the year, so that the compiler rejects a
// mix of both. Durations in days (ages, life spans) are plain int.
package day
