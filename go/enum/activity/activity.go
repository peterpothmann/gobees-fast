package activity

// Activity is the current task of a bee. NetLogo: activity (a string).
type Activity uint8

const (
	Hibernate Activity = iota
	NestConstruction
	Resting
	Searching
	ReturningEmpty
	ReturningUnhappyN
	ReturningUnhappyP
	NectarForaging
	CollectNectar
	BringingNectar
	ExpForagingN
	PollenForaging
	CollectPollen
	BringingPollen
	ExpForagingP
	Egglaying
	Nursing
	Emerging
)

var names = [...]string{
	"hibernate", "nestConstruction", "resting", "searching", "returningEmpty", "returningUnhappyN",
	"returningUnhappyP", "nectarForaging", "collectNectar", "bringingNectar", "expForagingN", "pollenForaging",
	"collectPollen", "bringingPollen", "expForagingP", "egglaying", "nursing", "emerging",
}

// String returns the NetLogo string.
func (a Activity) String() string {
	return names[a]
}
