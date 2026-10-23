package globals

// Ticks mirrors NetLogo's ticks counter.
//
// NetLogo calls tick at the START of Go [GO-03], so ticks is 1 during the first simulated day and 0 during Setup.
// The ark-tools resource.Tick is incremented AFTER each update and is therefore one less.
// All systems must use this resource, never resource.Tick, wherever the NetLogo code uses ticks.
type Ticks struct {
	Ticks int64
}
