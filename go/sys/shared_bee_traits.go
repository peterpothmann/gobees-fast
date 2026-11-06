package sys

import (
	"fmt"
	"math"

	"gobees/comp"
	"gobees/params"
)

// Threshold types and situations of thresholdLevel, as the NetLogo strings.
const (
	thEggLaying      = "eggLaying"
	thPollenForaging = "pollenForaging"
	thNectarForaging = "nectarForaging"
	thNursing        = "nursing"

	sitEgg                  = "egg"
	sitWorker               = "worker"
	sitYoungQueen           = "youngQueen"
	sitQueenInitiationPhase = "QueenInitiationPhase"
	sitQueenSocialPhase     = "QueenSocialPhase"
	sitPsith                = "Psith"
)

// thresholdTable holds the threshold of each type and situation. NotSetHigh means: never.
var thresholdTable = map[string]map[string]float64{
	thEggLaying: {
		sitEgg:                  params.NotSetHigh,
		sitWorker:               params.NotSetHigh,
		sitYoungQueen:           params.NotSetHigh,
		sitQueenInitiationPhase: 0.1,
		sitQueenSocialPhase:     0,
		sitPsith:                0.2,
	},
	thPollenForaging: {
		sitEgg:                  params.NotSetHigh,
		sitWorker:               0.9,
		sitYoungQueen:           params.NotSetHigh,
		sitQueenInitiationPhase: 0.7,
		sitQueenSocialPhase:     params.NotSetHigh,
		sitPsith:                params.NotSetHigh,
	},
	thNectarForaging: {
		sitEgg:                  params.NotSetHigh,
		sitWorker:               0.9,
		sitYoungQueen:           params.NotSetHigh,
		sitQueenInitiationPhase: 0.7,
		sitQueenSocialPhase:     params.NotSetHigh,
		sitPsith:                params.NotSetHigh,
	},
	thNursing: {
		sitEgg:                  params.NotSetHigh,
		sitWorker:               0.9,
		sitYoungQueen:           params.NotSetHigh,
		sitQueenInitiationPhase: 0.5,
		sitQueenSocialPhase:     0.9,
		sitPsith:                params.NotSetHigh,
	},
}

// thresholdLevel returns the threshold of a task type in a situation.
// An unknown type or situation is an assertion (NetLogo: "TH not set!").
func thresholdLevel(thType, situation string) float64 {
	th, ok := thresholdTable[thType][situation]
	if !ok {
		panic(fmt.Sprintf("Assertion violated in thresholdLevel: TH not set! (%s, %s)", thType, situation))
	}
	return th
}

// thresholds returns all four thresholds of a situation.
func thresholds(situation string) comp.Thresholds {
	return comp.Thresholds{
		Egglaying:      thresholdLevel(thEggLaying, situation),
		NectarForaging: thresholdLevel(thNectarForaging, situation),
		PollenForaging: thresholdLevel(thPollenForaging, situation),
		Nursing:        thresholdLevel(thNursing, situation),
	}
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

// Factors of the crop volume and the pollen load.
const (
	beeWeightToLoadFactor = 0.402      // Ings et al. 2006, nectar loads
	cropToPelletFactor    = 0.015 / 50 // [g/µl], BEEHAVE: 50 microlitres nectar load, 0.015 g pollen load
)

// cropVolume returns the crop volume of a bee of weight weight_mg [µl]. A result of 0 (weight 0) is an
// assertion, as in NetLogo. weight is in [mg].
//
// NetLogo: CropAndPelletSizeREP "nectar".
func cropVolume(weight float64, sp *params.Species) float64 {
	result := math.Min(sp.MaxCropVolume, weight*beeWeightToLoadFactor)
	if result == 0 {
		panic("Assertion violated: Error in cropVolume")
	}
	return result
}

// pollenLoad returns the pollen load of a bee of weight weight_mg [g]. A result of 0 (weight 0) is an
// assertion, as in NetLogo. weight is in [mg].
//
// NetLogo: CropAndPelletSizeREP "pollen".
func pollenLoad(weight float64, sp *params.Species) float64 {
	result := math.Min(sp.MaxPollenPellets, weight*beeWeightToLoadFactor*cropToPelletFactor)
	if result == 0 {
		panic("Assertion violated: Error in pollenLoad")
	}
	return result
}

// winterMortalityProb returns the probability of a queen to SURVIVE the winter (despite its name):
// 0.64 / (1 + e^(-22 * (relative weight - 0.32))), the same curve for all species.
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
