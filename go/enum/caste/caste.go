// Package caste contains the values of the NetLogo bee variable caste.
package caste

// Caste of a bee. NetLogo: caste (a string).
type Caste uint8

const (
	Undefined Caste = iota // brood whose caste is not yet decided
	Queen
	Worker
	Male
)

var names = [...]string{"undefined", "queen", "worker", "male"}

// String returns the NetLogo string.
func (c Caste) String() string {
	return names[c]
}
