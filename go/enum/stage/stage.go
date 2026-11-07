package stage

// Stage of a bee. NetLogo: stage (a string). Decision A in trans/data-model.md: a field, not a component.
type Stage uint8

const (
	Egg Stage = iota
	Larva
	Pupa
	Adult
)

var names = [...]string{"egg", "larva", "pupa", "adult"}

// String returns the NetLogo string.
func (s Stage) String() string {
	return names[s]
}
