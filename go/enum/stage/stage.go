package stage

// Stage of a bee: a field, not a component. NetLogo: stage (a string).
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
