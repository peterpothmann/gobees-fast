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

## Step F15b: components, globals and enums

Fields of the Ark components (`go/comp`) and resources (`go/globals`). The fields stay exported (package `obs`
reads them). Fields lose a prefix that repeats their type (`Stim`, `Th`, `All`, `Total`, `Cumul`), slices lose the
suffix `List`. The doc comment of each field gives the unit and the NetLogo variable.

Unchanged: `Bee.Number`, `Stage`, `Caste`, `Activity`, `Mated`, `AdultAge`, `BroodAge`, `Genetics.MtDNA`,
`Forager.CurrentFoodsource`, `PollenForager`, the flags and dates of `Colony`, `ColonyCounts.LarvaWorkerRatio`,
`Foodsource.PatchType`, `StartDay`, `StopDay`, `Calendar.Day`, `Landscape.FoodsourcesFile` (follows the parameter
in F15c), `Landscape.Masterpatches`, `Ticks`, `ActiveBees`, `NestSites`.

### Bee (`go/comp/bee.go`)

| Type | gobees | NetLogo | gobees-fast | Unit |
| --- | --- | --- | --- | --- |
| `Bee` | `ColonyID` | `colonyID` | `Colony` |  |
| `Bee` | `SpeciesID` | `speciesID` | `Species` |  |
| `Bee` | `EmergingDate` | `emergingDate` | `EmergenceDate` |  |
| `Bee` | `PersonalTime_s` | `personalTime_s` | `PersonalTime` | [s] |
| `Bee` | `Weight_mg` | `weight_mg` | `Weight` | [mg] |
| `Genetics` | `AllelesList` | `allelesList` | `Alleles` |  |
| `Genetics` | `SpermathecaList` | `spermathecaList` | `Spermatheca` |  |
| `Thresholds` | `ThEgglaying` | `thEgglaying` | `Egglaying` |  |
| `Thresholds` | `ThForagingNectar` | `thForagingNectar` | `NectarForaging` |  |
| `Thresholds` | `ThForagingPollen` | `thForagingPollen` | `PollenForaging` |  |
| `Thresholds` | `ThNursing` | `thNursing` | `Nursing` |  |
| `Forager` | `Cropvolume_myl` | `cropvolume_myl` | `CropVolume` | [µl] |
| `Forager` | `PollenPellets_g` | `pollenPellets_g` | `PollenPellets` | [g] |
| `Forager` | `GlossaLength_mm` | `glossaLength_mm` | `GlossaLength` | [mm] |
| `Forager` | `NectarsourceToGoTo` | `nectarsourceToGoTo` | `NectarSource` |  |
| `Forager` | `PollensourceToGoTo` | `pollensourceToGoTo` | `PollenSource` |  |
| `Forager` | `KnownMasterpatchesNectarList` | `knownMasterpatchesNectarList` | `KnownNectarPatches` |  |
| `Forager` | `KnownMasterpatchesPollenList` | `knownMasterpatchesPollenList` | `KnownPollenPatches` |  |
| `Forager` | `NectarLoadSquadron_kJ` | `nectarLoadSquadron_kJ` | `CarriedNectar` | [kJ] |
| `Forager` | `PollenLoadSquadron_g` | `pollenLoadSquadron_g` | `CarriedPollen` | [g] |
| `Forager` | `Expectation_NectarTrip_s` | `expectation_NectarTrip_s` | `ExpectedNectarTrip` | [s] |
| `Forager` | `Expectation_PollenTrip_s` | `expectation_PollenTrip_s` | `ExpectedPollenTrip` | [s] |
| `Development` | `CumulIncubationReceived_kJ` | `cumulIncubationReceived_kJ` | `IncubationReceived` | [kJ] |
| `Development` | `CumulTimeLarva_d` | `cumulTimeLarva_d` | `LarvaTime` | [d] |

### Colony (`go/comp/colony.go`)

| Type | gobees | NetLogo | gobees-fast | Unit |
| --- | --- | --- | --- | --- |
| `Colony` | `SpeciesIDcolony` | `speciesIDcolony` | `Species` |  |
| `ColonyStores` | `EnergyStore_kJ` | `energyStore_kJ` | `Energy` | [kJ] |
| `ColonyStores` | `PollenStore_g` | `pollenStore_g` | `Pollen` | [g] |
| `ColonyStores` | `IdealEnergyStore_kJ` | `idealEnergyStore_kJ` | `IdealEnergy` | [kJ] |
| `ColonyStores` | `IdealPollenStore_g` | `idealPollenStore_g` | `IdealPollen` | [g] |
| `ColonyStores` | `EnergyNeedToday_kJ` | `energyNeedToday_kJ` | `EnergyNeedToday` | [kJ] |
| `ColonyStores` | `PollenNeedLarvaeToday_g` | `pollenNeedLarvaeToday_g` | `LarvaePollenNeedToday` | [g] |
| `ColonyStores` | `SummedIncubationToday_kJ` | `summedIncubationToday_kJ` | `IncubationToday` | [kJ] |
| `ColonyStimuli` | `StimEgglaying` | `stimEgglaying` | `Egglaying` |  |
| `ColonyStimuli` | `StimNectarForaging` | `stimNectarForaging` | `NectarForaging` |  |
| `ColonyStimuli` | `StimNursing` | `stimNursing` | `Nursing` |  |
| `ColonyStimuli` | `StimPollenForaging` | `stimPollenForaging` | `PollenForaging` |  |
| `ColonyCounts` | `AllEggs` | `allEggs` | `Eggs` |  |
| `ColonyCounts` | `AllLarvae` | `allLarvae` | `Larvae` |  |
| `ColonyCounts` | `AllPupae` | `allPupae` | `Pupae` |  |
| `ColonyCounts` | `AllAdults` | `allAdults` | `Adults` |  |
| `ColonyCounts` | `AllAdultWorkers` | `allAdultWorkers` | `AdultWorkers` |  |
| `ColonyRange` | `AllPatchesInRangeList` | `allPatchesInRangeList` | `Layers` |  |
| `ColonyRange` | `MasterpatchesWithPollenlayersInFlowerAndRangeList` | `masterpatchesWithPollenlayersInFlowerAndRangeList` | `PollenPatches` |  |
| `ColonyRange` | `MasterpatchesWithNectarlayersInFlowerAndRangeList` | `masterpatchesWithNectarlayersInFlowerAndRangeList` | `NectarPatches` |  |

### Foodsources (`go/comp/comp.go`)

| Type | gobees | NetLogo | gobees-fast | Unit |
| --- | --- | --- | --- | --- |
| `Foodsource` | `Id_Beescout` | `id_Beescout` | `ID` |  |
| `Foodsource` | `Area_sqm` | `area_sqm` | `Area` | [m2] |
| `Foodsource` | `Radius_m` | `radius_m` | `Radius` | [m] |
| `Foodsource` | `NectarMax_myl` | `nectarMax_myl` | `NectarMax` | [µl] |
| `Foodsource` | `PollenMax_g` | `pollenMax_g` | `PollenMax` | [g] |
| `Foodsource` | `NectarFlowerVolume_myl` | `nectarFlowerVolume_myl` | `NectarPerFlower` | [µl] |
| `Foodsource` | `NectarConcentration_mol_per_l` | `nectarConcentration_mol/l` | `NectarConcentration` | [mol/l] |
| `Foodsource` | `CorollaDepth_mm` | `corollaDepth_mm` | `CorollaDepth` | [mm] |
| `Foodsource` | `InterFlowerTime_s` | `interFlowerTime_s` | `InterFlowerTime` | [s] |
| `FoodsourceState` | `Nectar_myl` | `nectar_myl` | `Nectar` | [µl] |
| `FoodsourceState` | `Pollen_g` | `pollen_g` | `Pollen` | [g] |
| `Layer` | `Masterpatch` | `masterpatch?` | `IsMasterpatch` |  |
| `Layer` | `MasterpatchID` | `masterpatchID` | `Masterpatch` |  |
| `Layer` | `LayersInPatchList` | `layersInPatchList` | `PatchLayers` |  |

### Globals (`go/globals`)

| Type | gobees | NetLogo | gobees-fast | Unit |
| --- | --- | --- | --- | --- |
| `Daytime` | `Daytime_s` | `Daytime_s` | `Time` | [s] |
| `Landscape` | `Scaling_NLpatches_per_m` | `Scaling_NLpatches/m` | `Scale` | [1/m] |
| `SpeciesNestSites` | `NestHabitatsList` | `nestHabitatsList` | `Habitats` |  |
| `SpeciesNestSites` | `NestsiteFoodsourceList` | `nestsiteFoodsourceList` | `Sites` |  |
| `SpeciesNestSites` | `NestSiteArea` | `nestSiteArea` | `Area` | [m2] |
| `Calendar` | `DailyForagingPeriod_s` | `DailyForagingPeriod_s` | `ForagingPeriod` | [s] |
| `Flowering` | `FoodsourcesInFlowerUpdate` | `FoodsourcesInFlowerUpdate?` | `Changed` |  |
| `BroodTotals` | `TotalEggs` | `TotalEggs` | `Eggs` |  |
| `BroodTotals` | `TotalLarvae` | `TotalLarvae` | `Larvae` |  |
| `BroodTotals` | `TotalPupae` | `TotalPupae` | `Pupae` |  |

Types: `globals.FoodsourcesToday` → `globals.Flowering`, `globals.Totals` → `globals.BroodTotals`.

### Enum values (`go/enum/activity`)

`String` still returns the NetLogo strings.

| gobees | NetLogo | gobees-fast |
| --- | --- | --- |
| `ReturningUnhappyN` | "returningUnhappyN" | `ReturningUnhappyNectar` |
| `ReturningUnhappyP` | "returningUnhappyP" | `ReturningUnhappyPollen` |
| `ExpForagingN` | "expForagingN" | `ReturningHappyNectar` |
| `ExpForagingP` | "expForagingP" | `ReturningHappyPollen` |

### Observers (`go/obs`)

| gobees | NetLogo code | gobees-fast |
| --- | --- | --- |
| `(DailyOutput).TotalMapAreaREP` | [AREA] | `totalMapArea` |

The fields of `obs.DailyTotals` keep the names of the output columns.
