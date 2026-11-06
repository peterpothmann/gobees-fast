package activity

// Activity is the current task of a bee. NetLogo: activity (a string).
type Activity uint8

const (
	Hibernate Activity = iota
	NestConstruction
	Resting
	Searching
	ReturningEmpty
	ReturningUnhappyNectar // back from a nectar trip, unhappy with the trip: looks for another layer or patch. NetLogo: returningUnhappyN
	ReturningUnhappyPollen // NetLogo: returningUnhappyP
	NectarForaging
	CollectNectar
	BringingNectar
	ReturningHappyNectar // back from a nectar trip with a load, happy with her patch. NetLogo: expForagingN
	PollenForaging
	CollectPollen
	BringingPollen
	ReturningHappyPollen // NetLogo: expForagingP
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
