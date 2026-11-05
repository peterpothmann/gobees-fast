package sys

import (
	"fmt"
	"math"

	"gobees/comp"
	"gobees/params"
)

// Threshold types and situations of ThresholdLevelREP [THR], as the NetLogo strings.
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

// thresholdTable holds the values of ThresholdLevelREP [THR]. NotSetHigh means: never.
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

// thresholdLevel returns the threshold of a task type in a situation [THR].
// An unknown type or situation is an assertion (NetLogo: "TH not set!").
func thresholdLevel(thType, situation string) float64 {
	th, ok := thresholdTable[thType][situation]
	if !ok {
		panic(fmt.Sprintf("Assertion violated in ThresholdLevelREP: TH not set! (%s, %s)", thType, situation))
	}
	return th
}

// thresholds returns all four thresholds of a situation, as set in [IQUE-05], [EMER-01] and [PSI-02].
func thresholds(situation string) comp.Thresholds {
	return comp.Thresholds{
		ThEgglaying:      thresholdLevel(thEggLaying, situation),
		ThForagingNectar: thresholdLevel(thNectarForaging, situation),
		ThForagingPollen: thresholdLevel(thPollenForaging, situation),
		ThNursing:        thresholdLevel(thNursing, situation),
	}
}

// proboscisLength returns the glossa length of a bee from her weight [mm] [PROB], finding 34
// (see trans/netlogo-findings.md#finding-34): linear between the minimum worker weight and the maximum queen weight.
func proboscisLength(weight_mg float64, sp *params.Species) float64 {
	minWeight_mg := sp.DevWeightPupationMin_mg
	maxWeight_mg := sp.DevWeight_Q_PupationMax_mg
	slope := (sp.Proboscis_max_mm - sp.Proboscis_min_mm) / (maxWeight_mg - minWeight_mg)
	length := sp.Proboscis_min_mm + (weight_mg-minWeight_mg)*slope
	if weight_mg < minWeight_mg || weight_mg > maxWeight_mg {
		panic(fmt.Sprintf("Wrong bee weight in ProboscisLengthREP: %v mg", weight_mg))
	}
	return length
}

// Factors of the crop volume and the pollen load [LOAD].
const (
	beeWeightToLoadFactor             = 0.402      // Ings et al. 2006, nectar loads
	beehaveCropToPelletFactor_ul_to_g = 0.015 / 50 // BEEHAVE: 50 microlitres nectar load, 0.015 g pollen load
)

// cropVolume returns the crop volume of a bee of weight weight_mg [µl] [LOAD]. A result of 0 (weight 0) is an
// assertion, as in NetLogo.
//
// NetLogo: CropAndPelletSizeREP "nectar". Since step F15 one function per forage, so the case of an unknown forage
// (an assertion in NetLogo) no longer exists.
func cropVolume(weight_mg float64, sp *params.Species) float64 {
	result := math.Min(sp.SpecMax_cropVolume_myl, weight_mg*beeWeightToLoadFactor)
	if result == 0 {
		panic("Assertion violated: Error in CropAndPelletSizeREP")
	}
	return result
}

// pollenLoad returns the pollen load of a bee of weight weight_mg [g] [LOAD]. A result of 0 (weight 0) is an
// assertion, as in NetLogo.
//
// NetLogo: CropAndPelletSizeREP "pollen".
func pollenLoad(weight_mg float64, sp *params.Species) float64 {
	result := math.Min(sp.SpecMax_pollenPellets_g, weight_mg*beeWeightToLoadFactor*beehaveCropToPelletFactor_ul_to_g)
	if result == 0 {
		panic("Assertion violated: Error in CropAndPelletSizeREP")
	}
	return result
}

// winterMortalityProb returns the probability of a queen to SURVIVE the winter [WMOR], finding 33
// (see trans/netlogo-findings.md#finding-33): 0.64 / (1 + e^(-22 * (relative weight - 0.32))).
func winterMortalityProb(weight_mg float64, sp *params.Species) float64 {
	minW := sp.DevWeight_Q_PupationMin_mg
	maxW := sp.DevWeight_Q_PupationMax_mg
	relWeight := (weight_mg - minW) / (maxW - minW)
	if relWeight < 0 || relWeight > 1 {
		panic(fmt.Sprintf("Wrong weight! (WintermortalityProbREP): %v mg", weight_mg))
	}
	return 0.64 / (1 + math.Exp(-22*(relWeight-0.32)))
}
