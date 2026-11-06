package sys

import (
	"gobees/comp"
	"gobees/params"
)

// The stimulus reporters of a colony, called by ActivityProc [ACT-02]. All stimuli are 0 or 1.

// egglayingStimulus returns the egg-laying stimulus of a colony [STEG]. sp is the colony's species.
// allAdults > 1 counts all adults, not only workers (finding 45, see trans/netlogo-findings.md#finding-45).
func egglayingStimulus(stores *comp.ColonyStores, counts *comp.ColonyCounts, sp *params.Species) float64 {
	// [STEG-01]
	egglayingStim := 0.0
	if (stores.Pollen > sp.MinPollenStore_g && (counts.Eggs+counts.Larvae) == 0) || counts.Adults > 1 {
		egglayingStim = 1
	}
	return egglayingStim
}

// nursingStimulus returns the nursing (incubation) stimulus of a colony [STNU]. sp is the colony's species.
func nursingStimulus(stores *comp.ColonyStores, counts *comp.ColonyCounts, sp *params.Species) float64 {
	// [STNU-01] NetLogo sets nursingStim to 0 twice, gobees-fast once.
	nursingStim := 0.0
	incubationRequiredToday_kJ := sp.DevQuotaIncubationToday_kJ * float64(counts.Eggs+counts.Larvae+counts.Pupae)
	if incubationRequiredToday_kJ > stores.IncubationToday {
		nursingStim = 1
	}
	return nursingStim
}

// nectarStimulus returns the nectar foraging stimulus of a colony and SETS its ideal energy store [STNE].
func nectarStimulus(stores *comp.ColonyStores, daytime_s, sunrise_s, dailyForagingPeriod_s float64) float64 {
	storeSize_d := 5.0
	minNectarStore_kJ := 20.0
	nectarStimTH := 0.005
	idealEnergyFactor := 6.0
	// [STNE-01]
	// float64(...): no fused multiply-add, see CompetitionPointDateREP
	stores.IdealEnergy = float64(idealEnergyFactor*stores.EnergyNeedToday*storeSize_d) + minNectarStore_kJ
	if stores.IdealEnergy < 0 {
		panic("Negative idealEnergyStore_kJ! (StimForagingNectarREP)")
	}
	// [STNE-02]
	nectarStim := (stores.IdealEnergy - stores.Energy) / stores.IdealEnergy
	if nectarStim > nectarStimTH {
		nectarStim = 1
	} else {
		nectarStim = 0
	}
	// [STNE-03]
	if daytime_s < sunrise_s || daytime_s > sunrise_s+dailyForagingPeriod_s {
		nectarStim = 0
	}
	return nectarStim
}

// pollenStimulus returns the pollen foraging stimulus of a colony and SETS its ideal pollen store [STPO].
// sp is the colony's species.
func pollenStimulus(stores *comp.ColonyStores, sp *params.Species, daytime_s, sunrise_s, dailyForagingPeriod_s float64) float64 {
	storeSize_d := 5.0
	pollenStimTH := 0.005
	// [STPO-01]
	stores.IdealPollen = float64(stores.LarvaePollenNeedToday*storeSize_d) + sp.MinPollenStore_g
	if stores.IdealPollen < 0 {
		panic("Negative idealPollenStore_kJ! (StimForagingPollenREP)")
	}
	// [STPO-02]
	pollenStim := (stores.IdealPollen - stores.Pollen) / stores.IdealPollen
	if pollenStim > pollenStimTH {
		pollenStim = 1
	} else {
		pollenStim = 0
	}
	// [STPO-03]
	if daytime_s < sunrise_s || daytime_s > sunrise_s+dailyForagingPeriod_s {
		pollenStim = 0
	}
	return pollenStim
}
