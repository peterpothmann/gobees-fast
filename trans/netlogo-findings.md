# Findings in the BEE-STEWARD NetLogo code

This file collects places in the NetLogo code that are easily lost when translating to Go, or that would change the model dynamics. Principle for gobees: the NetLogo code is the reference. Quirks are kept and documented, not fixed. Someone could fix them in another model version, though.

Source: `netlogo/__BEE-STEWARD_RC_ColourVersionMB.nlogo`. Codes in square brackets (e.g. `[FPER-02]`) refer to the annotations in the NetLogo code. Line numbers are deliberately omitted because they shift as comments are added.

## Findings from step 1 (setup, parameters, time)

<a id="finding-1"></a>

### 1. The weather list has 366 instead of 365 items

- **Code:** `[FPER-02]`, `Foraging_periodREP`
- **Finding:** The list `foragingHoursExample` contains the items `0` and `.5` at positions 41 and 42 (0-based). `0.5` was probably intended. As a result, the list has 366 items. The model reads item `Day - 1`. Compared to the probably intended list, day 42 gets 0 instead of 0.5, and from day 43 on each day gets the value intended for the previous day. Item 366 is never read.
- **Affects:** only `Weather = "foragingHoursExample"`. The default `"Constant 8 hrs"` is not affected.
- **Consequence for gobees:** copy the list exactly, with 366 items.
- **Second Check**: 23-10-2026, PP
- **Implemented:** `util/foraging_period.go`, list generated token by token from the NetLogo code. Test: `util/foraging_period_test.go` (366 items, day 42 = 0 h, day 43 = 0.5 h).

<a id="finding-2"></a>

### 2. `Habitats.csv` is needed for a simulation run

- **Code:** `[PAR-06]`, `[CSO]`, `AllCSOlistREP`
- **Finding:** During setup, `ParametersProc` calls `AllCSOlistREP`. The reporter reads the `HabitatsFile` and collects the `habitatType` of all rows whose `habitatColourID` contains the text `FlowerSpeciesList_`. For `_SYSTEM_Habitats.csv` the result is `["Legume" "Margin" "newCSO_Legume" "BirdBumble_Margin" "GameCover_Plot" "Plot"]`. `CreateLayersProc` needs this list to recognise entries like `["Margin"]` in a `flowerSpeciesList`. Rows with a numeric `habitatColourID` are skipped silently by `carefully`.
- **Consequence for gobees:** the `HabitatsFile` is an input file of every run, not only of map processing.
- **Second Check**: 23-10-2026, PP
- **Implemented:** `allCSOList` in `sys/parameters.go`. Test: `sys/parameters_test.go` (the 6 names for the default files).

<a id="finding-3"></a>

### 3. `ListFromStringREP` handles edge cases differently from Go

- **Code:** `[LIST]`, used in `[PAR-07]` and `[QINI-01]`
- **Finding:** The procedure splits a string at blanks. Multiple and leading blanks are skipped. Two edge cases differ from `strings.Fields` in Go:

  | Input | NetLogo | `strings.Fields` |
  | --- | --- | --- |
  | `""` | `[""]` (one empty item) | `[]` |
  | `"a b "` | `["a" "b" ""]` | `["a" "b"]` |

- **Why it matters:** `CropRotationListAsString` is `""` by default. NetLogo turns it into `[""]` with length 1, and `UpdateMorning_Proc` starts the crop rotation only for a length greater than 1.
- **Consequence for gobees:** a dedicated function with exactly this behaviour, plus a test for both edge cases.
- **Second Check**: 23-10-2026, PP
- **Implemented:** `util/list_from_string.go`. Test: `util/list_from_string_test.go`. `sys/parameters_test.go` checks that the default crop rotation gives `[""]`.

<a id="finding-4"></a>

### 4. Setup runs at `Day = 0`

- **Code:** `[INIT-10]`
- **Finding:** `SetupWithoutClearingProc` calls `UpdateMorning_Proc` at `ticks = 0`, i.e. at `Day = 0`. `Foraging_periodREP` is only called for `Day > 0`. `DailyForagingPeriod_s` therefore stays 0 until day 1.
- **Consequence for gobees:** the morning routine must run during setup with `Day = 0` and keep the condition `Day > 0`.
- **Second Check**: 23-10-2026, PP

<a id="finding-5"></a>

### 5. An unknown `ForagingMortalityModel` gives a mortality of 0

- **Code:** `[PAR-04]`
- **Finding:** `ParametersProc` sets `MortalityForager_per_s` with three separate `if` statements:

  ```netlogo
  if ForagingMortalityModel = "high"         [ set MortalityForager_per_s 1.0E-05 ]
  if ForagingMortalityModel = "intermediate" [ set MortalityForager_per_s 2.14E-06 ]
  if ForagingMortalityModel = "low"          [ set MortalityForager_per_s 2.75E-07 ]
  ```

  There is no `else` branch. To see which value the variable has when none of the three conditions matches, follow the order of `Setup`:

  1. `Setup` calls `ClearProc`, which runs `clear-all`. `clear-all` sets every variable of the `globals` block to 0, including `MortalityForager_per_s`.
  2. `SetupWithoutClearingProc` calls `ReadAllParametersProc` `[INIT-04]`. It sets `ForagingMortalityModel` from the parameter file (default `"high"`). `MortalityForager_per_s` is not in the file and stays 0.
  3. `SetupWithoutClearingProc` calls `ParametersProc` `[INIT-05]`. Only here, in `[PAR-04]`, `MortalityForager_per_s` can get a value other than 0.

  The comparison with `=` is case-sensitive. Any value other than exactly `"high"`, `"intermediate"` or `"low"` matches none of the conditions, e.g. `"High"`, a typo like `"intermediat"`, or `"medium"`. `MortalityForager_per_s` then keeps the 0 from step 1. NetLogo shows no error message.
  With 0, foragers never die during foraging and queens never die during nest search. The model runs normally, but with different dynamics.
- **Consequence for gobees:** keep the behaviour. In Go a `float64` also starts at 0, so a `switch` without `default` reproduces it exactly. A warning when reading the parameters (e.g. "unknown ForagingMortalityModel, foraging mortality is 0") would not change the dynamics, helps against typos and should be implemented.
- **Second Check**: 23-10-2026, PP
- **Implemented:** `mortalityForager` in `sys/parameters.go`, with the warning. Test: `sys/parameters_test.go` (`"High"`, `"medium"` and `""` give 0).

<a id="finding-6"></a>

### 6. The parameter file and `ParametersProc` share no names

- **Code:** `[INIT-05]`, `[PAR-03]`
- **Finding:** No variable set in `ParametersProc` (37 variables) also appears in `_SYSTEM_Parameters.csv` (47 parameters). `ParametersProc` therefore overwrites no value from the file, although it runs afterwards.
- **Check:** `python trans/check_param_overlap.py` (from the repository root). Rerun it when the parameter file or `ParametersProc` changes, e.g. for a project's own parameter file: `--params netlogo/<map>_Parameters.csv`.
- **Consequence for gobees:** the hard-coded values can be default parameters next to the file parameters. Their order does not matter.
- **Second Check**: 23-10-2026, PP

<a id="finding-7"></a>

### 7. `Date` is for display only

- **Code:** `[DATE]`
- **Finding:** The global `Date` is set in `UpdateMorning_Proc` and only shown in the interface. No model procedure reads it. The model works with `Day`, which `UpdateMorning_Proc` computes separately.
- **Consequence for gobees:** `DateREP` is not translated.
- **Second Check**: 23-10-2026, PP

<a id="finding-13"></a>

### 13. With the default interface settings, NetLogo reads `_SYSTEM_Example_Farm_Parameters.csv`

- **Code:** `[INIT-02]`, `[RPAR-01]`
- **Finding:** The interface stores `MyMap = "_SYSTEM_Example_Farm.png"` and `MyParametersFile = "_SYSTEM_Example_Farm_Parameters.csv"`. For exactly this combination, `[INIT-02]` sets `ProjectsOwnParameterFile?` to false, so `ReadAllParametersProc` loads `MyParametersFile`. A default run therefore uses `_SYSTEM_Example_Farm_Parameters.csv`, not `_SYSTEM_Parameters.csv`. The two files differ only in `SetColour` (`"Magenta"` vs. `"Yellow"`, display only) and in the spelling `true`/`TRUE`. For the model they are equal.
- **Consequence for gobees:** `main.go` reads `_SYSTEM_Example_Farm_Parameters.csv`, like a default NetLogo run.
- **Second Check**: 23-10-2026, PP
- **Implemented:** `main.go`. Test: `params/parameter_file_test.go` reads both files.

## Findings from step 2 (landscape and species)

<a id="finding-14"></a>

### 14. Hedges are merged by default, the smaller hedges lose their flowers

- **Code:** `[HEDG]`, `[FSRC-10]`
- **Finding:** `MergeHedges?` is TRUE in both default parameter files, so `MergeHedgesProc` always runs. It links every hedge patch ("Hedgerow") to its nearest non-hedge flower patch of at least 6500 m² and merges all hedges linked to the same patch into the largest of them. The largest hedge keeps its own position and its own `flowerSpeciesList` and gets the area of the others. The flower lists of the smaller hedges are lost, their area now carries the flower mix of the largest hedge. The smaller hedges get area 0 and are removed in `CreateLayersProc` (`RemoveEmptyFoodSources?`, also TRUE by default).
- **Also:** the procedure adds `nectarMax_myl` and `pollenMax_g` of the merged hedges. Both are still 0 at this point (they are set only in `CreateLayersProc`), so these additions have no effect.
- **To verify in NetLogo:** the order of hedges with equal area in `sort-on` (`[HEDG-02]`). It is either the `who` order or random. It matters if two hedges linked to the same patch have the same area: then it decides which hedge keeps its position and flower list.
- **Checked from the input file (2026-10-23):** equal areas are common in the example farm. Of the 44 hedges, 30 share their area with at least one other hedge, in 9 groups, e.g. 8 hedges with 100 m² and 5 with 300 m². The tie order therefore matters.
- **Verified in NetLogo 5.3.1 (2026-10-23, PP):** `crt 10`, then three times `show sort-on [ 0 ] turtles` gave three different orders, e.g. `[(turtle 4) (turtle 9) (turtle 6) ...]` and `[(turtle 9) (turtle 5) (turtle 7) ...]`. `sort-on` breaks ties RANDOMLY, not by `who`. Which of several equally large hedges keeps its position and flower list is therefore random.
- **Consequence for the merge order in gobees:** shuffle the hedges with the model RNG, then sort them by area (largest first) with a STABLE sort.
- **Consequence for gobees:** reproduce the merge exactly, incl. the loss of the flower lists. The two additions can be dropped (SIMPLIFY), they do not change any value.
- **Second Check**:
- **Implemented:** `mergeHedges` in `sys/create_foodsources.go`: shuffle, then `slices.SortStableFunc` by area. Tests: `TestMergeHedges`, `TestMergeHedgesEqualAreaIsRandom` (both equally large hedges become master in some of 50 runs), `TestMergeHedgesEquallyNearFieldIsRandom`.

<a id="finding-15"></a>

### 15. `layersInPatchList` groups by NetLogo patch cell, not by flower patch

- **Code:** `[LAYR-10]`, read by `Foraging_bestLayerREP`
- **Finding:** After the layers are created, every foodsource gets `layersInPatchList = sort [who] of foodsources-here`: all foodsources on the same NetLogo patch cell, sorted by `who`. If the centre of another flower patch lies in the same cell, its layers are in the list too. A forager at one flower patch can then choose a layer of the other flower patch as the "best layer" of its patch. Every layer gets the list, not only the masterpatch, although the comment at the declaration says "saved for masterpatches only".
- **What follows:** the bee stores the masterpatch it detected in `knownMasterpatchesNectarList` or `knownMasterpatchesPollenList`, but forages on the layer of the other flower patch. After the trip, `Foraging_PatchChoiceProc` looks up the `masterpatchID` of this layer in the list of known patches. Unless the bee also knows the other flower patch, the lookup fails and `AssertionProc` sets `AssertionViolated` ("Possible reason: 2 flower patches ... at same location?"). The bee continues with position -1 for the rest of the day, and `Go` stops the run at the start of the next day (`[GO-01]`).
- **Checked from the input file (2026-10-23, PP):** the 73 flower patches of `_SYSTEM_Example_Farm_Foodsources.txt` lie in 73 different patch cells, with every rounding rule for coordinates ending in .5 (finding 24). The case does not occur for the example farm.
- **Consequence for gobees:** compute the list by patch cell exactly as NetLogo does (rounded coordinates, see the patch cells in `trans/data-model.md`), not by masterpatch. Translate the assertion, so that gobees stops in the same situation.
- **Second Check**: 23.10.2026, PP
- **Implemented:** `setLayersInPatchList` in `sys/create_foodsources.go`, grouped by `util.Pcor`. Test: `TestCreateLayers` (two flower patches in cell (10, 20) share one list). The assertion in `Foraging_PatchChoiceProc` comes with step 6.

<a id="finding-16"></a>

### 16. The `who` order of layers is random

- **Code:** `[LAYR-03]`
- **Finding:** `CreateLayersProc` creates the layers with `hatch` inside `ask foodsources`. The layers of one flower patch get consecutive `who` numbers, but the order between flower patches follows the random order of `ask`. `layersInPatchList` (finding 15) is sorted by `who`, and `Foraging_bestLayerREP` keeps the first layer with the shortest handling time. For layers of different flower patches in the same cell with equal handling time, the random `who` order therefore decides which one a forager takes.
- **Consequence for gobees:** shuffle the flower patches before creating the layers, and give each layer a `Who` number in creation order (decision C and `who` numbers in `trans/data-model.md`).
- **Second Check**: 23.10.2026, PP
- **Implemented:** `createLayers` shuffles the flower patches before the layers get their `Who` numbers. Test: `TestCreateLayers`.

<a id="finding-17"></a>

### 17. `farmland?` is never read

- **Code:** `[FSRC-07]`, `[LAYR-05]`
- **Finding:** `CreateFoodsourcesProc` and `CreateLayersProc` set `farmland?` by comparing the `who` number with the parameter `MyFarmlandPatches`. No procedure reads `farmland?`. In `CreateLayersProc` the comparison even uses the layer's random `who` number (finding 16). Both have no effect.
- **Consequence for gobees:** do not translate `farmland?`. `MyFarmlandPatches` is read from the parameter file but not used (like `MaxForagingRange_m`, finding 8).
- **Second Check**: 23.10.2026, PP
- **Implemented:** `farmland?` and `MyFarmlandPatches` are not translated (`comp.Foodsource` has no such field).

<a id="finding-18"></a>

### 18. Removed foodsources leave gaps in the `who` numbers

- **Code:** `[FSRC-08]`, `[LAYR-07]`
- **Finding:** A flower patch outside `MapAreaIncluded` and a layer without nectar and pollen are first created and then removed with `die`. Each has already used up a `who` number. NetLogo never reuses `who` numbers within a run.
- **Consequence for gobees:** the `Who` counter must also count these removed foodsources, otherwise the `who` order differs from NetLogo whenever it matters (finding 16).
- **Second Check**: 23.10.2026, PP
- **Implemented:** `globals.WhoCounter`. `CreateFoodsources` counts every flower patch, also outside `MapAreaIncluded`, and every removed layer. Test: `TestCreateLayers` (gap between layers A and C).

<a id="finding-19"></a>

### 19. A flower patch with an empty flower list never offers food

- **Code:** `[LAYR-09]`
- **Finding:** `CreateLayersProc` splits only flower patches with a non-empty `flowerSpeciesList`. A flower patch with an empty list stays as it is: a masterpatch with `nectarMax_myl = pollenMax_g = 0` and `startDay = stopDay = 0`, the values after creation. It never offers food, but it can still be a nest site (`[SPEC-06]`). `_SYSTEM_Example_Farm_Foodsources.txt` contains no empty list.
- **Consequence for gobees:** keep such flower patches as masterpatches without food.
- **Second Check**: 23.10.2026, PP
- **Implemented:** `createLayers` in `sys/create_foodsources.go`. Test: `TestCreateLayers` (the Scrub patch with an empty list).

<a id="finding-23"></a>

### 23. `masterpatchesInRangeList` is never read

- **Code:** `PatchesInRangeProc` (not yet annotated), `CropRotationProc`
- **Finding:** `PatchesInRangeProc` sets the colony variable `masterpatchesInRangeList` to all masterpatches within `ForagingRangeMax_m`, and `CropRotationProc` resets it to `[]`. No procedure reads it. Searching foragers use `masterpatchesWithNectarlayersInFlowerAndRangeList` and `masterpatchesWithPollenlayersInFlowerAndRangeList` from `FoodsourcesInFlowerAndRangeProc`, which are built from `allPatchesInRangeList`. The `ask` in `PatchesInRangeProc` fills both lists in the same random order, so dropping one of them does not change the order of the other.
- **Consequence for gobees:** do not translate `masterpatchesInRangeList` (SIMPLIFY).
- **Second Check**:

<a id="finding-20"></a>

### 20. Species: file order, rounded standard deviation, substring match of nest habitats

- **Code:** `[SPEC-02]`, `[SPEC-03]`, `[NHAB-02]`
- **Finding:**
  - Species are created in the order of the rows in the `BeespeciesFile`, not in the order of `BeeSpeciesInitialQueensListAsString`. Their `who` numbers follow the file. A row that appears twice in the file creates the species twice. A species in the parameter list that is missing in the file creates no species, and `CreateInitialQueensProc` fails later.
  - `emergingDay_sd` is rounded with `round`. In the default species file: 14.5 → 15 (B. pascuorum), 20.5 → 21 (B. lapidarius), 30.5 → 31 (B. pratorum).
  - A nest habitat ending with `*` matches every habitat type on the map that contains the text before the `*` anywhere, e.g. `"Grassland*"` also matches `"Improved_Grassland"`.
- **Consequence for gobees:** reproduce all three. NetLogo `round` rounds .5 up, `math.Round` in Go rounds .5 away from zero, which gives the same result for positive values.
- **Second Check**: 23.10.2026, PP
- **Implemented:** `parseSpeciesFile` and `nestHabitatsList` in `sys/create_species.go`. Tests: `TestParseSpeciesFileOrder`, `TestParseSpeciesFileExample` (rounded standard deviations), `TestNestHabitatsList`.

<a id="finding-21"></a>

### 21. Reading the foodsources file depends on `file-read` behaviour

- **Code:** `[FSRC-03]`
- **Finding:** After reading the scaling value from line 1 with `file-read`, the procedure skips the header with `file-read-line`: if the next line has at most 10 characters (old format: number of flower patches), one more line is skipped. This works only if `file-read` has already consumed the end of line 1. With `_SYSTEM_Example_Farm_Foodsources.txt` (line 1 `10`, line 2 `73`, line 3 header, 73 data lines) the model evidently works.
- **Verified in NetLogo 5.3.1 (2026-10-23, PP):** after `Setup` with the default settings, `show count foodsources with [ masterpatch? ]` gave **36**. Of the 73 flower patches read, 36 remain as masterpatches after merging hedges and removing empty layers. The file is therefore read completely. The number can vary between runs, because the merge order of equally large hedges is random (finding 14).
- **Test value for gobees:** 73 flower patches read. After setup, 36 masterpatches in this NetLogo run. To be compared as a distribution over several runs, not as an exact value.
- **Consequence for gobees:** reproduce the effective behaviour: line 1 scaling, line 2 skipped, line 3 skipped if line 2 has at most 10 characters. Test with the example file: 73 flower patches are read.
- **Second Check**: 23.10.2026, PP
- **Implemented:** `parseFoodsourcesFile` in `sys/foodsources_file.go`. Tests: `TestParseFoodsourcesFileExampleFarm` (73 flower patches), `TestParseFoodsourcesFileHeaderFormats`. `TestCreateFoodsourcesExampleFarm`: 36 masterpatches with each of 20 seeds, the same number as in NetLogo.

<a id="finding-24"></a>

### 24. Coordinates ending in .5 lie exactly on a patch border

- **Code:** `[LAYR-10]` (`foodsources-here`), `NestSitesSearchingProc` and `CreateBadgersProc` (`pxcor`, `pycor`, not yet annotated)
- **Finding:** 13 of the 73 flower patches of `_SYSTEM_Example_Farm_Foodsources.txt` have an x or y coordinate ending in .5, e.g. patch 2 (Hedgerow, 132 / 170.5) and patch 20 (Scrub, 202 / 121.5). Such a point lies exactly on the border between two NetLogo patch cells. NetLogo's rounding rule for .5 decides the cell. Colonies and badgers are placed at the patch centre (`pxcor`, `pycor`) of their foodsource, and Hedgerow and Scrub are nest habitats. With the other rounding direction, a colony would sit one patch (10 m with the example scaling) further away, which changes all its flight distances slightly. For `layersInPatchList` the rule makes no difference with the example farm: under every rule all 73 patches lie in different cells.
- **Verified in NetLogo 5.3.1 (2026-10-23, PP):** after `Setup`, `ask foodsources with [ ycor = 121.5 ] [ show (list xcor ycor pxcor pycor) ]` (flower patch 20, Scrub) gave `[202 121.5 202 122]` for all 6 layers. NetLogo rounds .5 UP. (The hedge at y = 170.5 no longer exists after `Setup`, it was merged.) The x direction was not checked separately, NetLogo uses the same rule for both.
- **Side result:** the 6 layers have the `who` numbers 360, 363, 366, 367, 368 and 373. The gaps are layers of the same flower patch that were removed as empty, which confirms finding 18.
- **Consequence for gobees:** patch cell of a point = `math.Floor(v + 0.5)` (round half up; all coordinates of the world are >= 0). Test with the 13 border coordinates of the example farm, e.g. 121.5 → 122 and 210.5 → 211.
- **Second Check**: 23.10.2026, PP
- **Implemented:** `util.Pcor`. Test: `TestPcor` with the 13 border coordinates of the example farm.

## Findings from the model map and from `Go`

<a id="finding-8"></a>

### 8. The foraging range from the parameter file is not used

- **Code:** `[PAR-03]`
- **Finding:** `MaxForagingRange_m` is in `_SYSTEM_Parameters.csv` (value 10000), but no procedure reads it. The model uses `ForagingRangeMax_m`, which `ParametersProc` sets to a fixed 758 m.
- **Consequence for gobees:** keep 758 m as a fixed value. Read `MaxForagingRange_m` from the file, but do not use it.
- **Second Check**:

<a id="finding-9"></a>

### 9. `deadCols` loses `speciesNameColony`

- **Code:** not yet annotated (`DieProc`)
- **Finding:** Dead colonies change their breed in `DieProc` via `set breed deadCols`. `colonies-own` has the variable `speciesNameColony`, `deadCols-own` does not. The comment at the declaration demands an exact copy. The value is lost when the colony dies.
- **Who reads `deadCols`:** checked on 2026-10-23. Only three places: the egg accounting assertion in `CheckNumbersProc` on day 364, `UpdateColoniesProc` (deletes all dead colonies on 1 January unless `KeepDeadColonies?`), and the GUI (white dots if `ShowDeadCols?`). No output, plot or BehaviorSpace experiment reads `deadCols`, and no model procedure iterates over all `turtles`. Nobody reads `speciesNameColony` of a dead colony, so the loss has no effect.
- **Consequence for gobees:** no effect on the dynamics. gobees deletes the colony entity at death and writes its data into a record of dead colonies, including `speciesNameColony` (decision E in `trans/data-model.md`). The egg accounting check runs at death instead of on day 364.
- **Second Check**: 23.10.2026, PP

<a id="finding-10"></a>

### 10. The value in the flower list acts as flowers per m²

- **Code:** `[LAYR-05]`
- **Finding:** The code calls the value from `flowerSpeciesList` "relative abundance". It is, however, multiplied by the area and by pollen or nectar *per flower*: `pollenMax_g = area_sqm × AbundanceBoost × value × pollen_g/flower` and `nectarMax_myl = area_sqm × AbundanceBoost × value × nectar_ml/flower × 1000`. Its unit is therefore flowers per m².
- **Status:** confirmed when annotating `CreateLayersProc` on 2026-10-23. The name has no effect on the dynamics. Only the interpretation of the input data is affected.
- **Second Check**: 23.10.2026, PP
- **Implemented:** `createLayers` in `sys/create_foodsources.go` (`RelativeAbundance` stores the value without `AbundanceBoost`). Test: `TestCreateLayers` (100 m2 x 2 x 0.01 g = 2 g pollen).

<a id="finding-11"></a>

### 11. In `ActivityProc` the last matching task wins

- **Code:** not yet annotated (`ActivityProc`)
- **Finding:** The four comparisons of stimulus and threshold are separate consecutive `if` statements. Each matching condition overwrites `activity`. If several match, the order of precedence is therefore: nectar foraging before pollen foraging before nursing before egg laying. The comment in the code ("ordered by their importance") does not say whether the first or the last check takes precedence.
- **Consequence for gobees:** keep the exact order of the checks, do not write them as `if … else if`.
- **Second Check**: 23.10.2026, PP

<a id="finding-12"></a>

### 12. NetLogo `ticks` and ark-tools `resource.Tick` differ by 1

- **Code:** `[GO-03]`
- **Finding:** NetLogo calls `tick` at the start of `Go`. During the first simulated day `ticks` is 1, during `Setup` it is 0. ark-tools increments `resource.Tick` only after each update, so its value is 0 during the first day.
- **Consequence for gobees:** the system `sys.Tick` writes the NetLogo ticks into `globals.Ticks`. All systems read only this resource. The test `sys/tick_test.go` covers this.
- **Second Check**:

<a id="finding-22"></a>

### 22. Background mortality of adults has an inverted condition

- **Code:** not yet annotated (`Development_Mortality_AdultsProc`)
- **Finding:** The condition is `MortalityAdultsBackground_daily > 0 and random-float 1 > MortalityAdultsBackground_daily`. A bee therefore dies with probability `1 - MortalityAdultsBackground_daily`, not with `MortalityAdultsBackground_daily`. `ParametersProc` sets the value to 0, so the first part of the condition is false and nobody dies from it.
- **Affects:** nothing with the hard-coded value 0. A value of e.g. 0.01 would kill 99 % of all adults (workers, males and all adult queens, including hibernating ones) every day.
- **Consequence for gobees:** translate the condition exactly as it is. A test with value 0 checks that no bee dies.
- **Second Check**: 23.10.2026, PP
