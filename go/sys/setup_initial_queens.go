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
	"gobees/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// initialQueensSetup creates the hibernating, mated queens of the first year [INIT-09].
//
// NetLogo: CreateInitialQueensProc.
type initialQueensSetup struct{}

func (s *initialQueensSetup) Initialize(w *ecs.World) {
	file := ecs.GetResource[params.ParameterFile](w)
	constants := ecs.GetResource[params.Constants](w)
	derived := ecs.GetResource[params.Derived](w)
	rng := rand.New(ecs.GetResource[resource.Rand](w))
	beeMapper := newBeeMapper(w)
	speciesByName := speciesIndices(w)
	speciesMap := newSpeciesTable(w)

	// [IQUE-01]
	for _, sq := range derived.BeeSpeciesInitialQueensList {
		if file.InitialQueensModifierBehaviorSpace == 0 {
			file.InitialQueensModifierBehaviorSpace = 1 // changes the parameter, as NetLogo changes the global
		}
		newQueens := int(util.Round(float64(sq.Queens) * file.InitialQueensModifierBehaviorSpace))

		// [IQUE-04] one-of the species with this name
		candidates := speciesByName[sq.Species]
		if len(candidates) == 0 {
			panic(fmt.Sprintf("species %q of BeeSpeciesInitialQueensList does not exist (see [SPEC-02])", sq.Species))
		}
		spIndex := candidates[rng.IntN(len(candidates))]
		sp := speciesMap.Get(spIndex)

		// [IQUE-02] ORDER: irrelevant, each queen only uses her own random draws.
		for range newQueens {
			bee := comp.Bee{
				// [IQUE-03]
				AdultAge: 180,
				BroodAge: 36,
				Caste:    caste.Queen,
				Mated:    true,
				Number:   1,
				// [IQUE-04]
				Species:  spIndex,
				Stage:    stage.Adult,
				Activity: activity.Hibernate,
				// [IQUE-06] -3599 to 3600 s, no effect (finding 31, see trans/netlogo-findings.md#finding-31)
				PersonalTime: util.RandomInt(rng, 2*3600) + (constants.GetUpTime_s - 3600),
			}
			genetics := comp.Genetics{
				MtDNA:       util.RandomFloat(rng, 139.9),
				Alleles:     []float64{util.RandomFloat(rng, 139.9), util.RandomFloat(rng, 139.9)},
				Spermatheca: []float64{util.RandomFloat(rng, 139.9)},
			}
			// [IQUE-07] emergingDate starts at 0, so it is drawn at least once. In the first year the absolute day
			// is the day of the year.
			emerging := day.Abs(0)
			for emerging <= 0 || emerging >= day.Abs(sp.SeasonStop) {
				emerging = day.Abs(util.Round(util.RandomNormal(rng, sp.EmergingDay_mean, sp.EmergingDay_sd)))
			}
			bee.EmergenceDate = day.At(emerging)
			// [IQUE-09]
			minWeight_mg := sp.DevWeight_Q_PupationMin_mg
			maxWeight_mg := sp.DevWeight_Q_PupationMax_mg
			bee.Weight = util.RandomNormal(rng, (maxWeight_mg+minWeight_mg)/2, (maxWeight_mg-minWeight_mg)/4)
			bee.Weight = min(max(bee.Weight, minWeight_mg), maxWeight_mg)

			// [IQUE-08] no foodsource known yet (zero entities, empty lists)
			forager := comp.Forager{
				GlossaLength:       proboscisLength(bee.Weight, sp),
				CropVolume:         cropVolume(bee.Weight, sp),
				PollenPellets:      pollenLoad(bee.Weight, sp),
				KnownNectarPatches: []ecs.Entity{},
				KnownPollenPatches: []ecs.Entity{},
			}
			th := thresholds(sitQueenInitiationPhase) // [IQUE-05]
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
