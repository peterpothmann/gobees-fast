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

	"github.com/mlange-42/ark/ecs"
)

// forager runs the foraging trips. activities.act calls it.
// It creates no entities. A forager that dies on her trip is removed by beeDeath.die.
//
// NetLogo: ForagingProc, Foraging_searchingProc, DetectionProbREP, Foraging_SortKnownPatchesListREP,
// Foraging_collectNectarPollenProc, HandlingTime_s_REP, Foraging_bestLayerREP, Foraging_PatchChoiceProc,
// Foraging_costs&choiceProc, Foraging_unloadingProc.
type forager struct {
	rng        *rand.Rand
	parameters *params.Parameters
	constants  *params.Constants
	derived    *params.Derived
	landscape  *globals.Landscape
	death      beeDeath
	trip       ecs.Event // event.ForagingTrip
	died       ecs.Event // event.DiedForaging

	beeMap        *ecs.Map2[comp.Bee, comp.Forager]
	colonyMap     *ecs.Map4[comp.Position, comp.Colony, comp.ColonyStores, comp.ColonyRange]
	foodsourceMap *ecs.Map4[comp.Position, comp.Foodsource, comp.FoodsourceState, comp.Layer]
	speciesMap    *speciesTable

	survivalPerSecond float64 // survival of one second of a trip, NaN until the first trip
	logSurvival       float64 // ln survivalPerSecond
}

func newForager(w *ecs.World, rng *rand.Rand) forager {
	return forager{
		rng:           rng,
		parameters:    ecs.GetResource[params.Parameters](w),
		constants:     ecs.GetResource[params.Constants](w),
		derived:       ecs.GetResource[params.Derived](w),
		landscape:     ecs.GetResource[globals.Landscape](w),
		death:         newBeeDeath(w),
		trip:          w.Event(event.ForagingTrip).For(ecs.C[comp.Bee]()),
		died:          w.Event(event.DiedForaging).For(ecs.C[comp.Bee]()),
		beeMap:        ecs.NewMap2[comp.Bee, comp.Forager](w),
		colonyMap:     ecs.NewMap4[comp.Position, comp.Colony, comp.ColonyStores, comp.ColonyRange](w),
		foodsourceMap: ecs.NewMap4[comp.Position, comp.Foodsource, comp.FoodsourceState, comp.Layer](w),
		speciesMap:    newSpeciesTable(w),

		survivalPerSecond: math.NaN(),
	}
}

// forage lets the bee e make one foraging trip. The bee may die on the trip.
func (f *forager) forage(w *ecs.World, e ecs.Entity) {
	b, fo := f.beeMap.Get(e)
	// output only, the observers count the trips
	f.trip.Emit(e)
	if b.Activity == activity.PollenForaging {
		fo.PollenForager = true
		fo.CurrentFoodsource = fo.PollenSource
	} else {
		fo.PollenForager = false
		fo.CurrentFoodsource = fo.NectarSource
	}
	if fo.CurrentFoodsource.IsZero() {
		b.Activity = activity.Searching
	} else if fo.PollenForager {
		b.Activity = activity.CollectPollen
	} else {
		b.Activity = activity.CollectNectar
	}
	f.search(e)
	f.collect(e)
	if !f.tripCosts(w, e) {
		return // died on the trip: the rest of her trip is skipped
	}
	f.unload(e)
	if fo.CarriedPollen+fo.CarriedNectar > 0 {
		panic("Bee carries pollen or nectar after unloading! (forage)")
	}
}

// sortKnownPatches returns the known masterpatches of the bee e without duplicates, sorted by
// distance to her colony.
func (f *forager) sortKnownPatches(e ecs.Entity, knownPatchesList []ecs.Entity) []ecs.Entity {
	b, _ := f.beeMap.Get(e)
	// remove-duplicates keeps the first occurrence, sort-by is stable
	newList := make([]ecs.Entity, 0, len(knownPatchesList))
	for _, p := range knownPatchesList {
		if !slices.Contains(newList, p) {
			newList = append(newList, p)
		}
	}
	colPos, _, _, _ := f.colonyMap.Get(b.Colony)
	distance := func(p ecs.Entity) float64 {
		pos, _, _, _ := f.foodsourceMap.Get(p)
		return distanceBetween(colPos.X, colPos.Y, pos.X, pos.Y)
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

// search lets a searching bee e try to find a flower patch.
func (f *forager) search(e ecs.Entity) {
	b, fo := f.beeMap.Get(e)
	if b.Activity == activity.Searching {
		chosenMasterpatch := ecs.Entity{}
		colPos, _, _, colRange := f.colonyMap.Get(b.Colony)
		xcol, ycol := colPos.X, colPos.Y
		var myMasterpatchesWithFoodList []ecs.Entity
		var detection []float64
		if fo.PollenForager {
			myMasterpatchesWithFoodList = colRange.PollenPatches
			detection = f.detections(&colRange.PollenDetection, colRange.PollenPatches, xcol, ycol)
		} else {
			myMasterpatchesWithFoodList = colRange.NectarPatches
			detection = f.detections(&colRange.NectarDetection, colRange.NectarPatches, xcol, ycol)
		}
		// ORDER: MATTERS. NetLogo draws once per candidate in random order and takes the last detected one, that is one
		// of the detected candidates, each with the same probability. Here: one draw per candidate in list order, and
		// the k-th detected candidate replaces the choice with probability 1/k (reservoir sampling), the same
		// distribution without a shuffle. Without either, the later candidates of the list would win.
		detected := 0
		for i, p := range myMasterpatchesWithFoodList {
			if f.rng.Float64() < detection[i] {
				detected++
				if detected == 1 || f.rng.IntN(detected) == 0 {
					chosenMasterpatch = p
				}
			}
		}
		if !chosenMasterpatch.IsZero() {
			fo.CurrentFoodsource = f.bestLayer(e, chosenMasterpatch)
			if !fo.CurrentFoodsource.IsZero() {
				if fo.PollenForager {
					fo.KnownPollenPatches = f.sortKnownPatches(e,
						append([]ecs.Entity{chosenMasterpatch}, fo.KnownPollenPatches...))
					b.Activity = activity.CollectPollen
				} else {
					fo.KnownNectarPatches = f.sortKnownPatches(e,
						append([]ecs.Entity{chosenMasterpatch}, fo.KnownNectarPatches...))
					b.Activity = activity.CollectNectar
				}
			} else {
				b.Activity = activity.ReturningEmpty // the patch was emptied earlier today
			}
		} else {
			b.Activity = activity.ReturningEmpty
		}
	}
	if b.Activity == activity.Searching {
		panic("Bee is still searching! (search)")
	}
}

// detectionProb returns the probability that a bee of the colony at (xcol, ycol) detects the masterpatch
// patchWho on a search flight.
// detections returns the detection probability of each masterpatch of patches from a colony at (xcol, ycol), from
// cache if it holds one per patch, else computed and stored in cache. The caller resets cache to nil when it changes
// the list (comp.ColonyRange).
func (f *forager) detections(cache *[]float64, patches []ecs.Entity, xcol, ycol float64) []float64 {
	if *cache == nil || len(*cache) != len(patches) {
		probs := make([]float64, len(patches))
		for i, p := range patches {
			probs[i] = f.detectionProb(p, xcol, ycol)
		}
		*cache = probs
	}
	return *cache
}

func (f *forager) detectionProb(patchWho ecs.Entity, xcol, ycol float64) float64 {
	pos, fs, _, _ := f.foodsourceMap.Get(patchWho)
	dist := distanceBetween(pos.X, pos.Y, xcol, ycol) / f.landscape.Scale // [m]
	patchRadius := fs.Radius                                              // [m]
	relevantDistance := dist - patchRadius                                // [m]
	if relevantDistance < 0 {
		relevantDistance = 0
	}
	detProb := math.Pow(math.E, f.parameters.DetectionLambda*relevantDistance) // NetLogo: e ^ x
	// no cut-off at ForagingRangeMax_m as in NetLogo: the colony's lists hold only masterpatches in range
	return precision(detProb, 10)
}

// collect lets the bee e take a load from her foodsource.
func (f *forager) collect(e ecs.Entity) {
	b, fo := f.beeMap.Get(e)
	if b.Activity == activity.CollectPollen {
		_, _, state, _ := f.foodsourceMap.Get(fo.CurrentFoodsource)
		if state.Pollen > 0 {
			fo.CarriedPollen = min(fo.PollenPellets*float64(b.Number), state.Pollen)
			b.Activity = activity.BringingPollen
			memoPollenLoad := fo.CarriedPollen
			state.Pollen = state.Pollen - memoPollenLoad
		} else {
			b.Activity = activity.ReturningEmpty
		}
	}
	if b.Activity == activity.CollectNectar {
		_, fs, state, _ := f.foodsourceMap.Get(fo.CurrentFoodsource)
		if state.Nectar > 0 {
			nectarRemoved := min(fo.CropVolume*float64(b.Number), state.Nectar)
			fo.CarriedNectar = nectarRemoved * f.constants.Foraging.SucroseEnergy * fs.NectarConcentration
			b.Activity = activity.BringingNectar
			state.Nectar = state.Nectar - nectarRemoved
		} else {
			b.Activity = activity.ReturningEmpty
		}
	}
}

// handlingTime returns the time [s] the bee e needs to collect one full load at the foodsource myPatch.
// pollenPatch selects pollen or nectar.
func (f *forager) handlingTime(e ecs.Entity, myPatch ecs.Entity, pollenPatch bool) float64 {
	b, fo := f.beeMap.Get(e)
	_, fs, state, _ := f.foodsourceMap.Get(myPatch)
	maxHandlingTime := f.constants.Foraging.MaxHandlingTime // [s]
	handlingTime := -999.0                                  // [s]
	fillingLevel := 0.0

	if !pollenPatch {
		if fs.NectarMax > 0 {
			fillingLevel = state.Nectar / fs.NectarMax
		}
		beeWeight := b.Weight / 1000       // W [g]
		glossaLength := fo.GlossaLength    // G [mm]
		corollaDepth := fs.CorollaDepth    // C [mm]
		nectarVolume := fs.NectarPerFlower // V [µl]
		// The coefficients of the fitted equations stay in the code.
		// float64(...): no fused multiply-add, so the result equals NetLogo (Java)
		accessTime := 0.3 + float64(0.04*corollaDepth) // Ta [s]
		// the same at every visit of this bee to this foodsource: computed again only for another foodsource
		// (comp.Forager.IngestionTime)
		if fo.IngestionSource != myPatch {
			numerator := log10(nectarVolume + 1)
			// the same for every flower the bee visits: computed once per adult bee (comp.Forager.IngestionFactor)
			if fo.IngestionFactor == 0 {
				fo.IngestionFactor = 0.3 * math.Pow(beeWeight, 0.3333) * glossaLength
			}
			num := fo.IngestionFactor
			base := 1.41 - corollaDepth/glossaLength
			if base < 0.001 {
				base = 0.001
			}
			den := powMinus04(base) - float64(0.3*accessTime)
			denominator := log10(num/den + 1)
			fo.IngestionSource, fo.IngestionTime = myPatch, numerator/denominator
		}
		ingestionTime := fo.IngestionTime // Ti [s]

		handlingTimePerFlower := maxHandlingTime // [s]
		if fillingLevel > 0 {
			handlingTimePerFlower = (fs.InterFlowerTime+accessTime)/fillingLevel + ingestionTime
		}
		flowersVisited := 1.0
		if nectarVolume < fo.CropVolume {
			flowersVisited = fo.CropVolume / nectarVolume
		}
		handlingTime = handlingTimePerFlower * flowersVisited
	} else {
		timeInFlowers := f.constants.Foraging.PollenLoadTime // [s]
		flowersNeededForPollenLoad := f.constants.Foraging.FlowersPerPollenLoad
		if fs.PollenMax > 0 {
			fillingLevel = state.Pollen / fs.PollenMax
		} else {
			fillingLevel = 0
		}
		if fillingLevel > 0 {
			handlingTime = fs.InterFlowerTime*flowersNeededForPollenLoad/fillingLevel + timeInFlowers
		} else {
			handlingTime = maxHandlingTime
		}
	}
	if handlingTime > maxHandlingTime {
		handlingTime = maxHandlingTime
	}
	return handlingTime
}

// log10 is NetLogo's "log x 10": ln x / ln 10 (math.Log10 can differ in the last digits).
func log10(x float64) float64 {
	return math.Log(x) / ln10
}

// ln10 is math.Log(10), computed once: the compiler does not fold math.Log of a constant.
var ln10 = math.Log(10)

// powMinus04 is math.Pow(x, -0.4) bit for bit for 0 < x < +Inf, without the special cases and the integer part
// of math.Pow. For a power y with 0 < |y| <= 0.5, math.Pow computes Exp(|y|·Log x) and takes 1/· for y < 0.
// TestPowMinus04 checks the equality, also for a new Go version.
func powMinus04(x float64) float64 {
	return 1 / math.Exp(0.4*math.Log(x))
}

// bestLayer returns the best layer for the bee e among the layers of the flower patch of
// myCurrentFoodsource, or the zero entity if none has the bee's forage.
func (f *forager) bestLayer(e ecs.Entity, myCurrentFoodsource ecs.Entity) ecs.Entity {
	b, fo := f.beeMap.Get(e)
	memoBestHandlingtime := params.NotSetHigh
	memoBestEEF := params.NotSetLow
	memoBestPatch := ecs.Entity{}
	colPos, _, _, _ := f.colonyMap.Get(b.Colony)
	fsPos, _, _, layer := f.foodsourceMap.Get(myCurrentFoodsource)
	patchDistance := distanceBetween(colPos.X, colPos.Y, fsPos.X, fsPos.Y) / f.landscape.Scale // [m]
	sp := f.speciesMap.Get(b.Species)

	for _, currentLayer := range layer.PatchLayers {
		_, fs, state, _ := f.foodsourceMap.Get(currentLayer)
		if fo.PollenForager && state.Pollen > 0 {
			handlingTime := f.handlingTime(e, currentLayer, fo.PollenForager) // [s]
			if handlingTime < memoBestHandlingtime {
				memoBestPatch = currentLayer
				memoBestHandlingtime = handlingTime
			}
		}
		// cost for the cohort, gain for one bee. No effect on the choice: eef + 1 = gain / cost, and the number
		// divides all candidates alike. SIMPLIFY: cost for one bee (may change ties at the last bit).
		if !fo.PollenForager && state.Nectar > 0 {
			handlingTime := f.handlingTime(e, currentLayer, fo.PollenForager) // [s]
			energyCosts := (float64(2*patchDistance) +
				float64(handlingTime*sp.FlightVelocity*f.constants.Foraging.EnergyFactorOnFlower)) *
				sp.FlightCosts * b.Weight * float64(b.Number) // [kJ]
			energyGain := fo.CropVolume * f.constants.Foraging.SucroseEnergy * fs.NectarConcentration // [kJ]
			eef := (energyGain - energyCosts) / energyCosts
			if eef > memoBestEEF {
				memoBestPatch = currentLayer
				memoBestEEF = eef
			}
		}
	}
	return memoBestPatch
}

// choosePatch lets the bee e choose the foodsource of her next trip of this forage.
// tripDuration is the duration of the trip just made [s].
func (f *forager) choosePatch(e ecs.Entity, tripDuration float64) {
	b, fo := f.beeMap.Get(e)
	preferenceClosePatchesProb := f.constants.Foraging.PreferCloseProb
	happy := true
	gotoNewLayergroup := false
	bestLayer := ecs.Entity{}
	myExpectation := fo.ExpectedNectarTrip // [s]
	myKnownMasterpatchesList := fo.KnownNectarPatches
	searchProbBase := f.constants.Foraging.SearchProb
	if fo.PollenForager {
		myExpectation = fo.ExpectedPollenTrip
		myKnownMasterpatchesList = fo.KnownPollenPatches
	}
	myCurrentPatchPosition := -1.0

	if b.Activity == activity.ReturningEmpty {
		gotoNewLayergroup = true
	} else {
		// the best layer always belongs to the known flower patch, so the masterpatch is always found.
		// NetLogo fails here if two flower patches lie in the same patch cell.
		_, _, _, layer := f.foodsourceMap.Get(fo.CurrentFoodsource)
		myMasterpatchID := layer.Masterpatch
		position := slices.Index(myKnownMasterpatchesList, myMasterpatchID)
		if position < 0 {
			panic("No number for myMasterpatchID (choosePatch): the current foodsource belongs to no known flower patch")
		}
		myCurrentPatchPosition = float64(position)
		unhappyProb := (tripDuration - myExpectation) / tripDuration
		if unhappyProb < 0 {
			unhappyProb = 0
		}
		if f.rng.Float64() < searchProbBase+unhappyProb {
			happy = false
		}
		newExpectation := (myExpectation + tripDuration) / 2
		if fo.PollenForager {
			fo.ExpectedPollenTrip = newExpectation
		} else {
			fo.ExpectedNectarTrip = newExpectation
		}
	}

	if !happy {
		if fo.PollenForager {
			b.Activity = activity.ReturningUnhappyPollen
		} else {
			b.Activity = activity.ReturningUnhappyNectar
		}
		bestLayer = f.bestLayer(e, fo.CurrentFoodsource)
		if bestLayer == fo.CurrentFoodsource {
			gotoNewLayergroup = true
		} else {
			fo.CurrentFoodsource = bestLayer
		}
	}

	// random 0 is 0 (randomInt)
	if gotoNewLayergroup {
		newPatchPosition := -999.0
		if f.rng.Float64() < preferenceClosePatchesProb {
			newPatchPosition = randomInt(f.rng, myCurrentPatchPosition+1) - 1
		} else {
			newPatchPosition = myCurrentPatchPosition + 1
		}
		if newPatchPosition < 0 || newPatchPosition >= float64(len(myKnownMasterpatchesList)) {
			fo.CurrentFoodsource = ecs.Entity{}
		} else {
			fo.CurrentFoodsource = f.bestLayer(e, myKnownMasterpatchesList[int(newPatchPosition)])
		}
	}
	if fo.PollenForager {
		fo.PollenSource = fo.CurrentFoodsource
	} else {
		fo.NectarSource = fo.CurrentFoodsource
	}
}

// tripCosts applies the costs of the trip of the bee e in time, energy and mortality, then lets
// her choose her next foodsource. It returns false if she died on the trip.
//
// NetLogo: Foraging_costs&choiceProc.
func (f *forager) tripCosts(w *ecs.World, e ecs.Entity) bool {
	b, fo := f.beeMap.Get(e)
	persTimeSave := b.PersonalTime
	saveNumber := float64(b.Number)
	saveWeight := b.Weight                                     // [mg]
	saveFlightCosts := f.speciesMap.Get(b.Species).FlightCosts // [kJ/(m mg)]

	if b.Caste != caste.Worker && b.Caste != caste.Queen {
		panic(fmt.Sprintf("%v %v %d shouldn't be foraging (tripCosts)", b.Stage, b.Caste, e.ID())) // Ark entity ID, not NetLogo's who
	}
	if saveFlightCosts < 0 {
		panic("saveFlightCosts local variable not set (tripCosts)")
	}
	colPos, colony, stores, _ := f.colonyMap.Get(b.Colony)
	spColony := f.speciesMap.Get(colony.Species)
	// survival of one second, NetLogo: (1 - MortalityForager_per_s * ForagingMortalityFactor). A trip of t seconds
	// survives with survivalPerSecond^t = exp(t ln survivalPerSecond), the logarithm is computed again only when the
	// parameters change (tests). NetLogo computes the power: other last bits, the same distribution (docs/changes.md).
	if s := 1 - float64(f.derived.ForagerMortality*f.parameters.ForagingMortalityFactor); s != f.survivalPerSecond {
		f.survivalPerSecond, f.logSurvival = s, math.Log(s)
	}

	if b.Activity == activity.ReturningEmpty {
		stores.Energy = stores.Energy - float64(spColony.SearchLength*saveFlightCosts*saveNumber*saveWeight)
		tripDuration := spColony.SearchLength / spColony.FlightVelocity // [s]
		b.PersonalTime = b.PersonalTime + tripDuration
		survivalChance := math.Exp(tripDuration * f.logSurvival)
		mortalityRisk := 1 - survivalChance
		if f.rng.Float64() < mortalityRisk {
			f.died.Emit(e) // output only, not in NetLogo
			f.death.die(w, e, "foraging: empty")
			return false
		}
	}

	// the handling time uses the foodsource after the load was taken, as in NetLogo: if the load emptied it, the
	// handling time is the maximum of 3600 s
	handlingTime := 0.0 // [s]
	if b.Activity == activity.BringingNectar || b.Activity == activity.BringingPollen {
		if b.Activity == activity.BringingNectar {
			handlingTime = f.handlingTime(e, fo.CurrentFoodsource, fo.PollenForager)
			b.Activity = activity.ReturningHappyNectar
		}
		if b.Activity == activity.BringingPollen {
			handlingTime = f.handlingTime(e, fo.CurrentFoodsource, fo.PollenForager)
			b.Activity = activity.ReturningHappyPollen
		}
		fsPos, _, _, _ := f.foodsourceMap.Get(fo.CurrentFoodsource)
		distance := distanceBetween(colPos.X, colPos.Y, fsPos.X, fsPos.Y)
		scaling := f.landscape.Scale
		stores.Energy = stores.Energy - float64(
			(float64(2*distance/scaling)+
				float64(handlingTime*spColony.FlightVelocity*f.constants.Foraging.EnergyFactorOnFlower))*
				saveFlightCosts*saveWeight*saveNumber)
		tripDuration := (2 * distance / scaling / spColony.FlightVelocity) + handlingTime // [s]
		b.PersonalTime = b.PersonalTime + tripDuration
		survivalChance := math.Exp(tripDuration * f.logSurvival)
		mortalityRisk := 1 - survivalChance
		if f.rng.Float64() < mortalityRisk {
			f.died.Emit(e) // output only, not in NetLogo
			f.death.die(w, e, "foraging: N or P forager")
			return false
		}
	}

	if b.PersonalTime-persTimeSave <= 0 {
		panic("No time - or negative time - passed for this bee! (tripCosts)")
	}
	f.choosePatch(e, b.PersonalTime-persTimeSave)
	return true
}

// unload puts the load of the bee e into her colony's stores.
func (f *forager) unload(e ecs.Entity) {
	b, fo := f.beeMap.Get(e)
	if b.Activity == activity.ReturningHappyNectar || b.Activity == activity.ReturningHappyPollen ||
		b.Activity == activity.ReturningUnhappyNectar || b.Activity == activity.ReturningUnhappyPollen {
		nectarIncrease := fo.CarriedNectar
		pollenIncrease := fo.CarriedPollen
		_, _, stores, _ := f.colonyMap.Get(b.Colony)
		stores.Energy = stores.Energy + nectarIncrease
		stores.Pollen = stores.Pollen + pollenIncrease
		fo.CarriedNectar = 0
		fo.CarriedPollen = 0
		b.PersonalTime = b.PersonalTime + f.speciesMap.Get(b.Species).UnloadingTime
	} else if fo.CarriedNectar+fo.CarriedPollen > 0 {
		panic("Bee did not unload nectar or pollen in unload!")
	}
}
