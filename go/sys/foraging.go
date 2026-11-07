package sys

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/event"
	"gobees/globals"
	"gobees/params"
	"gobees/util"

	"github.com/mlange-42/ark/ecs"
)

// forager runs the foraging trips [FOR] and the procedures they call. Called by ActivityProc [ACT-05].
// Creates no entities. A forager that dies on her trip is removed by DieProc.
//
// NetLogo: ForagingProc, Foraging_searchingProc, DetectionProbREP, Foraging_SortKnownPatchesListREP,
// Foraging_collectNectarPollenProc, HandlingTime_s_REP, Foraging_bestLayerREP, Foraging_PatchChoiceProc,
// Foraging_costs&choiceProc, Foraging_unloadingProc.
type forager struct {
	rng       *rand.Rand
	file      *params.ParameterFile
	constants *params.Constants
	derived   *globals.Derived
	landscape *globals.Landscape
	death     beeDeath
	trip      ecs.Event // event.ForagingTrip
	died      ecs.Event // event.DiedForaging

	beeMap        *beeMapper
	colonyMap     *ecs.Map4[comp.Position, comp.Colony, comp.ColonyStores, comp.ColonyRange]
	foodsourceMap *ecs.Map4[comp.Position, comp.Foodsource, comp.FoodsourceState, comp.Layer]
	speciesMap    *speciesTable
}

func newForager(w *ecs.World, rng *rand.Rand) forager {
	return forager{
		rng:           rng,
		file:          ecs.GetResource[params.ParameterFile](w),
		constants:     ecs.GetResource[params.Constants](w),
		derived:       ecs.GetResource[globals.Derived](w),
		landscape:     ecs.GetResource[globals.Landscape](w),
		death:         newBeeDeath(w),
		trip:          w.Event(event.ForagingTrip).For(ecs.C[comp.Bee]()),
		died:          w.Event(event.DiedForaging).For(ecs.C[comp.Bee]()),
		beeMap:        newBeeMapper(w),
		colonyMap:     ecs.NewMap4[comp.Position, comp.Colony, comp.ColonyStores, comp.ColonyRange](w),
		foodsourceMap: ecs.NewMap4[comp.Position, comp.Foodsource, comp.FoodsourceState, comp.Layer](w),
		speciesMap:    newSpeciesTable(w),
	}
}

// ForagingProc lets the bee e make one foraging trip [FOR]. The bee may die on the trip.
func (f *forager) ForagingProc(w *ecs.World, e ecs.Entity) {
	_, b, _, _, fo, _ := f.beeMap.Get(e)
	// [FOR-01] totalForagingTripsToday: output only, counted by the observers (step F4)
	f.trip.Emit(e)
	// [FOR-02]
	if b.Activity == activity.PollenForaging {
		fo.PollenForager = true
		fo.CurrentFoodsource = fo.PollensourceToGoTo
	} else {
		fo.PollenForager = false
		fo.CurrentFoodsource = fo.NectarsourceToGoTo
	}
	// [FOR-03]
	if fo.CurrentFoodsource.IsZero() {
		b.Activity = activity.Searching
	} else if fo.PollenForager {
		b.Activity = activity.CollectPollen
	} else {
		b.Activity = activity.CollectNectar
	}
	// [FOR-04] activityList: GUI log, not translated.
	f.Foraging_searchingProc(e)
	f.Foraging_collectNectarPollenProc(e)
	if !f.Foraging_costsAndChoiceProc(w, e) {
		return // died on the trip: DieProc ends her code (finding 28, see trans/netlogo-findings.md#finding-28)
	}
	f.Foraging_unloadingProc(e)
	// [FOR-05]
	if fo.PollenLoadSquadron_g+fo.NectarLoadSquadron_kJ > 0 {
		panic("Bee carries pollen or nectar after unloading! (ForagingProc)")
	}
}

// Foraging_SortKnownPatchesListREP returns the known masterpatches of the bee e without duplicates, sorted by
// distance to her colony [SORT].
func (f *forager) Foraging_SortKnownPatchesListREP(e ecs.Entity, knownPatchesList []ecs.Entity) []ecs.Entity {
	_, b, _, _, _, _ := f.beeMap.Get(e)
	// [SORT-01] remove-duplicates keeps the first occurrence, sort-by is stable
	newList := make([]ecs.Entity, 0, len(knownPatchesList))
	for _, p := range knownPatchesList {
		if !slices.Contains(newList, p) {
			newList = append(newList, p)
		}
	}
	colPos, _, _, _ := f.colonyMap.Get(b.ColonyID)
	distance := func(p ecs.Entity) float64 {
		pos, _, _, _ := f.foodsourceMap.Get(p)
		return util.Distance(colPos.X, colPos.Y, pos.X, pos.Y)
	}
	slices.SortStableFunc(newList, func(p1, p2 ecs.Entity) int {
		d1, d2 := distance(p1), distance(p2)
		if d1 < d2 {
			return -1
		}
		if d2 < d1 {
			return 1
		}
		return 0
	})
	return newList
}

// Foraging_searchingProc lets a searching bee e try to find a flower patch [SRCH].
func (f *forager) Foraging_searchingProc(e ecs.Entity) {
	_, b, _, _, fo, _ := f.beeMap.Get(e)
	// [SRCH-01]
	if b.Activity == activity.Searching {
		chosenMasterpatch := ecs.Entity{}
		colPos, _, _, colRange := f.colonyMap.Get(b.ColonyID)
		xcol, ycol := colPos.X, colPos.Y
		var myMasterpatchesWithFoodList []ecs.Entity
		if fo.PollenForager {
			myMasterpatchesWithFoodList = colRange.MasterpatchesWithPollenlayersInFlowerAndRangeList
		} else {
			myMasterpatchesWithFoodList = colRange.MasterpatchesWithNectarlayersInFlowerAndRangeList
		}
		// [SRCH-02] one draw per candidate in random order, the last detected one is chosen
		candidates := slices.Clone(myMasterpatchesWithFoodList)
		util.Shuffle(f.rng, candidates)
		for _, p := range candidates {
			if f.rng.Float64() < f.DetectionProbREP(p, xcol, ycol) {
				chosenMasterpatch = p
			}
		}
		// [SRCH-03]
		if !chosenMasterpatch.IsZero() {
			fo.CurrentFoodsource = f.Foraging_bestLayerREP(e, chosenMasterpatch)
			if !fo.CurrentFoodsource.IsZero() {
				if fo.PollenForager {
					fo.KnownMasterpatchesPollenList = f.Foraging_SortKnownPatchesListREP(e,
						append([]ecs.Entity{chosenMasterpatch}, fo.KnownMasterpatchesPollenList...))
					b.Activity = activity.CollectPollen
				} else {
					fo.KnownMasterpatchesNectarList = f.Foraging_SortKnownPatchesListREP(e,
						append([]ecs.Entity{chosenMasterpatch}, fo.KnownMasterpatchesNectarList...))
					b.Activity = activity.CollectNectar
				}
			} else {
				b.Activity = activity.ReturningEmpty // the patch was emptied earlier today
			}
		} else {
			b.Activity = activity.ReturningEmpty
		}
	}
	// [SRCH-04]
	if b.Activity == activity.Searching {
		panic("Bee is still searching! (Foraging_searchingProc)")
	}
}

// DetectionProbREP returns the probability that a bee of the colony at (xcol, ycol) detects the masterpatch
// patchWho on a search flight [DET].
func (f *forager) DetectionProbREP(patchWho ecs.Entity, xcol, ycol float64) float64 {
	// [DET-01]
	pos, fs, _, _ := f.foodsourceMap.Get(patchWho)
	dist_m := util.Distance(pos.X, pos.Y, xcol, ycol) / f.landscape.Scaling_NLpatches_per_m
	patchRadius_m := fs.Radius_m
	relevantDistance_m := dist_m - patchRadius_m
	if relevantDistance_m < 0 {
		relevantDistance_m = 0
	}
	detProb := math.Pow(math.E, f.file.Lambda_detectProb*relevantDistance_m) // NetLogo: e ^ x
	// [DET-02] NetLogo sets detProb to 0 beyond ForagingRangeMax_m. Never true here: the colony's lists hold only
	// masterpatches in range [PRNG] (removed in step F3).
	// [DET-03]
	return precision(detProb, 10)
}

// Foraging_collectNectarPollenProc lets the bee e take a load from her foodsource [COLL].
func (f *forager) Foraging_collectNectarPollenProc(e ecs.Entity) {
	_, b, _, _, fo, _ := f.beeMap.Get(e)
	// [COLL-01] NetLogo also checks for "expForagingP", which the activity never is here (removed in step F3)
	if b.Activity == activity.CollectPollen {
		_, _, state, _ := f.foodsourceMap.Get(fo.CurrentFoodsource)
		if state.Pollen_g > 0 {
			fo.PollenLoadSquadron_g = min(fo.PollenPellets_g*float64(b.Number), state.Pollen_g)
			b.Activity = activity.BringingPollen
			memoPollenLoad := fo.PollenLoadSquadron_g
			state.Pollen_g = state.Pollen_g - memoPollenLoad // cumulPollenVisits: never read, removed in step F3
		} else {
			b.Activity = activity.ReturningEmpty
		}
	}
	// [COLL-02] NetLogo also checks for "expForagingN", which the activity never is here (removed in step F3)
	if b.Activity == activity.CollectNectar {
		_, fs, state, _ := f.foodsourceMap.Get(fo.CurrentFoodsource)
		if state.Nectar_myl > 0 {
			nectarRemoved := min(fo.Cropvolume_myl*float64(b.Number), state.Nectar_myl)
			fo.NectarLoadSquadron_kJ = nectarRemoved * f.constants.EnergySucrose_kJ_per_mymol * fs.NectarConcentration_mol_per_l
			b.Activity = activity.BringingNectar
			state.Nectar_myl = state.Nectar_myl - nectarRemoved // cumulNectarVisits: never read, removed in step F3
		} else {
			b.Activity = activity.ReturningEmpty
		}
	}
}

// HandlingTime_s_REP returns the time [s] the bee e needs to collect one full load at the foodsource myPatch
// [HAND]. pollenPatch selects pollen or nectar.
func (f *forager) HandlingTime_s_REP(e ecs.Entity, myPatch ecs.Entity, pollenPatch bool) float64 {
	_, b, _, _, fo, _ := f.beeMap.Get(e)
	_, fs, state, _ := f.foodsourceMap.Get(myPatch)
	maxHandlingTime_s := 60.0 * 60
	handlingTime_s := -999.0
	fillingLevel := 0.0

	if !pollenPatch {
		// [HAND-01]
		if fs.NectarMax_myl > 0 {
			fillingLevel = state.Nectar_myl / fs.NectarMax_myl
		}
		W_beeWeight_g := b.Weight_mg / 1000
		G_lengthGlossa_mm := fo.GlossaLength_mm
		C_CorollaDepth_mm := fs.CorollaDepth_mm
		V_nectarVolume_myl := fs.NectarFlowerVolume_myl
		// float64(...): no fused multiply-add, see CompetitionPointDateREP
		Ta_accessTime_s := 0.3 + float64(0.04*C_CorollaDepth_mm)
		numerator := log10(V_nectarVolume_myl + 1)
		num := 0.3 * math.Pow(W_beeWeight_g, 0.3333) * G_lengthGlossa_mm
		base := 1.41 - C_CorollaDepth_mm/G_lengthGlossa_mm
		if base < 0.001 {
			base = 0.001
		}
		den := math.Pow(base, -0.4) - float64(0.3*Ta_accessTime_s)
		denominator := log10(num/den + 1)
		Ti_ingestionTime_s := numerator / denominator

		handlingTimePerFlower_s := maxHandlingTime_s
		if fillingLevel > 0 {
			handlingTimePerFlower_s = (fs.InterFlowerTime_s+Ta_accessTime_s)/fillingLevel + Ti_ingestionTime_s
		}
		flowersVisited := 1.0
		if V_nectarVolume_myl < fo.Cropvolume_myl {
			flowersVisited = fo.Cropvolume_myl / V_nectarVolume_myl
		}
		handlingTime_s = handlingTimePerFlower_s * flowersVisited
	} else {
		// [HAND-02]
		timeInFlowers_s := 257.4
		flowersNeededForPollenLoad := 58.0
		if fs.PollenMax_g > 0 {
			fillingLevel = state.Pollen_g / fs.PollenMax_g
		} else {
			fillingLevel = 0
		}
		if fillingLevel > 0 {
			handlingTime_s = fs.InterFlowerTime_s*flowersNeededForPollenLoad/fillingLevel + timeInFlowers_s
		} else {
			handlingTime_s = maxHandlingTime_s
		}
	}
	// [HAND-03]
	if handlingTime_s > maxHandlingTime_s {
		handlingTime_s = maxHandlingTime_s
	}
	return handlingTime_s
}

// log10 is NetLogo's "log x 10": ln x / ln 10 (math.Log10 can differ in the last digits).
func log10(x float64) float64 {
	return math.Log(x) / math.Log(10)
}

// Foraging_bestLayerREP returns the best layer for the bee e among the layers of the flower patch of
// myCurrentFoodsource, or the zero entity if none has the bee's forage [BLYR].
func (f *forager) Foraging_bestLayerREP(e ecs.Entity, myCurrentFoodsource ecs.Entity) ecs.Entity {
	_, b, _, _, fo, _ := f.beeMap.Get(e)
	memoBestHandlingtime := params.NotSetHigh
	memoBestEEF := params.NotSetLow
	memoBestPatch := ecs.Entity{}
	// [BLYR-01]
	colPos, _, _, _ := f.colonyMap.Get(b.ColonyID)
	fsPos, _, _, layer := f.foodsourceMap.Get(myCurrentFoodsource)
	distanceColonyFoodpatch_m := util.Distance(colPos.X, colPos.Y, fsPos.X, fsPos.Y) / f.landscape.Scaling_NLpatches_per_m
	sp := f.speciesMap.Get(b.SpeciesID)

	for _, currentLayer := range layer.LayersInPatchList {
		_, fs, state, _ := f.foodsourceMap.Get(currentLayer)
		// [BLYR-02]
		if fo.PollenForager && state.Pollen_g > 0 {
			handlingTime_s := f.HandlingTime_s_REP(e, currentLayer, fo.PollenForager)
			if handlingTime_s < memoBestHandlingtime {
				memoBestPatch = currentLayer
				memoBestHandlingtime = handlingTime_s
			}
		}
		// [BLYR-03] cost for the cohort, gain for one bee (no effect on the choice, see the annotation)
		if !fo.PollenForager && state.Nectar_myl > 0 {
			handlingTime_s := f.HandlingTime_s_REP(e, currentLayer, fo.PollenForager)
			energyCostsThisLayer_kJ := (float64(2*distanceColonyFoodpatch_m) +
				float64(handlingTime_s*sp.FlightVelocity_m_per_s*f.constants.EnergyFactorOnFlower)) *
				sp.FlightCosts_kJ_per_m_per_mg * b.Weight_mg * float64(b.Number)
			energyGainThisLayer_kJ := fo.Cropvolume_myl * f.constants.EnergySucrose_kJ_per_mymol * fs.NectarConcentration_mol_per_l
			eef := (energyGainThisLayer_kJ - energyCostsThisLayer_kJ) / energyCostsThisLayer_kJ
			if eef > memoBestEEF {
				memoBestPatch = currentLayer
				memoBestEEF = eef
			}
		}
	}
	return memoBestPatch
}

// Foraging_PatchChoiceProc lets the bee e choose the foodsource of her next trip of this forage [PCH].
func (f *forager) Foraging_PatchChoiceProc(e ecs.Entity, currentTripDuration_s float64) {
	_, b, _, _, fo, _ := f.beeMap.Get(e)
	// [PCH-01]
	preferenceClosePatchesProb := 0.9
	happy := true
	gotoNewLayergroup := false
	bestLayer := ecs.Entity{}
	myExpectation_s := fo.Expectation_NectarTrip_s
	myKnownMasterpatchesList := fo.KnownMasterpatchesNectarList
	searchProbBase := 0.1
	if fo.PollenForager {
		myExpectation_s = fo.Expectation_PollenTrip_s
		myKnownMasterpatchesList = fo.KnownMasterpatchesPollenList
	}
	myCurrentPatchPosition := -1.0

	if b.Activity == activity.ReturningEmpty {
		// [PCH-02]
		gotoNewLayergroup = true
	} else {
		// [PCH-03] Since step F5 the best layer always belongs to the known flower patch, so the masterpatch is
		// always found. NetLogo fails here if two flower patches lie in the same patch cell (finding 15, see
		// trans/netlogo-findings.md#finding-15).
		_, _, _, layer := f.foodsourceMap.Get(fo.CurrentFoodsource)
		myMasterpatchID := layer.MasterpatchID
		position := slices.Index(myKnownMasterpatchesList, myMasterpatchID)
		if position < 0 {
			panic("No number for myMasterpatchID (Foraging_PatchChoiceProc): the current foodsource belongs to no known flower patch")
		}
		myCurrentPatchPosition = float64(position)
		unhappyProb := (currentTripDuration_s - myExpectation_s) / currentTripDuration_s
		if unhappyProb < 0 {
			unhappyProb = 0
		}
		if f.rng.Float64() < searchProbBase+unhappyProb {
			happy = false
		}
		newExpectation := (myExpectation_s + currentTripDuration_s) / 2
		if fo.PollenForager {
			fo.Expectation_PollenTrip_s = newExpectation
		} else {
			fo.Expectation_NectarTrip_s = newExpectation
		}
	}

	// [PCH-04]
	if !happy {
		if fo.PollenForager {
			b.Activity = activity.ReturningUnhappyP
		} else {
			b.Activity = activity.ReturningUnhappyN
		}
		bestLayer = f.Foraging_bestLayerREP(e, fo.CurrentFoodsource)
		if bestLayer == fo.CurrentFoodsource {
			gotoNewLayergroup = true
		} else {
			fo.CurrentFoodsource = bestLayer
		}
	}

	// [PCH-05] random 0 is 0 (util.RandomInt)
	if gotoNewLayergroup {
		newPatchPosition := -999.0
		if f.rng.Float64() < preferenceClosePatchesProb {
			newPatchPosition = util.RandomInt(f.rng, myCurrentPatchPosition+1) - 1
		} else {
			newPatchPosition = myCurrentPatchPosition + 1
		}
		if newPatchPosition < 0 || newPatchPosition >= float64(len(myKnownMasterpatchesList)) {
			fo.CurrentFoodsource = ecs.Entity{}
		} else {
			fo.CurrentFoodsource = f.Foraging_bestLayerREP(e, myKnownMasterpatchesList[int(newPatchPosition)])
		}
	}
	if fo.PollenForager {
		fo.PollensourceToGoTo = fo.CurrentFoodsource
	} else {
		fo.NectarsourceToGoTo = fo.CurrentFoodsource
	}
}

// Foraging_costsAndChoiceProc applies the costs of the trip of the bee e in time, energy and mortality, then lets
// her choose her next foodsource [COST]. It returns false if she died on the trip.
//
// NetLogo: Foraging_costs&choiceProc ("&" is not allowed in Go names).
func (f *forager) Foraging_costsAndChoiceProc(w *ecs.World, e ecs.Entity) bool {
	_, b, _, _, fo, _ := f.beeMap.Get(e)
	persTimeSave := b.PersonalTime_s
	saveNumber := float64(b.Number)
	saveWeight_mg := b.Weight_mg
	saveFlightCosts_kJ_per_m_per_mg := f.speciesMap.Get(b.SpeciesID).FlightCosts_kJ_per_m_per_mg

	if b.Caste != caste.Worker && b.Caste != caste.Queen {
		panic(fmt.Sprintf("%v %v %d shouldn't be foraging (Foraging_costs&timeProc)", b.Stage, b.Caste, e.ID())) // Ark entity ID, not NetLogo's who (step F12)
	}
	if saveFlightCosts_kJ_per_m_per_mg < 0 {
		panic("saveFlightCosts local variable not set (Foraging_costs&timeProc)")
	}
	colPos, colony, stores, _ := f.colonyMap.Get(b.ColonyID)
	spColony := f.speciesMap.Get(colony.SpeciesIDcolony)
	// survival of one second, NetLogo: (1 - MortalityForager_per_s * ForagingMortalityFactor)
	survivalPerSecond := 1 - float64(f.derived.MortalityForager_per_s*f.file.ForagingMortalityFactor)

	// [COST-01]
	if b.Activity == activity.ReturningEmpty {
		stores.EnergyStore_kJ = stores.EnergyStore_kJ - float64(spColony.SearchLength_m*saveFlightCosts_kJ_per_m_per_mg*saveNumber*saveWeight_mg)
		tripDuration_s := spColony.SearchLength_m / spColony.FlightVelocity_m_per_s
		b.PersonalTime_s = b.PersonalTime_s + tripDuration_s
		survivalChance := math.Pow(survivalPerSecond, tripDuration_s)
		mortalityRisk := 1 - survivalChance
		if f.rng.Float64() < mortalityRisk {
			f.died.Emit(e) // output only, not in NetLogo
			f.death.DieProc(w, e, "foraging: empty")
			return false
		}
	}

	// [COST-02] the handling time uses the foodsource after the load was taken (finding 47, see
	// trans/netlogo-findings.md#finding-47)
	handlingTime_s := 0.0
	if b.Activity == activity.BringingNectar || b.Activity == activity.BringingPollen {
		if b.Activity == activity.BringingNectar {
			handlingTime_s = f.HandlingTime_s_REP(e, fo.CurrentFoodsource, fo.PollenForager)
			b.Activity = activity.ExpForagingN
		}
		if b.Activity == activity.BringingPollen {
			handlingTime_s = f.HandlingTime_s_REP(e, fo.CurrentFoodsource, fo.PollenForager)
			b.Activity = activity.ExpForagingP
		}
		fsPos, _, _, _ := f.foodsourceMap.Get(fo.CurrentFoodsource)
		distance := util.Distance(colPos.X, colPos.Y, fsPos.X, fsPos.Y)
		scaling := f.landscape.Scaling_NLpatches_per_m
		stores.EnergyStore_kJ = stores.EnergyStore_kJ - float64(
			(float64(2*distance/scaling)+
				float64(handlingTime_s*spColony.FlightVelocity_m_per_s*f.constants.EnergyFactorOnFlower))*
				saveFlightCosts_kJ_per_m_per_mg*saveWeight_mg*saveNumber)
		tripDuration_s := (2 * distance / scaling / spColony.FlightVelocity_m_per_s) + handlingTime_s
		b.PersonalTime_s = b.PersonalTime_s + tripDuration_s
		survivalChance := math.Pow(survivalPerSecond, tripDuration_s)
		mortalityRisk := 1 - survivalChance
		if f.rng.Float64() < mortalityRisk {
			f.died.Emit(e) // output only, not in NetLogo
			f.death.DieProc(w, e, "foraging: N or P forager")
			return false
		}
	}

	// [COST-03]
	if b.PersonalTime_s-persTimeSave <= 0 {
		panic("No time - or negative time - passed for this bee! (Foraging_costs&timeProc)")
	}
	f.Foraging_PatchChoiceProc(e, b.PersonalTime_s-persTimeSave)
	return true
}

// Foraging_unloadingProc puts the load of the bee e into her colony's stores [UNLD].
func (f *forager) Foraging_unloadingProc(e ecs.Entity) {
	_, b, _, _, fo, _ := f.beeMap.Get(e)
	// [UNLD-01]
	if b.Activity == activity.ExpForagingN || b.Activity == activity.ExpForagingP ||
		b.Activity == activity.ReturningUnhappyN || b.Activity == activity.ReturningUnhappyP {
		nectarIncrease := fo.NectarLoadSquadron_kJ
		pollenIncrease := fo.PollenLoadSquadron_g
		_, _, stores, _ := f.colonyMap.Get(b.ColonyID)
		stores.EnergyStore_kJ = stores.EnergyStore_kJ + nectarIncrease
		stores.PollenStore_g = stores.PollenStore_g + pollenIncrease
		fo.NectarLoadSquadron_kJ = 0
		fo.PollenLoadSquadron_g = 0
		b.PersonalTime_s = b.PersonalTime_s + f.speciesMap.Get(b.SpeciesID).TimeUnloading
	} else if fo.NectarLoadSquadron_kJ+fo.PollenLoadSquadron_g > 0 {
		panic("Bee did not unload nectar or pollen in Foraging_unloadingProc!")
	}
}
