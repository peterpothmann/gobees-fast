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
- **Implemented:** `AllCSOlistREP` in `sys/parameters.go`. Test: `sys/parameters_test.go` (the 6 names for the default files).

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
- **Implemented:** `MergeHedgesProc` in `sys/create_foodsources.go`: shuffle, then `slices.SortStableFunc` by area. Tests: `TestMergeHedges`, `TestMergeHedgesEqualAreaIsRandom` (both equally large hedges become master in some of 50 runs), `TestMergeHedgesEquallyNearFieldIsRandom`.

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
- **Implemented:** `CreateLayersProc` shuffles the flower patches before the layers get their `Who` numbers. Test: `TestCreateLayers`.

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
- **Implemented:** `CreateLayersProc` in `sys/create_foodsources.go`. Test: `TestCreateLayers` (the Scrub patch with an empty list).

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
- **Implemented:** `parseSpeciesFile` and `NestHabitatsListREP` in `sys/create_species.go`. Tests: `TestParseSpeciesFileOrder`, `TestParseSpeciesFileExample` (rounded standard deviations), `TestNestHabitatsListREP`.

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

- **Code:** `[LAYR-10]` (`foodsources-here`), `CreateBadgersProc` (`pxcor`, `pycor`, not yet annotated)
- **Finding:** 13 of the 73 flower patches of `_SYSTEM_Example_Farm_Foodsources.txt` have an x or y coordinate ending in .5, e.g. patch 2 (Hedgerow, 132 / 170.5) and patch 20 (Scrub, 202 / 121.5). Such a point lies exactly on the border between two NetLogo patch cells. NetLogo's rounding rule for .5 decides the cell. Badgers are placed at the patch centre (`pxcor`, `pycor`) of their foodsource. For `layersInPatchList` the rule makes no difference with the example farm: under every rule all 73 patches lie in different cells.
- **Correction (2026-10-24, annotating step 3, PP):** an earlier version said that colonies are placed at the patch centre of their foodsource too, so that the rounding moves colonies. That is wrong. `NestSitesSearchingProc` places a colony at the centre of a RANDOM patch whose centre lies within the circle of the foodsource (`[NEST-03]`, finding 25). No rounding of the foodsource position is involved, so the rule does not affect colonies.
- **Verified in NetLogo 5.3.1 (2026-10-23, PP):** after `Setup`, `ask foodsources with [ ycor = 121.5 ] [ show (list xcor ycor pxcor pycor) ]` (flower patch 20, Scrub) gave `[202 121.5 202 122]` for all 6 layers. NetLogo rounds .5 UP. (The hedge at y = 170.5 no longer exists after `Setup`, it was merged.) The x direction was not checked separately, NetLogo uses the same rule for both.
- **Side result:** the 6 layers have the `who` numbers 360, 363, 366, 367, 368 and 373. The gaps are layers of the same flower patch that were removed as empty, which confirms finding 18.
- **Consequence for gobees:** patch cell of a point = `math.Floor(v + 0.5)` (round half up; all coordinates of the world are >= 0). Test with the 13 border coordinates of the example farm, e.g. 121.5 → 122 and 210.5 → 211.
- **Second Check**: 23.10.2026, PP
- **Implemented:** `util.Pcor`. Test: `TestPcor` with the 13 border coordinates of the example farm.

## Findings from step 3 (queens and colony foundation)

<a id="finding-25"></a>

### 25. A colony is placed in a random patch within the nest-site circle

- **Code:** `[NEST-03]`
- **Finding:** `NestSitesSearchingProc` does not place the colony at the chosen foodsource. It takes `one-of patches with [ distance myself < radius_m * Scaling_NLpatches/m ]`, i.e. a random patch whose CENTRE lies within the circle of the foodsource, and moves the queen to that patch centre. The colony is created there. Large flower patches therefore spread their colonies over many cells.
- **Edge case:** if no patch centre lies within the circle (a very small foodsource whose centre is far from every patch centre), `one-of` reports `nobody` and `ask nobody` stops NetLogo with a runtime error.
- **Checked from the input file (2026-10-24):** in the example farm every possible nest site of B. terrestris (Grassland, Hedgerow, Scrub, Woodland) has at least one patch centre in its circle. 15 small hedges and woods (100 to 300 m², radius 0.56 to 0.98 patches) have exactly one. The error does not occur.
- **Consequence for gobees:** reproduce the random choice among the patch centres within the circle (strict `<`, distance in patch units). Translate the runtime error as an error. This also corrects finding 24.
- **Second Check**: 24.10.2026, PP
- **Implemented:** `NestSitesSearchingProc` in `sys/emergence_new_queens.go` with `util.PatchCentresWithin` (strict `<`). No patch centre: panic. Tests: `TestNestSearchFoundsColonies` (colonies at patch centres, the 100 m2 hedge only at (50, 50)), `TestPatchCentresWithin`.

<a id="finding-26"></a>

### 26. Queens searching a nest die with about 19 % per day

- **Code:** `[NEST-04]`
- **Finding:** a queen without a nest finds one with probability `chanceFindNest` (0.2 in the species file). If she finds none, she dies with probability `1 - (1 - MortalityForager_per_s) ^ (NestSearchTime_h * 3600)`. With the default values (`"high"`: 1e-5 per s, 6 h): `1 - 0.99999 ^ 21600 = 0.194`. Per day 20 % of the searching queens found a nest and 0.8 × 0.194 = 15.5 % die. In the long run 0.2 / (0.2 + 0.155) = 56 % of the emerged queens found a colony, 44 % die during the search.
- **Consequence for gobees:** nothing special, but a strong lever of the dynamics and worth knowing for the validation. With an unknown `ForagingMortalityModel` no queen dies here (finding 5).
- **Second Check**: 24.10.2026, PP
- **Implemented:** `NestSitesSearchingProc` and `searchMortality` in `sys/emergence_new_queens.go`. Test: `TestNestSearchMortality` (2000 queens, about 19.4 % die).

<a id="finding-27"></a>

### 27. The area-weighted nest-site choice can run past the end of the list

- **Code:** `[NSFS-01]`
- **Finding:** `NestSiteFoodSourceREP` draws `p = random-float 1` and walks through the nest sites, summing `area_sqm / nestSiteArea`, until the sum exceeds `p`. In floating point the full sum can end slightly below 1. If `p` lies between this sum and 1, the loop reads past the end of the list and NetLogo stops with a runtime error.
- **Consequence for gobees:** reproduce the loop. Running past the end is an error, as in NetLogo.
- **Second Check**: 24.10.2026, PP 
- **Implemented:** `NestSiteFoodSourceREP` in `sys/emergence_new_queens.go`: running past the end of the list panics.

<a id="finding-28"></a>

### 28. `die` ends the agent's code at once

- **Code:** `[DIE]`, e.g. `[EMER-03]`
- **Finding:** `DieProc` calls `die`. In NetLogo, `die` ends the execution of the dying agent immediately: commands after `DieProc` in the same `ask` block are not run for this agent. Example: a queen that dies over winter in `EmergenceNewQueensProc` never gets `activity "resting"` and `colonyID -1`.
- **Consequence for gobees:** after the Go equivalent of `DieProc` the calling code must stop working on this agent (`return` or `continue`). This is a rule for every translated procedure, not a quirk.
- **Second Check**: 24.10.2026, PP
- **Implemented:** `beeDeath.DieProc` in `sys/die.go`; the callers `continue` or `return` after it. Test: `TestDieRemovesFromActiveBees`.

<a id="finding-29"></a>

### 29. `allPatchesInRangeList` is in random order

- **Code:** `[PRNG-01]`
- **Finding:** `PatchesInRangeProc` builds the list of foodsources within the foraging range with `fput` inside `ask foodsources with [...]`. Since `ask` visits the agents in random order, the list is in random order. `FoodsourcesInFlowerAndRangeProc` keeps this order for the nectar and pollen lists, also for the lists of masterpatches.
- **Consequence for gobees:** shuffle the foodsources, then build the list in this order. Whether the order matters for the foragers is decided in step 6 (`Foraging_searchingProc`).
- **Second Check**: 24.10.2026, PP
- **Implemented:** `PatchesInRangeProc` in `sys/create_colonies.go`: shuffle, then append. Test: `TestPatchesInRangeAndFlowerLists`.

<a id="finding-30"></a>

### 30. `COLONIES_IBM` counts living colonies only

- **Code:** `[CCOL-05]`
- **Finding:** a new colony is individual-based while `count Colonies with [ cohortBased? = false ] < COLONIES_IBM`. Dead colonies are `deadCols` (finding 9) and are no longer counted, so after an individual-based colony dies the next new colony becomes individual-based again. The default is `COLONIES_IBM = 0`: all colonies are cohort-based.
- **Consequence for gobees:** count only living individual-based colonies.
- **Second Check**: 24.10.2026, PP
- **Implemented:** `countIBMColonies` in `sys/create_colonies.go` counts living colonies. Test: `TestColoniesIBM`.

<a id="finding-31"></a>

### 31. The initial queens' `personalTime_s` does not match its comment

- **Code:** `[IQUE-06]`
- **Finding:** `set personalTime_s random (2 * 3600) + (GetUpTime_s - 3600)` with the comment "between 7:00 and 9:00 am". With `GetUpTime_s = 1` [PAR-03] the result is -3599 to 3600 s. The comment probably dates from a time when `GetUpTime_s` was 8 h. No effect: `UpdateMorning_Proc` sets a new `personalTime_s` for every active bee each morning, and hibernating queens are not active.
- **Consequence for gobees:** translate as it is, or leave out (SIMPLIFY). It has no effect on the dynamics.
- **Second Check**: 24.10.2026, PP
- **Implemented:** `CreateInitialQueens` in `sys/create_initial_queens.go` keeps the formula. Test: `TestCreateInitialQueensExampleFarm` (whole seconds from -3599 to 3600).

<a id="finding-32"></a>

### 32. Queens emerging from a cohort are identical copies

- **Code:** `[EMER-01]`
- **Finding:** a queen agent with `number = n` (a cohort from a cohort-based colony) is split at emergence: `hatch (n - 1)` creates copies with all her values, i.e. the same weight, alleles, spermatheca and emergence date. The initial queens are created with `number 1` and are not affected. The copies then draw winter survival and nest search independently, but with the same weight and therefore the same survival probability.
- **Consequence for gobees:** create `n - 1` copies of the queen entity with all components, `Number = 1` for all.
- **Second Check**: 24.10.2026, PP
- **Implemented:** `copyBee` in `sys/emergence_new_queens.go` (lists are copied, not shared). Test: `TestEmergenceSplitsCohort`.

<a id="finding-33"></a>

### 33. One winter survival curve for all species, at most 64 %

- **Code:** `[WMOR]`
- **Finding:** `WintermortalityProbREP` reports the probability to SURVIVE the winter (despite its name): `0.64 / (1 + e^(-22 * (relative weight - 0.32)))`, with the weight relative to the minimum and maximum queen weight of the species. The curve was fitted to B. terrestris (Beekman et al. 1998) and is used for all species. Even the heaviest queens survive with at most 64 %.
- **Consequence for gobees:** Keep the NetLogo name.
- **Second Check**: 24.10.2026, PP
- **Implemented:** `WintermortalityProbREP` in `sys/bee_traits.go`. Tests: `TestWintermortalityProbREP`, `TestEmergenceWinterMortality` (2000 queens at relative weight 0.32, about 32 % survive).

<a id="finding-34"></a>

### 34. One proboscis line from the lightest worker to the heaviest queen

- **Code:** `[PROB]`
- **Finding:** `ProboscisLengthREP` interpolates linearly between `proboscis_min_mm` at the minimum WORKER weight (`devWeightPupationMin_mg`) and `proboscis_max_mm` at the maximum QUEEN weight (`devWeight_Q_PupationMax_mg`). Workers, males and queens are on the same line. A weight outside this range is an assertion.
- **Consequence for gobees:** translate as it is, incl. the assertion.
- **Second Check**: 24.10.2026, PP
- **Implemented:** `ProboscisLengthREP` in `sys/bee_traits.go`. Test: `TestProboscisLength`.

<a id="finding-35"></a>

### 35. A cuckoo bee takes any colony as host

- **Code:** `[PSI-02]`
- **Finding:** `PsithyrusNestSearchProc` chooses the host with `one-of colonies`: a colony of any species, also one that another cuckoo bee has already taken. The probabilities (find one nest 0.05, get access 0.25, be killed 0.25, kill the host queen 0.5) are marked as arbitrary in the code. Not part of the default runs.
- **Consequence for gobees:** translate as it is.
- **Second Check**: 24.10.2026, PP
- **Implemented:** `PsithyrusNestSearchProc` in `sys/emergence_new_queens.go`. Not tested yet (not part of the default runs).

## Findings from step 4a (morning routine)

<a id="finding-36"></a>

### 36. The day of the year uses the divisor 365.00000001

- **Code:** `[MORN-02]`
- **Finding:** `Day = round (ticks mod 365.00000001)`. With `ticks mod 365`, the last day of each year would be 0. With the slightly larger divisor, `ticks = 365` gives 365. From the second year on, the remainder lies a little below the intended integer (e.g. `ticks = 366` gives 0.99999999), and `round` corrects this. The error grows by 1e-8 per year and would only matter after about 50 million years.
- **Result:** `Day = ((ticks - 1) mod 365) + 1` for `ticks > 0`, and `Day = 0` in Setup.
- **Consequence for gobees:** compute `Day` with integers. A test compares the integer formula with the NetLogo formula for ticks 0, 1, 364, 365, 366, 730 and 731.
- **Second Check**: 24.10.2026, PP
- **Implemented:** `dayOfYear` in `sys/update_morning.go`. Test: `TestDayOfYear` (the worked examples, and the integer formula against the NetLogo formula for 100 years). The step 4a validation compares `Day` with NetLogo over 400 days.

<a id="finding-37"></a>

### 37. After the competition point, colony death reads the brood counts of the previous morning

- **Code:** `[UCOL-02]`, `[UCOL-06]`
- **Finding:** `UpdateColoniesProc` checks `allEggs + allLarvae + allPupae = 0` before it recomputes these counts in the same block. The check therefore uses the values of the previous morning. Brood that emerged or died since then is still counted, so the workers and queens die one day later than with current counts. Only workers and queens with `adultAge > 10` die. Younger adults stay and die on a later morning, once they are older than 10 days.
- **Consequence for gobees:** keep the order: deaths first, then the statistics. Do not compute the counts on demand.
- **Second Check**: 24.10.2026, PP
- **Implemented:** `UpdateColonies.updateColony` in `sys/update_colonies.go`, deaths before statistics. Test: `TestUpdateColoniesDeathAfterCP` (the adults survive one more day because of the pupae counted on the previous morning).

<a id="finding-38"></a>

### 38. Males keep their colony alive

- **Code:** `[UCOL-02]`, `[UCOL-04]`, `[UCOL-11]`, `[CHK-03]`
- **Finding:** adult males keep their `colonyID`. Only young queens that leave the nest get `colonyID = -1` (`[GO-12]`). A colony dies only when no bee with its `colonyID` is left (`[UCOL-11]`). Males are not killed in `[UCOL-02]`, and in `[UCOL-04]` they count as adults.
- **Affects:**
  - A colony with only males left stays alive until the last male dies: after `MaxLifespanMales` (30 days), at the end of the season, or when `UpdateSeasonalEventsProc` kills all males. It counts as a living colony, i.e. for `COLONIES_IBM` (finding 30, see trans/netlogo-findings.md#finding-30) and for the loop condition `count colonies > 0` (`[GO-06]`).
  - Brood whose only adults are males is not killed for "no adults left" (`[UCOL-04]`). This needs brood without any worker or queen, which is probably rare.
- **Consequence for gobees:** keep the `ColonyID` of males. A colony entity is deleted only when no bee refers to it. So a bee never refers to a deleted colony, and the `ecs.Entity` references stay valid (decision B in `trans/data-model.md`).
- **Second Check**: 24.10.2026, PP
- **Implemented:** `sys/update_colonies.go` keeps males in `[UCOL-02]` and counts them as adults in `[UCOL-04]`. A colony is deleted only in `[UCOL-11]`, when no bee refers to it. Tests: `TestUpdateColoniesDeathAfterCP` (a male alone keeps the colony alive), `TestUpdateColoniesNoAdults` (a male keeps the brood alive).

<a id="finding-39"></a>

### 39. `CheckNumbersProc` is not only a check

- **Code:** `[CHK-02]`, `[MORN-12]`
- **Finding:** besides its assertions, `CheckNumbersProc` sets `TotalEggs`, `TotalLarvae`, `TotalPupae`, `TotalAdultWorkers`, `TotalAdultQueens`, `TotalAdultMales` and `TotalAdults`, and resets `TotalForagingTripsToday` to 0. `UpdateSeasonalEventsProc` reads `TotalEggs + TotalLarvae + TotalPupae` on the next morning, because it runs before `CheckNumbersProc`. It kills all males if, among other things, this sum is 0.
- **Affects:** a translation that drops `CheckNumbersProc` as "only checks" changes when the males die, and the outputs. The BehaviorSpace metrics of these totals are morning values.
- **Consequence for gobees:** a system `CheckNumbers` that sets the totals and panics where NetLogo asserts.
- **Second Check**: 24.10.2026, PP
- **Implemented:** system `CheckNumbers` in `sys/check_numbers.go`, totals in `globals.Totals`. Test: `TestCheckNumbersTotals` (all totals, `TotalForagingTripsToday` reset to 0). The colony check on day 364 is switched off in `model/default.go` (`SkipColonyCheck`) until step 7 translates the end of season.

## Findings from step 4b (brood)

<a id="finding-40"></a>

### 40. Once a colony produces queens, every female larva becomes a queen larva

- **Code:** `[CAST-01]`, `[QPD-02]`
- **Finding:** `DetermineCaste_REP` makes a female larva at larval day 3 a queen if `weight_mg >= dev_Q_DeterminationWeight_mg` and her colony has `queenProduction? = true`. `dev_Q_DeterminationWeight_mg` is 0 for all seven species in `_SYSTEM_BumbleSpecies_UK_01.csv`, so the weight criterion has no effect. From the evening on which `queenProduction?` becomes true, every female larva that reaches larval day 3 becomes a queen larva, however well she was fed.
- **Affects:** whether a queen larva becomes an adult queen depends only on her food and incubation afterwards. B. terrestris needs 590 mg by `broodAge` 29, otherwise she dies (`larvaDeathsWeight`, `[MORB-02]`). After the switch point the queen lays only male eggs (`[EGG-02]`), so the young queens come from the last female eggs laid before the switch point.
- **Consequence for gobees:** translate as it is, with the weight comparison.
- **Second Check**: 24.10.2026
- **Implemented:** `DetermineCaste_REP` in `sys/development.go`, with the weight comparison. Test: `TestDetermineCaste_REP` (every weight gives a queen with queen production, the weight criterion works with a value above 0).

<a id="finding-41"></a>

### 41. NetLogo evaluates `and` and `or` from left to right, Go does not

- **Code:** `[DEVL-03]`
- **Finding:** in NetLogo 5.3.1 `and`, `or` and `xor` have the same precedence and are evaluated from left to right. Checked headless on 2026-10-24: `true or false and false` gives `false`, `true or true and false and false` gives `false`. In Go `&&` binds tighter than `||`. The pupation condition `caste = "worker" or caste = "male" and <incubation> and <age> and <weight>` means `(worker or male) and ...` in NetLogo. Written the same way in Go, it would mean `worker || (male && ...)`, and every worker larva would pupate on the first evening.
- **Search of the whole code (2026-10-24):** a script listed all statements that mix `and` and `or` at the same bracket level. This is the only one. The Go code of steps 1 to 4a brackets every mix of `&&` and `||`.
- **Consequence for gobees:** brackets in Go: `(worker || male) && ...`. A test checks that a worker larva that is too light does not pupate.
- **Second Check**: 24.10.2026
- **Implemented:** `Development_LarvaeProc` in `sys/development.go`, `(worker || male) && ...`. Test: `TestDevelopmentLarvaePrecedence` (worker and male larvae that are too light, too young or too cold stay larvae).

<a id="finding-42"></a>

### 42. `queenProductionDate` is the laying day of the last worker larvae

- **Code:** `[QPD-01]`
- **Finding:** `QueenProductionDateProc` runs in the evening before `DevelopmentProc`. The larvae it uses (`cumulTimeLarva_d = 3`) reached larval day 3 on the previous evening and got their caste then, while `queenProduction?` was still false, so they became workers. `queenProductionDate = ticks - broodAge` is their laying day. The first queen larvae are those that reach larval day 3 tonight, in `DevelopmentProc` after this procedure. With eggs laid every day, they were laid one day later.
- **Affects:** `queenProductionDate` is about one day earlier than the laying day of the first queen-destined eggs, contrary to the comment in the code ("these larvae were laid as eggs on that day"). The competition point (`[CPD-01]`) then comes about 0.7 days earlier (y = 0.7 x + 15.5, rounded). The comment calls the date "theoretical", so the effect may be accepted.
- **Consequence for gobees:** translate as it is, in the same order of the systems (QueenProductionDate before Development).
- **Second Check**: 24.10.2026
- **Implemented:** `QueenProductionDate` runs before `Development` in `model/default.go`. Test: `TestQueenProductionDateBeforeDevelopment` (date 102 = laying day of the worker larva, the larva laid on day 103 becomes a queen).

<a id="finding-43"></a>

### 43. Individual-based colonies pay for one egg per batch

- **Code:** `[EGG-01]`
- **Finding:** `EgglayingProc` computes the cost as `beesInCohort * devWeightEgg_mg * pollenToBodymassFactor / 1000`. A cohort-based colony lays 1 cohort of `batchsize` eggs (`beesInCohort = batchsize`). An individual-based colony lays `batchsize` cohorts of 1 egg (`beesInCohort = 1`) and therefore pays the cost of one egg for all `batchsize` eggs. The check whether the stores cover the cost uses the same lower value.
- **Affects:** only colonies with `cohortBased? = false`, i.e. runs with `COLONIES_IBM > 0`. `COLONIES_IBM` is 0 in `_SYSTEM_Parameters.csv` and `_SYSTEM_Example_Farm_Parameters.csv`, so the default runs are not affected.
- **Consequence for gobees:** translate as it is. A test with an individual-based colony checks the cost of one egg.
- **Second Check**: 24.10.2026
- **Implemented:** `eggLayer.EgglayingProc` in `sys/egglaying.go`. Test: `TestEgglayingIndividualBased` (12 cohorts of 1 egg, stores reduced by 0.0015 g and 0.0093 kJ).

<a id="finding-44"></a>

### 44. A Psithyrus queen gets the thresholds of a social-phase host queen

- **Code:** `[MORA-02]`
- **Finding:** every evening, `Development_Mortality_AdultsProc` gives every mated queen in a colony with adult workers the thresholds `"QueenSocialPhase"`. The condition does not check the species, so a Psithyrus queen in a host colony with workers loses her `"Psith"` thresholds. Her egg-laying threshold drops from 0.2 to 0, and her nursing threshold from `NotSetHigh` to 0.9. The foraging thresholds stay `NotSetHigh`. A bee does not store which situation its thresholds came from, so only these values change. She stays recognisable as Psithyrus by `speciesName`.
- **Affects:** in `ActivityProc` the last matching task wins (finding 11, see trans/netlogo-findings.md#finding-11). While the host brood still needs incubation today (`StimNursingREP` = 1), nursing overwrites egg laying, and the Psithyrus rule in `ActivityProc` turns nursing into resting. The Psithyrus queen then rests for 30 minutes instead of laying eggs. She lays her batch only once the host brood has received today's incubation, and on a day on which it never does, she lays none. With the `"Psith"` thresholds she would lay eggs at the first opportunity. The change of the egg-laying threshold (0.2 to 0) has no effect, because the stimulus is 0 or 1.
- **Consequence for gobees:** translate as it is, without a species check. A test checks that a Psithyrus queen in a host colony with workers gets the social-phase thresholds. The effect on egg laying is tested with `ActivityProc` in step 5.
- **Second Check**: 24.10.2026
- **Implemented:** `Development_Mortality_AdultsProc` in `sys/development.go`, without species check. Test: `TestDevelopmentSocialPhaseThresholds` (host queen and Psithyrus queen get the social-phase thresholds, a queen without workers keeps hers). The effect on egg laying follows with `ActivityProc` in step 5. Step 5: `TestActivityProcChoice` (a Psithyrus queen with the social-phase thresholds rests while the host brood needs heat, and lays eggs once it is warm).

## Findings from step 5 (daily activity queue)

<a id="finding-45"></a>

### 45. The egg-laying stimulus counts all adults, not workers

- **Code:** `[STEG-01]`
- **Finding:** `StimEgglayingREP` is 1 if `(pollenStore_g > minPollenStore_g and allEggs + allLarvae = 0) or allAdults > 1`. The comment says `allAdults > 1` means "at least 1 worker is present". `allAdults` (`[UCOL-06]`) counts every adult with this `colonyID`: the queen herself, workers, males, young queens and a Psithyrus queen.
- **Affects:** colonies with adults other than the queen but no workers. (1) A colony with a Psithyrus queen: the host queen, still in the founding phase, has the egg-laying stimulus 1 every day and lays a batch on every day with enough stores and no brood that needs heat, instead of waiting until her brood has pupated. (2) Late in the season, when all workers have died but males or young queens are still in the nest: the queen keeps laying (after the switch point male eggs, after the competition point the eggs die in `[MORB-04]`).
- **Consequence for gobees:** translate as it is, with `allAdults > 1`. A test checks that a colony with the queen and one male has the stimulus 1.
- **Second Check**: 24.10.2026
- **Implemented:** `StimEgglayingREP` in `sys/stimuli.go`. Test: `TestStimuli` (queen and one male: stimulus 1).

<a id="finding-46"></a>

### 46. Eggs get no heat on the day they are laid

- **Code:** `[ACT-03]`, `[STNU-01]`, `[INC-02]`
- **Finding:** in `ActivityProc` nursing comes after egg laying and wins (finding 11, see trans/netlogo-findings.md#finding-11). The nursing thresholds of queens are 0.5 and 0.9, so a queen chooses egg laying only when the nursing stimulus of her colony is 0. The stimulus then stays 0 for the rest of the day: the need is computed from this morning's brood counts and is fixed for the day, the summed heat only grows. Nobody nurses after the egg laying, so the new eggs get no heat on that day. From the next day on they are in the counts and get their share.
- **Exception:** a Psithyrus queen with the `"Psith"` thresholds (nursing never) lays eggs also while the host brood still needs heat. Bees that nurse afterwards share their heat by this morning's brood count, but every brood agent, also the new eggs, adds the share. The colony then distributes more heat than it pays for.
- **Affects:** an egg reaches the hatching threshold `devIncubationHatchingTH_kJ` = quota × `devAgeHatchingMin_d` (`[SPEC-05]`, `[DEVE-01]`) one day after the minimum age if it gets exactly the quota every day. It hatches at the minimum age only if the daily surplus (`[INC-01]`, up to the heat of one turn) adds up to one quota. This fits the comment in `[QPD-01]`: the realised egg phase is about 6.4 days, with a minimum of 5. The worked examples in `[DEVE-01]`, `[DEVL-02]` and `[DEVL-03]` start with heat on the laying day, so they show the earliest possible days.
- **Consequence for gobees:** nothing to change, the effect follows from the order of the checks. A test with `ActivityProc` checks that eggs laid by a queen get no heat on their first day.
- **Second Check**: 24.10.2026
- **Implemented:** follows from `ActivityLoop.ActivityProc` and `BroodIncubationProc` in `sys/activity_loop.go`. Test: `TestActivityLoopDay` (one day of a colony with 12 eggs, a queen and 20 workers: the old eggs get at least the quota, the new batch gets 0 kJ).

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
- **Implemented:** `CreateLayersProc` in `sys/create_foodsources.go` (`RelativeAbundance` stores the value without `AbundanceBoost`). Test: `TestCreateLayers` (100 m2 x 2 x 0.01 g = 2 g pollen).

<a id="finding-11"></a>

### 11. In `ActivityProc` the last matching task wins

- **Code:** `[ACT-03]`
- **Finding:** The four comparisons of stimulus and threshold are separate consecutive `if` statements. Each matching condition overwrites `activity`. If several match, the order of precedence is therefore: nectar foraging before pollen foraging before nursing before egg laying. The comment in the code ("ordered by their importance") does not say whether the first or the last check takes precedence.
- **Consequence for gobees:** keep the exact order of the checks, do not write them as `if … else if`.
- **Second Check**: 23.10.2026, PP
- **Implemented:** `ActivityLoop.ActivityProc` in `sys/activity_loop.go`, four separate `if`. Test: `TestActivityProcChoice` (a founding queen with all stimuli 1 forages nectar).

<a id="finding-12"></a>

### 12. NetLogo `ticks` and ark-tools `resource.Tick` differ by 1

- **Code:** `[GO-03]`
- **Finding:** NetLogo calls `tick` at the start of `Go`. During the first simulated day `ticks` is 1, during `Setup` it is 0. ark-tools increments `resource.Tick` only after each update, so its value is 0 during the first day.
- **Consequence for gobees:** the system `sys.Tick` writes the NetLogo ticks into `globals.Ticks`. All systems read only this resource. The test `sys/tick_test.go` covers this.
- **Second Check**:

<a id="finding-22"></a>

### 22. Background mortality of adults has an inverted condition

- **Code:** `[MORA-03]`
- **Finding:** The condition is `MortalityAdultsBackground_daily > 0 and random-float 1 > MortalityAdultsBackground_daily`. A bee therefore dies with probability `1 - MortalityAdultsBackground_daily`, not with `MortalityAdultsBackground_daily`. `ParametersProc` sets the value to 0, so the first part of the condition is false and nobody dies from it.
- **Affects:** nothing with the hard-coded value 0. A value of e.g. 0.01 would kill 99 % of all adults (workers, males and all adult queens, including hibernating ones) every day.
- **Consequence for gobees:** translate the condition exactly as it is. A test with value 0 checks that no bee dies.
- **Second Check**: 23.10.2026, PP
- **Implemented:** `Development_Mortality_AdultsProc` in `sys/development.go`, condition as in NetLogo. Test: `TestDevelopmentBackgroundMortality` (0: nobody dies; 0.01: about 99 % die).
