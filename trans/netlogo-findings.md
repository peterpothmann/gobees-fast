# Findings in the BEE-STEWARD NetLogo code

This file collects places in the NetLogo code that are easily lost when translating to Go, or that would change the model dynamics. Principle for gobees: the NetLogo code is the reference. Quirks are kept and documented, not fixed. Someone could fix them in another model version, though.

Source: `netlogo/__BEE-STEWARD_RC_ColourVersionMB.nlogo`. Codes in square brackets (e.g. `[FPER-02]`) refer to the annotations in the NetLogo code. Line numbers are deliberately omitted because they shift as comments are added.

## Findings from step 1 (setup, parameters, time)

### 1. The weather list has 366 instead of 365 items

- **Code:** `[FPER-02]`, `Foraging_periodREP`
- **Finding:** The list `foragingHoursExample` contains the items `0` and `.5` at positions 41 and 42 (0-based). `0.5` was probably intended. As a result, the list has 366 items. The model reads item `Day - 1`. Compared to the probably intended list, day 42 gets 0 instead of 0.5, and from day 43 on each day gets the value intended for the previous day. Item 366 is never read.
- **Affects:** only `Weather = "foragingHoursExample"`. The default `"Constant 8 hrs"` is not affected.
- **Consequence for gobees:** copy the list exactly, with 366 items.
- **Second Check**: 23-10-2026, PP
- **Implemented:** `util/foraging_period.go`, list generated token by token from the NetLogo code. Test: `util/foraging_period_test.go` (366 items, day 42 = 0 h, day 43 = 0.5 h).

### 2. `Habitats.csv` is needed for a simulation run

- **Code:** `[PAR-06]`, `[CSO]`, `AllCSOlistREP`
- **Finding:** During setup, `ParametersProc` calls `AllCSOlistREP`. The reporter reads the `HabitatsFile` and collects the `habitatType` of all rows whose `habitatColourID` contains the text `FlowerSpeciesList_`. For `_SYSTEM_Habitats.csv` the result is `["Legume" "Margin" "newCSO_Legume" "BirdBumble_Margin" "GameCover_Plot" "Plot"]`. `CreateLayersProc` needs this list to recognise entries like `["Margin"]` in a `flowerSpeciesList`. Rows with a numeric `habitatColourID` are skipped silently by `carefully`.
- **Consequence for gobees:** the `HabitatsFile` is an input file of every run, not only of map processing.
- **Second Check**: 23-10-2026, PP
- **Implemented:** `allCSOList` in `sys/parameters.go`. Test: `sys/parameters_test.go` (the 6 names for the default files).

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

### 4. Setup runs at `Day = 0`

- **Code:** `[INIT-10]`
- **Finding:** `SetupWithoutClearingProc` calls `UpdateMorning_Proc` at `ticks = 0`, i.e. at `Day = 0`. `Foraging_periodREP` is only called for `Day > 0`. `DailyForagingPeriod_s` therefore stays 0 until day 1.
- **Consequence for gobees:** the morning routine must run during setup with `Day = 0` and keep the condition `Day > 0`.
- **Second Check**: 23-10-2026, PP

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

### 6. The parameter file and `ParametersProc` share no names

- **Code:** `[INIT-05]`, `[PAR-03]`
- **Finding:** No variable set in `ParametersProc` (37 variables) also appears in `_SYSTEM_Parameters.csv` (47 parameters). `ParametersProc` therefore overwrites no value from the file, although it runs afterwards.
- **Check:** `python trans/check_param_overlap.py` (from the repository root). Rerun it when the parameter file or `ParametersProc` changes, e.g. for a project's own parameter file: `--params netlogo/<map>_Parameters.csv`.
- **Consequence for gobees:** the hard-coded values can be default parameters next to the file parameters. Their order does not matter.
- **Second Check**: 23-10-2026, PP

### 7. `Date` is for display only

- **Code:** `[DATE]`
- **Finding:** The global `Date` is set in `UpdateMorning_Proc` and only shown in the interface. No model procedure reads it. The model works with `Day`, which `UpdateMorning_Proc` computes separately.
- **Consequence for gobees:** `DateREP` is not translated.
- **Second Check**: 23-10-2026, PP

### 13. With the default interface settings, NetLogo reads `_SYSTEM_Example_Farm_Parameters.csv`

- **Code:** `[INIT-02]`, `[RPAR-01]`
- **Finding:** The interface stores `MyMap = "_SYSTEM_Example_Farm.png"` and `MyParametersFile = "_SYSTEM_Example_Farm_Parameters.csv"`. For exactly this combination, `[INIT-02]` sets `ProjectsOwnParameterFile?` to false, so `ReadAllParametersProc` loads `MyParametersFile`. A default run therefore uses `_SYSTEM_Example_Farm_Parameters.csv`, not `_SYSTEM_Parameters.csv`. The two files differ only in `SetColour` (`"Magenta"` vs. `"Yellow"`, display only) and in the spelling `true`/`TRUE`. For the model they are equal.
- **Consequence for gobees:** `main.go` reads `_SYSTEM_Example_Farm_Parameters.csv`, like a default NetLogo run.
- **Second Check**: 23-10-2026, PP
- **Implemented:** `main.go`. Test: `params/parameter_file_test.go` reads both files.

## Findings from the model map and from `Go`

### 8. The foraging range from the parameter file is not used

- **Code:** `[PAR-03]`
- **Finding:** `MaxForagingRange_m` is in `_SYSTEM_Parameters.csv` (value 10000), but no procedure reads it. The model uses `ForagingRangeMax_m`, which `ParametersProc` sets to a fixed 758 m.
- **Consequence for gobees:** keep 758 m as a fixed value. Read `MaxForagingRange_m` from the file, but do not use it.
- **Second Check**:

### 9. `deadCols` loses `speciesNameColony`

- **Code:** not yet annotated (`DieProc`)
- **Finding:** Dead colonies change their breed in `DieProc` via `set breed deadCols`. `colonies-own` has the variable `speciesNameColony`, `deadCols-own` does not. The comment at the declaration demands an exact copy. The value is lost when the colony dies.
- **Consequence for gobees:** relevant for output only. When annotating `DieProc`, check whether any output reads the value for dead colonies.
- **Second Check**:

### 10. The value in the flower list acts as flowers per m²

- **Code:** not yet annotated (`CreateLayersProc`)
- **Finding:** The code calls the value from `flowerSpeciesList` "relative abundance". It is, however, multiplied by the area and by pollen or nectar *per flower*: `pollenMax_g = area_sqm × AbundanceBoost × value × pollen_g/flower`. Its unit is therefore flowers per m².
- **Status:** my own reading, to be confirmed when annotating `CreateLayersProc`.
- **Second Check**:

### 11. In `ActivityProc` the last matching task wins

- **Code:** not yet annotated (`ActivityProc`)
- **Finding:** The four comparisons of stimulus and threshold are separate consecutive `if` statements. Each matching condition overwrites `activity`. If several match, the order of precedence is therefore: nectar foraging before pollen foraging before nursing before egg laying. The comment in the code ("ordered by their importance") does not say whether the first or the last check takes precedence.
- **Consequence for gobees:** keep the exact order of the checks, do not write them as `if … else if`.
- **Second Check**:

### 12. NetLogo `ticks` and ark-tools `resource.Tick` differ by 1

- **Code:** `[GO-03]`
- **Finding:** NetLogo calls `tick` at the start of `Go`. During the first simulated day `ticks` is 1, during `Setup` it is 0. ark-tools increments `resource.Tick` only after each update, so its value is 0 during the first day.
- **Consequence for gobees:** the system `sys.Tick` writes the NetLogo ticks into `globals.Ticks`. All systems read only this resource. The test `sys/tick_test.go` covers this.
- **Second Check**:
