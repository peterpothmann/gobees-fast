# Names in gobees-fast, gobees and NetLogo

gobees-fast renames the types, functions and fields that gobees named after BEE-STEWARD (NetLogo), so that the code
reads without the NetLogo code at hand. This table is the link between the three: the name in gobees (and, with the
suffixes `Proc` and `REP`, mostly in NetLogo), the code of the annotation in the NetLogo model
(`netlogo/__BEE-STEWARD_RC_ColourVersionMB.nlogo`) and the name in gobees-fast.

Rules:

- No NetLogo suffixes (`Proc`, `REP`) or prefixes (`Foraging_`, `Development_`), Go spelling.
- Unexported unless another package uses it.
- No unit in the name. The unit stands in the doc comment, e.g. `// handling time of one load [s]`.
- The output columns keep their NetLogo names, so that the comparison with NetLogo can match them.

## Step F15a: parts, functions and methods in `go/sys`

### Parts

The parts are the Ark systems that the groups `Setup`, `Morning`, `Day` and `Evening` (`go/sys/schedule.go`) call in
order. A type and a function cannot share a name in one package, so the parts are nouns and the functions verbs.

| gobees | NetLogo code | gobees-fast | File (gobees-fast) |
| --- | --- | --- | --- |
| `CreateFoodsources` | [INIT-06] | `landscapeSetup` | `setup_landscape.go` |
| `CreateSpecies` | [INIT-07] | `nestSitesSetup` | `setup_nest_sites.go` |
| `CreateBadgers` | [INIT-08] | `badgersSetup` | `setup_badgers.go` |
| `CreateInitialQueens` | [INIT-09] | `initialQueensSetup` | `setup_initial_queens.go` |
| `Tick` | [GO-03] | `tick` | `morning_tick.go` |
| `UpdateMorning` | [MORN-01] to [MORN-04] | `calendarUpdate` | `morning_calendar.go` |
| `UpdateFoodsources` | [MORN-05] | `flowerRefill` | `morning_flower_refill.go` |
| `UpdateSeasonalEvents` | [MORN-06] | `seasonalEvents` | `morning_seasonal_events.go` |
| `EmergenceNewQueens` | [MORN-07] | `queenEmergence` | `morning_queen_emergence.go` |
| `WakeUpBees` | [MORN-08] | `wakeUp` | `morning_wake_up.go` |
| `CropRotation` | [MORN-09] | `cropRotation` | `morning_crop_rotation.go` |
| `UpdateColonies` | [MORN-10] | `colonyUpdate` | `morning_colony_update.go` |
| `CheckNumbers` | [MORN-12] | `broodTotals` | `morning_brood_totals.go` |
| `NeedNectarPollenLarvaeToday` | [GO-05] | `larvaeNeeds` | `morning_larvae_needs.go` |
| `ActivityLoop` | [GO-06] to [GO-10] | `activities` | `day_activities.go` |
| `EndOfDayResting` | [GO-11] | `nightRest` | `day_night_rest.go` |
| `QueensLeavingNest` | [GO-12] | `queensLeavingNest` | `evening_queens_leaving_nest.go` |
| `FeedLarvae` | [GO-13] | `larvaeFeeding` | `evening_larvae_feeding.go` |
| `QueenProductionDate` | [GO-14] | `queenProductionDate` | `evening_queen_production_date.go` |
| `Development` | [GO-15] | `development` | `evening_development.go` |
| `MortalityBrood` | [GO-16] | `broodMortality` | `evening_brood_mortality.go` |
| `BadgersOnTheProwl` | [GO-17] | `badgers` | `evening_badgers.go` |
| `Extinction` | [GO-20] | `extinction` | `evening_extinction.go` |

### Functions and methods

| gobees (NetLogo) | NetLogo code | gobees-fast |
| --- | --- | --- |
| `CreateFoodsourcesProc` | [FSRC] | `createFoodsources` |
| `CreateLayersProc` | [LAYR] | `createLayers` |
| `MergeHedgesProc` | [HEDG] | `mergeHedges` |
| `NestHabitatsListREP` | [NHAB] | `expandNestHabitats` |
| `CreateBadgersProc` | [BADC] | `createBadgers` |
| `CropRotationProc` | [CROP] | `cropRotation.rotate` |
| `CreateColoniesProc` | [CCOL] | `colonyFounder.found` |
| `PatchesInRangeProc` | [PRNG] | `colonyFounder.patchesInRange` |
| `FoodsourcesInFlowerAndRangeProc` | [FRNG] | `colonyFounder.flowersInRange` |
| `NestSitesSearchingProc` | [NEST] | `queenEmergence.searchNestSite` |
| `NestSiteFoodSourceREP` | [NSFS] | `queenEmergence.chooseNestSite` |
| `PsithyrusNestSearchProc` | [PSI] | `queenEmergence.searchHostColony` |
| `DieProc` | [DIE-01] to [DIE-03] | `beeDeath.die`, `colonyDeath.die`, `badgerDeath.die` |
| `ActivityProc` | [ACT] | `activities.act` |
| `NextActiveBeeREP` | [NEXT] | `activities.nextActiveBee` |
| `BroodIncubationProc` | [INC] | `activities.incubateBrood` |
| `EgglayingProc` | [EGG] | `eggLayer.layEggs` |
| `EggsParameterSettingProc` | [EGGP] | `eggLayer.initEggs` |
| `ForagingProc` | [FOR] | `forager.forage` |
| `Foraging_searchingProc` | [SRCH] | `forager.search` |
| `Foraging_collectNectarPollenProc` | [COLL] | `forager.collect` |
| `Foraging_costs&choiceProc` (gobees: `Foraging_costsAndChoiceProc`) | [COST] | `forager.tripCosts` |
| `Foraging_PatchChoiceProc` | [PCH] | `forager.choosePatch` |
| `Foraging_SortKnownPatchesListREP` | [SORT] | `forager.sortKnownPatches` |
| `Foraging_bestLayerREP` | [BLYR] | `forager.bestLayer` |
| `Foraging_unloadingProc` | [UNLD] | `forager.unload` |
| `HandlingTime_s_REP` | [HAND] | `forager.handlingTime` |
| `DetectionProbREP` | [DET] | `forager.detectionProb` |
| `Development_EggsProc` | [DEVE] | `development.developEgg` |
| `Development_LarvaeProc` | [DEVL] | `development.developLarva` |
| `Development_PupaeProc` | [DEVP] | `development.developPupa` |
| `Development_Mortality_AdultsProc` | [MORA] | `development.ageAdult` |
| `DetermineCaste_REP` | [CAST] | `determineCaste` |
| `MaxWeightGainToday_mg_REP` | [WGAIN] | `maxWeightGainToday` |
| `CompetitionPointDateREP` | [CPD] | `competitionPointDate` |
| `CropAndPelletSizeREP` "nectar" | [LOAD] | `cropVolume` |
| `CropAndPelletSizeREP` "pollen" | [LOAD] | `pollenLoad` |
| `ProboscisLengthREP` | [PROB] | `proboscisLength` |
| `ThresholdLevelREP` | [THR] | `thresholdLevel` |
| `WintermortalityProbREP` | [WMOR] | `winterMortalityProb` |
| `StimEgglayingREP` | [STEG] | `egglayingStimulus` |
| `StimNursingREP` | [STNU] | `nursingStimulus` |
| `StimForagingNectarREP` | [STNE] | `nectarStimulus` |
| `StimForagingPollenREP` | [STPO] | `pollenStimulus` |
