package day

import "strconv"

// Abs is an absolute day of the run: NetLogo's ticks, 1 on the first day, 0 in Setup (globals.Ticks).
type Abs int

// OfYear is a day of the year, 1 to 365 (globals.Calendar.Day, 0 in Setup).
type OfYear int

// YearEnd returns the last day of the year of the absolute day d (365 for the days 1 to 365, 0 in Setup), so that
// YearEnd(d) + a day of the year is that day in the next year. NetLogo: 365 * ceiling (ticks / 365).
func YearEnd(d Abs) Abs {
	return 365 * ((d + 364) / 365)
}

// Date is an absolute day that may be unset. The zero value is unset. NetLogo writes NotSetHigh (999999999999999)
// into an unset date, and every comparison of the model with it reads as "the date has not come yet": ticks >
// NotSetHigh is false. Before keeps this meaning.
type Date struct {
	day Abs
	set bool
}

// At returns the date of day d.
func At(d Abs) Date {
	return Date{day: d, set: true}
}

// IsSet reports whether the date is set.
func (d Date) IsSet() bool {
	return d.set
}

// Day returns the day of a set date. It panics for an unset date, which has no day.
func (d Date) Day() Abs {
	if !d.set {
		panic("day: Day of an unset date")
	}
	return d.day
}

// Before reports whether the date is set and lies before day t. False for an unset date, as ticks > NotSetHigh in
// NetLogo.
func (d Date) Before(t Abs) bool {
	return d.set && d.day < t
}

// Is reports whether the date is set and is day t. False for an unset date.
func (d Date) Is(t Abs) bool {
	return d.set && d.day == t
}

// String returns the day, or "unset".
func (d Date) String() string {
	if !d.set {
		return "unset"
	}
	return strconv.Itoa(int(d.day))
}
