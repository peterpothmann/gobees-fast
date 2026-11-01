package globals

// WhoCounter hands out NetLogo who numbers. NetLogo numbers all turtles of all breeds with one counter and
// never reuses a number within a run. A turtle that dies right after its creation has used up its number
// (finding 18, see trans/netlogo-findings.md#finding-18), so Next must be called for it too.
type WhoCounter struct {
	next int64
}

// Next returns the who number of the next created turtle.
func (c *WhoCounter) Next() int64 {
	n := c.next
	c.next++
	return n
}

// Landscape holds the globals that CreateFoodsourcesProc [FSRC] sets.
type Landscape struct {
	Scaling_NLpatches_per_m float64 // NetLogo patches per metre, 1 / line 1 of the foodsources file, 8 decimals [FSRC-02]
}
