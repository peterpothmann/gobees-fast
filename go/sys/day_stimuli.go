package sys

import (
	"gobees/comp"
	"gobees/params"
)

// The stimuli of a colony, computed by activities.act. All stimuli are 0 or 1.

// egglayingStimulus returns the egg-laying stimulus of a colony. sp is the colony's species.
// counts.Adults > 1 counts all adults, not only workers.
func egglayingStimulus(stores *comp.ColonyStores, counts *comp.ColonyCounts, sp *params.Species) float64 {
	egglayingStim := 0.0
	if (stores.Pollen > sp.MinPollenStore && (counts.Eggs+counts.Larvae) == 0) || counts.Adults > 1 {
		egglayingStim = 1
	}
	return egglayingStim
}

// nursingStimulus returns the nursing (incubation) stimulus of a colony. sp is the colony's species.
func nursingStimulus(stores *comp.ColonyStores, counts *comp.ColonyCounts, sp *params.Species) float64 {
	nursingStim := 0.0
	incubationRequiredToday := sp.DailyIncubationQuota * float64(counts.Eggs+counts.Larvae+counts.Pupae) // [kJ]
	if incubationRequiredToday > stores.IncubationToday {
		nursingStim = 1
	}
	return nursingStim
}

// nectarStimulus returns the nectar foraging stimulus of a colony and SETS its ideal energy store.
// daytime, sunrise and foragingPeriod are in [s].
func nectarStimulus(stores *comp.ColonyStores, c *params.StimulusConstants, daytime, sunrise,
	foragingPeriod float64) float64 {
	storeSize := c.NectarStoreSize     // [d]
	minNectarStore := c.MinNectarStore // [kJ]
	nectarStimTH := c.NectarThreshold
	idealEnergyFactor := c.IdealEnergyFactor
	// float64(...): no fused multiply-add, so the result equals NetLogo (Java)
	stores.IdealEnergy = float64(idealEnergyFactor*stores.EnergyNeedToday*storeSize) + minNectarStore
	if stores.IdealEnergy < 0 {
		panic("Negative IdealEnergy! (nectarStimulus)")
	}
	nectarStim := (stores.IdealEnergy - stores.Energy) / stores.IdealEnergy
	if nectarStim > nectarStimTH {
		nectarStim = 1
	} else {
		nectarStim = 0
	}
	if daytime < sunrise || daytime > sunrise+foragingPeriod {
		nectarStim = 0
	}
	return nectarStim
}

// pollenStimulus returns the pollen foraging stimulus of a colony and SETS its ideal pollen store.
// sp is the colony's species.
// daytime, sunrise and foragingPeriod are in [s].
func pollenStimulus(stores *comp.ColonyStores, sp *params.Species, c *params.StimulusConstants, daytime, sunrise,
	foragingPeriod float64) float64 {
	storeSize := c.PollenStoreSize // [d]
	pollenStimTH := c.PollenThreshold
	stores.IdealPollen = float64(stores.LarvaePollenNeedToday*storeSize) + sp.MinPollenStore
	if stores.IdealPollen < 0 {
		panic("Negative IdealPollen! (pollenStimulus)")
	}
	pollenStim := (stores.IdealPollen - stores.Pollen) / stores.IdealPollen
	if pollenStim > pollenStimTH {
		pollenStim = 1
	} else {
		pollenStim = 0
	}
	if daytime < sunrise || daytime > sunrise+foragingPeriod {
		pollenStim = 0
	}
	return pollenStim
}
