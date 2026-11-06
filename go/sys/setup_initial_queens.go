package sys

import (
	"fmt"
	"math/rand/v2"

	"gobees/comp"
	"gobees/day"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// initialQueensSetup creates the hibernating, mated queens of the first year.
//
// NetLogo: CreateInitialQueensProc.
type initialQueensSetup struct{}

func (s *initialQueensSetup) Initialize(w *ecs.World) {
	file := ecs.GetResource[params.Parameters](w)
	constants := ecs.GetResource[params.Constants](w)
	derived := ecs.GetResource[params.Derived](w)
	rng := rand.New(ecs.GetResource[resource.Rand](w))
	beeMapper := newBeeMapper(w)
	speciesByName := speciesIndices(w)
	speciesMap := newSpeciesTable(w)

	for _, sq := range derived.InitialQueens {
		if file.InitialQueensFactor == 0 {
			file.InitialQueensFactor = 1 // changes the parameter, as NetLogo changes the global
		}
		newQueens := int(round(float64(sq.Queens) * file.InitialQueensFactor))

		// one-of the species with this name
		candidates := speciesByName[sq.Species]
		if len(candidates) == 0 {
			panic(fmt.Sprintf("species %q of initialQueens does not exist", sq.Species))
		}
		spIndex := candidates[rng.IntN(len(candidates))]
		sp := speciesMap.Get(spIndex)

		// ORDER: irrelevant, each queen only uses her own random draws.
		for range newQueens {
			bee := comp.Bee{
				AdultAge: constants.InitialQueens.AdultAge,
				BroodAge: constants.InitialQueens.BroodAge,
				Caste:    caste.Queen,
				Mated:    true,
				Number:   1,
				Species:  spIndex,
				Stage:    stage.Adult,
				Activity: activity.Hibernate,
				// -3599 to 3600 s, no effect: the morning routine sets a new personal time for every
				// active bee, and hibernating queens are not active.
				PersonalTime: randomInt(rng, 2*3600) + (constants.Activities.GetUpTime - 3600),
			}
			genetics := comp.Genetics{
				MtDNA:       randomFloat(rng, constants.InitialQueens.AlleleRange),
				Alleles:     []float64{randomFloat(rng, constants.InitialQueens.AlleleRange), randomFloat(rng, constants.InitialQueens.AlleleRange)},
				Spermatheca: []float64{randomFloat(rng, constants.InitialQueens.AlleleRange)},
			}
			// emergingDate starts at 0, so it is drawn at least once. In the first year the absolute day
			// is the day of the year.
			emerging := day.Abs(0)
			for emerging <= 0 || emerging >= day.Abs(sp.SeasonStop) {
				emerging = day.Abs(round(randomNormal(rng, sp.EmergenceDayMean, sp.EmergenceDaySD)))
			}
			bee.EmergenceDate = day.At(emerging)
			minWeight := sp.QueenPupationWeightMin // [mg]
			maxWeight := sp.QueenPupationWeightMax // [mg]
			bee.Weight = randomNormal(rng, (maxWeight+minWeight)/2, (maxWeight-minWeight)/4)
			bee.Weight = min(max(bee.Weight, minWeight), maxWeight)

			// no foodsource known yet (zero entities, empty lists)
			forager := comp.Forager{
				GlossaLength:       proboscisLength(bee.Weight, sp),
				CropVolume:         cropVolume(bee.Weight, sp, &constants.Foraging),
				PollenPellets:      pollenLoad(bee.Weight, sp, &constants.Foraging),
				KnownNectarPatches: []ecs.Entity{},
				KnownPollenPatches: []ecs.Entity{},
			}
			th := thresholds(&constants.Thresholds, sitQueenInitiationPhase)
			beeMapper.NewEntity(&comp.Position{}, &bee, &genetics, &th, &forager, &comp.Development{})
		}
	}
}

func (s *initialQueensSetup) Update(w *ecs.World) {}

func (s *initialQueensSetup) Finalize(w *ecs.World) {}

// beeMapper creates bee entities with all their components.
type beeMapper = ecs.Map6[comp.Position, comp.Bee, comp.Genetics, comp.Thresholds, comp.Forager, comp.Development]

func newBeeMapper(w *ecs.World) *beeMapper {
	return ecs.NewMap6[comp.Position, comp.Bee, comp.Genetics, comp.Thresholds, comp.Forager, comp.Development](w)
}

// speciesIndices returns the indices of the species by name, in file order. NetLogo: Species with [ name = x ].
func speciesIndices(w *ecs.World) map[string][]int {
	result := map[string][]int{}
	for i, sp := range ecs.GetResource[params.Input](w).Species {
		result[sp.Name] = append(result[sp.Name], i)
	}
	return result
}
