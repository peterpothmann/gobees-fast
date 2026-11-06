# Names in gobees-fast, gobees and NetLogo

gobees-fast renames the types, functions and fields that gobees named after BEE-STEWARD (NetLogo), so that the code
reads without the NetLogo code at hand. This table is the link between the three: the name in gobees (and, with the
suffixes `Proc` and `REP`, mostly in NetLogo), the code of the annotation in the NetLogo model
(`netlogo/__BEE-STEWARD_RC_ColourVersionMB.nlogo`) and the name in gobees-fast.

Rules:

- No NetLogo suffixes (`Proc`, `REP`) or prefixes (`Foraging_`, `Development_`), Go spelling.
- Unexported unless another package uses it.
- No unit in the name. The unit stands in the doc comment, e.g. `// handling time of one load [s]`.
- The keys of the run files are the Go names in lowerCamelCase (step F15c), their units are in the F15c tables.
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
`Foodsource.PatchType`, `StartDay`, `StopDay`, `Calendar.Day`, `Landscape.FoodsourcesFile` (`File` since
F15c), `Landscape.Masterpatches`, `Ticks`, `ActiveBees`, `NestSites`.

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

## Step F15c: parameters and the keys of the run files

All keys of the run file and the landscape file are the Go names in lowerCamelCase, without unit and `?`. The
Go types `params.Parameters` and `params.SpeciesFile` carry the NetLogo column in the tag `netlogo` where it
differs from the key: the converter `tools/netlogo2gobees` finds the columns by it. The tables are also the
reference of the keys: the unit of a value stands only here and in the doc comment of the Go field.

Types: `params.ParameterFile` → `params.Parameters` (field `Params.File` → `Params.Parameters`, the fields `file`
of the systems → `parameters`), `params.Weather` → `params.ForagingPeriod` (method `ForagingPeriod_s` → `Get`,
field `Years_s` → `Years`, `Input.Weather` → `Input.ForagingPeriod`).

### Sections of the run file

| gobees Go | NetLogo | gobees-fast Go (`input.Config`) | Key | Unit | Content |
| --- | --- | --- | --- | --- | --- |
| `Parameters` | the columns of the parameter file | `Parameters` | `parameters` |  | the model parameters |
| `Output` | `KeepDeadColonies?` | `Output` | `output` |  | output options |
| `InitialQueens` | `BeeSpeciesInitialQueensListAsString` | `InitialQueens` | `initialQueens` |  | species (`species`) and number of initial queens (`queens`) |
| `ForagingPeriod_s` | `Weather` | `ForagingPeriod` | `foragingPeriod` | [s] | 365 values per year, the years repeat |
| `StewardshipOptions` | from the `HabitatsFile` | `StewardshipOptions` | `stewardshipOptions` |  | names of the stewardship options |
| `Species` | rows of the `BeespeciesFile` | `Species` | `species` |  | bee species |
| `Flowers` | rows of the `FlowerspeciesFile` | `Flowers` | `flowers` |  | flower species |
| `Landscape` | foodsources file of `MyMap` | `Landscape` | `landscape` |  | landscape file of the start |
| `CropRotation` | `CropRotationListAsString` | `CropRotation` | `cropRotation` |  | landscape files of the crop rotation |

### Section `parameters` (`params.Parameters`)

| gobees Go | NetLogo | gobees-fast Go | Key | Unit | Meaning |
| --- | --- | --- | --- | --- | --- |
| `AbundanceBoost` | `AbundanceBoost` | `AbundanceBoost` | `abundanceBoost` |  | factor on the flower abundance of all foodsources |
| `FoodSourceLimit` | `FoodSourceLimit` | `FoodsourceLimit` | `foodsourceLimit` | [queen trips] | minimum size of a foodsource |
| `ForagingMortalityFactor` | `ForagingMortalityFactor` | `ForagingMortalityFactor` | `foragingMortalityFactor` |  | factor on the foraging mortality |
| `ForagingMortalityModel` | `ForagingMortalityModel` | `ForagingMortalityModel` | `foragingMortalityModel` |  | "high", "intermediate" or "low" |
| `InitialQueensModifierBehaviorSpace` | `InitialQueensModifierBehaviorSpace` | `InitialQueensFactor` | `initialQueensFactor` |  | factor on the initial number of queens |
| `Lambda_detectProb` | `Lambda_detectProb` | `DetectionLambda` | `detectionLambda` | [1/m] | decay of the detection probability with distance |
| `MaxHibernatingQueens` | `MaxHibernatingQueens` | `MaxHibernatingQueens` | `maxHibernatingQueens` |  | maximum number of hibernating queens on 1 January |
| `MergeHedges` | `MergeHedges?` | `MergeHedges` | `mergeHedges` |  | merge small hedge patches |
| `MinSizeFoodSources` | `MinSizeFoodSources?` | `LimitFoodsources` | `limitFoodsources` |  | set nectar or pollen amounts below foodsourceLimit to 0 |
| `N_Badgers` | `N_Badgers` | `Badgers` | `badgers` |  | number of badgers |
| `RemoveEmptyFoodSources` | `RemoveEmptyFoodSources?` | `RemoveEmptyFoodsources` | `removeEmptyFoodsources` |  | remove foodsources without nectar and pollen |
| `SexLocus` | `SexLocus?` | `SexLocus` | `sexLocus` |  | simulate the sex locus (diploid males) |
| `StopExtinct` | `StopExtinct?` | `StopExtinct` | `stopExtinct` |  | stop the run when the population is extinct |
| `UnlimitedMales` | `UnlimitedMales?` | `UnlimitedMales` | `unlimitedMales` |  | queens always find a mate |
| `WinterMortality` | `WinterMortality?` | `WinterMortality` | `winterMortality` |  | winter mortality of hibernating queens |

### Section `output` (`params.Output`)

| gobees Go | NetLogo | gobees-fast Go | Key | Unit | Meaning |
| --- | --- | --- | --- | --- | --- |
| `KeepDeadColonies` | `KeepDeadColonies?` | `KeepDeadColonies` | `keepDeadColonies` |  | count the dead colonies of all years |

### Section `species` (`params.SpeciesFile`)

| gobees Go | NetLogo | gobees-fast Go | Key | Unit | Meaning |
| --- | --- | --- | --- | --- | --- |
| `Name` | `name` | `Name` | `name` |  | name in initialQueens |
| `Parasitic` | `(by the name Psithyrus)` | `Parasitic` | `parasitic` |  | cuckoo bumblebee |
| `EmergingDay_mean` | `emergingDay_mean` | `EmergenceDayMean` | `emergenceDayMean` | [d] | mean day of the year of the emergence from hibernation |
| `EmergingDay_sd` | `emergingDay_sd` | `EmergenceDaySD` | `emergenceDaySD` | [d] | its standard deviation |
| `NestHabitatsList` | `nestHabitatsList` | `NestHabitats` | `nestHabitats` |  | habitat types of the nest sites |
| `Proboscis_min_mm` | `proboscis_min_mm` | `ProboscisMin` | `proboscisMin` | [mm] |  |
| `Proboscis_max_mm` | `proboscis_max_mm` | `ProboscisMax` | `proboscisMax` | [mm] |  |
| `GrowthFactor` | `growthFactor` | `GrowthFactor` | `growthFactor` |  |  |
| `SeasonStop` | `seasonStop` | `SeasonStop` | `seasonStop` | [day of the year] |  |
| `MaxLifespanWorkers` | `maxLifespanWorkers` | `MaxLifespanWorkers` | `maxLifespanWorkers` | [d] |  |
| `Batchsize` | `batchsize` | `BatchSize` | `batchSize` |  | eggs per batch |
| `FlightVelocity_m_per_s` | `flightVelocity_m/s` | `FlightVelocity` | `flightVelocity` | [m/s] |  |
| `SearchLength_m` | `searchLength_m` | `SearchLength` | `searchLength` | [m] | flight distance of a search trip |
| `TimeUnloading` | `timeUnloading` | `UnloadingTime` | `unloadingTime` | [s] |  |
| `SpecMax_cropVolume_myl` | `specMax_cropVolume_myl` | `MaxCropVolume` | `maxCropVolume` | [µl] | crop volume of the largest bee |
| `SpecMax_pollenPellets_g` | `specMax_pollenPellets_g` | `MaxPollenPellets` | `maxPollenPellets` | [g] | pollen load of the largest bee |
| `MinToMaxFactor` | `minToMaxFactor` | `MinToMaxFactor` | `minToMaxFactor` |  |  |
| `DevAgeHatchingMin_d` | `devAgeHatchingMin_d` | `HatchingAgeMin` | `hatchingAgeMin` | [d] |  |
| `DevAgePupationMin_d` | `devAgePupationMin_d` | `PupationAgeMin` | `pupationAgeMin` | [d] |  |
| `DevAgeEmergingMin_d` | `devAgeEmergingMin_d` | `EmergenceAgeMin` | `emergenceAgeMin` | [d] |  |
| `DevWeightEgg_mg` | `devWeightEgg_mg` | `EggWeight` | `eggWeight` | [mg] |  |
| `DevWeightPupationMin_mg` | `devWeightPupationMin_mg` | `PupationWeightMin` | `pupationWeightMin` | [mg] |  |
| `DevWeightPupationMax_mg` | `devWeightPupationMax_mg` | `PupationWeightMax` | `pupationWeightMax` | [mg] |  |
| `PollenToBodymassFactor` | `pollenToBodymassFactor` | `PollenToBodymassFactor` | `pollenToBodymassFactor` |  |  |
| `Dev_Q_DeterminationWeight_mg` | `dev_Q_DeterminationWeight_mg` | `QueenDeterminationWeight` | `queenDeterminationWeight` | [mg] |  |
| `DevAge_Q_PupationMin_d` | `devAge_Q_PupationMin_d` | `QueenPupationAgeMin` | `queenPupationAgeMin` | [d] |  |
| `DevWeight_Q_PupationMin_mg` | `devWeight_Q_PupationMin_mg` | `QueenPupationWeightMin` | `queenPupationWeightMin` | [mg] |  |
| `DevWeight_Q_PupationMax_mg` | `devWeight_Q_PupationMax_mg` | `QueenPupationWeightMax` | `queenPupationWeightMax` | [mg] |  |
| `DevAge_Q_EmergingMin_d` | `devAge_Q_EmergingMin_d` | `QueenEmergenceAgeMin` | `queenEmergenceAgeMin` | [d] |  |
| `DailyNestSiteChance` | `dailyNestSiteChance` | `DailyNestSiteChance` | `dailyNestSiteChance` |  |  |

### Section `flowers` (`params.FlowerSpecies`)

| gobees Go | NetLogo | gobees-fast Go | Key | Unit | Meaning |
| --- | --- | --- | --- | --- | --- |
| `Name` | `Flowerspecies` | `Name` | `name` |  |  |
| `Pollen_g_per_flower` | `pollen_g/flower` | `PollenPerFlower` | `pollenPerFlower` | [g] |  |
| `Nectar_ml_per_flower` | `nectar_ml/flower` | `NectarPerFlower` | `nectarPerFlower` | [ml] |  |
| `Concentration_mol_per_l` | `concentration_mol/l` | `Concentration` | `concentration` | [mol/l] | sugar concentration of the nectar |
| `StartDay` | `startDay` | `StartDay` | `startDay` | [day of the year] | first day of the flowering period |
| `StopDay` | `stopDay` | `StopDay` | `stopDay` | [day of the year] | first day after the flowering period |
| `CorollaDepth_mm` | `corollaDepth_mm` | `CorollaDepth` | `corollaDepth` | [mm] |  |
| `IntFlowerTime_s` | `intFlowerTime_s` | `InterFlowerTime` | `interFlowerTime` | [s] | time from one flower to the next |

### Landscape file (`input.Landscape`, `input.Patch`)

The keys `x` and `y` would need quotes (`y` is a boolean in YAML 1.1), so `xcor` and `ycor` stay. Until step F15b the keys were `metres_per_patch`, `size_sqm` and `flowerSpeciesList`. `params.FlowerPatch` (`Id_Beescout` → `ID`, `Area_sqm` → `Area`, `FlowerSpeciesList` → `Flowers`) and `sys.flowerPatch` (`Radius_m` → `Radius`, `NectarMax_myl` → `NectarMax`, `PollenMax_g` → `PollenMax`) follow.

| gobees Go | NetLogo | gobees-fast Go | Key | Unit | Meaning |
| --- | --- | --- | --- | --- | --- |
| `MetresPerPatch` | `line 1` | `MetresPerPatch` | `metresPerPatch` | [m] | metres per NetLogo patch |
| `Patches` | `lines 2 to n` | `Patches` | `patches` |  | the flower patches |
| `Id` | `id` | `ID` | `id` |  | column 1 |
| `PatchType` | `patchType` | `PatchType` | `patchType` |  | habitat type |
| `X` | `xcor` | `X` | `xcor` | [NetLogo patches] |  |
| `Y` | `ycor` | `Y` | `ycor` | [NetLogo patches] |  |
| `Area_sqm` | `size_sqm` | `Area` | `area` | [m2] |  |
| `FlowerSpeciesList` | `flowerSpeciesList` | `Flowers` | `flowers` |  | flower species and their abundance |

### Derived values (`params.Derived`, `params.Species`, Go only)

| gobees Go | NetLogo | gobees-fast Go | Unit |
| --- | --- | --- | --- |
| `Derived.BeeSpeciesInitialQueensList` | `BeeSpeciesInitialQueensList` | `Derived.InitialQueens` |  |
| `Derived.SpeciesList` | `SpeciesList` | `Derived.SpeciesNames` |  |
| `Derived.MortalityForager_per_s` | `MortalityForager_per_s` | `Derived.ForagerMortality` | [1/s] |
| `Derived.FoodsourcesFile` | `FoodsourcesFile` | `Derived.StartLandscape` |  |
| `globals.Landscape.FoodsourcesFile` | `FoodsourcesFile` | `globals.Landscape.File` |  |
| `Derived.CSS_OptionsList` | `CSS_OptionsList` | `Derived.StewardshipOptions` |  |
| `Derived.CropRotationList` | `CropRotationList` | `Derived.CropRotation` |  |
| `Species.FlightCosts_kJ_per_m_per_mg` | `flightCosts_kJ/m/mg` | `Species.FlightCosts` | [kJ/(m mg)] |
| `Species.DevQuotaIncubationToday_kJ` | `devQuotaIncubationToday_kJ` | `Species.DailyIncubationQuota` | [kJ] |
| `Species.DevAgeHatchingMax_d` | `devAgeHatchingMax_d` | `Species.HatchingAgeMax` | [d] |
| `Species.DevAgePupationMax_d` | `devAgePupationMax_d` | `Species.PupationAgeMax` | [d] |
| `Species.DevAgeEmergingMax_d` | `devAgeEmergingMax_d` | `Species.EmergenceAgeMax` | [d] |
| `Species.DevIncubationHatchingTH_kJ` | `devIncubationHatchingTH_kJ` | `Species.HatchingIncubation` | [kJ] |
| `Species.DevIncubationPupationTH_kJ` | `devIncubationPupationTH_kJ` | `Species.PupationIncubation` | [kJ] |
| `Species.DevIncubationEmergingTH_kJ` | `devIncubationEmergingTH_kJ` | `Species.EmergenceIncubation` | [kJ] |
| `Species.Dev_larvalAge_QueenDetermination_d` | `dev_larvalAge_QueenDetermination_d` | `Species.QueenDeterminationAge` | [d] |
| `Species.DevAge_Q_PupationMax_d` | `devAge_Q_PupationMax_d` | `Species.QueenPupationAgeMax` | [d] |
| `Species.DevAge_Q_EmergingMax_d` | `devAge_Q_EmergingMax_d` | `Species.QueenEmergenceAgeMax` | [d] |
| `Species.DevIncubation_Q_PupationTH_kJ` | `devIncubation_Q_PupationTH_kJ` | `Species.QueenPupationIncubation` | [kJ] |
| `Species.DevIncubation_Q_EmergingTH_kJ` | `devIncubation_Q_EmergingTH_kJ` | `Species.QueenEmergenceIncubation` | [kJ] |
| `Species.MinPollenStore_g` | `minPollenStore_g` | `Species.MinPollenStore` | [g] |

### Constants (`params.Constants`, Go only, keys with step F17)

| gobees Go | NetLogo | gobees-fast Go | Unit |
| --- | --- | --- | --- |
| `CallItaDay_s` | `CallItaDay_s` | `CallItADay` | [s] |
| `EnergyRequiredForPollenAssimilation_kJ_per_g` | `EnergyRequiredForPollenAssimilation_kJ_per_g` | `PollenAssimilationEnergy` | [kJ/g] |
| `EnergySucrose_kJ_per_mymol` | `EnergySucrose_kJ/mymol` | `SucroseEnergy` | [kJ/µmol] |
| `ForagingRangeMax_m` | `ForagingRangeMax_m` | `MaxForagingRange` | [m] |
| `GetUpTime_s` | `GetUpTime_s` | `GetUpTime` | [s] |
| `LarvaWorkerRatioTH` | `LarvaWorkerRatioTH` | `LarvaWorkerRatioLimit` |  |
| `MetabolicRateFlight_W_per_kg` | `MetabolicRateFlight_W/kg` | `FlightMetabolicRate` | [W/kg] |
| `MortalityAdultsBackground_daily` | `MortalityAdultsBackground_daily` | `AdultBackgroundMortality` | [1/d] |
| `N_ForeignAlleles` | `N_ForeignAlleles` | `ForeignAlleles` |  |
| `NestSearchTime_h` | `NestSearchTime_h` | `NestSearchTime` | [h] |
| `QueenDestinedEggsBeforeSP_d` | `QueenDestinedEggsBeforeSP_d` | `QueenEggsBeforeSwitchPoint` | [d] |
| `Sunrise_s` | `Sunrise_s` | `Sunrise` | [s] |

Unchanged: `DailySwitchProbability`, `EnergyFactorOnFlower`, `MaxLifespanMales`, `NotSetHigh`, `NotSetLow`.
Values that the code still converts at run time (`NestSearchTime` in hours, `nectarPerFlower` in ml, the
constants written as `24 * 3600`) keep their unit until a step of their own.
