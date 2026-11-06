package sys

import (
	"fmt"
	"math"

	"gobees/comp"
	"gobees/params"
)

// Situations of thresholds, as the NetLogo strings.
const (
	sitEgg                  = "egg"
	sitWorker               = "worker"
	sitYoungQueen           = "youngQueen"
	sitQueenInitiationPhase = "QueenInitiationPhase"
	sitQueenSocialPhase     = "QueenSocialPhase"
	sitPsith                = "Psith"
)

// thresholds returns the four thresholds of a situation, from the values of the run file c. A task that is not
// possible in a situation has the threshold NotSetHigh, i.e. never. An unknown situation is an assertion (NetLogo:
// "TH not set!").
func thresholds(c *params.Thresholds, situation string) comp.Thresholds {
	never := params.NotSetHigh
	switch situation {
	case sitEgg, sitYoungQueen:
		return comp.Thresholds{Egglaying: never, NectarForaging: never, PollenForaging: never, Nursing: never}
	case sitWorker:
		return comp.Thresholds{Egglaying: never, NectarForaging: c.NectarForaging.Worker,
			PollenForaging: c.PollenForaging.Worker, Nursing: c.Nursing.Worker}
	case sitQueenInitiationPhase:
		return comp.Thresholds{Egglaying: c.EggLaying.QueenInitiationPhase,
			NectarForaging: c.NectarForaging.QueenInitiationPhase, PollenForaging: c.PollenForaging.QueenInitiationPhase,
			Nursing: c.Nursing.QueenInitiationPhase}
	case sitQueenSocialPhase:
		return comp.Thresholds{Egglaying: c.EggLaying.QueenSocialPhase, NectarForaging: never, PollenForaging: never,
			Nursing: c.Nursing.QueenSocialPhase}
	case sitPsith:
		return comp.Thresholds{Egglaying: c.EggLaying.Psithyrus, NectarForaging: never, PollenForaging: never,
			Nursing: never}
	}
	panic(fmt.Sprintf("Assertion violated in thresholds: TH not set! (%s)", situation))
}

// proboscisLength returns the glossa length of a bee from her weight [mm]: linear between the minimum worker weight
// and the maximum queen weight, the same line for workers, males and queens.
// weight is in [mg].
func proboscisLength(weight float64, sp *params.Species) float64 {
	minWeight := sp.PupationWeightMin      // [mg]
	maxWeight := sp.QueenPupationWeightMax // [mg]
	slope := (sp.ProboscisMax - sp.ProboscisMin) / (maxWeight - minWeight)
	length := sp.ProboscisMin + (weight-minWeight)*slope
	if weight < minWeight || weight > maxWeight {
		panic(fmt.Sprintf("Wrong bee weight in proboscisLength: %v mg", weight))
	}
	return length
}

// cropVolume returns the crop volume of a bee of weight weight_mg [µl]. A result of 0 (weight 0) is an
// assertion, as in NetLogo. weight is in [mg].
//
// NetLogo: CropAndPelletSizeREP "nectar".
func cropVolume(weight float64, sp *params.Species, c *params.ForagingConstants) float64 {
	result := math.Min(sp.MaxCropVolume, weight*c.CropFactor)
	if result == 0 {
		panic("Assertion violated: Error in cropVolume")
	}
	return result
}

// pollenLoad returns the pollen load of a bee of weight weight_mg [g]. A result of 0 (weight 0) is an
// assertion, as in NetLogo. weight is in [mg].
//
// NetLogo: CropAndPelletSizeREP "pollen".
func pollenLoad(weight float64, sp *params.Species, c *params.ForagingConstants) float64 {
	result := math.Min(sp.MaxPollenPellets, weight*c.CropFactor*c.PelletFactor)
	if result == 0 {
		panic("Assertion violated: Error in pollenLoad")
	}
	return result
}

// winterMortalityProb returns the probability of a queen to SURVIVE the winter (despite its name):
// 0.64 / (1 + e^(-22 * (relative weight - 0.32))), the same curve for all species (Beekman et al. 1998). The
// coefficients of the fitted curve stay in the code.
// weight is in [mg].
func winterMortalityProb(weight float64, sp *params.Species) float64 {
	minW := sp.QueenPupationWeightMin
	maxW := sp.QueenPupationWeightMax
	relWeight := (weight - minW) / (maxW - minW)
	if relWeight < 0 || relWeight > 1 {
		panic(fmt.Sprintf("Wrong weight! (winterMortalityProb): %v mg", weight))
	}
	return 0.64 / (1 + math.Exp(-22*(relWeight-0.32)))
}
