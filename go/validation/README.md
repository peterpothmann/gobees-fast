# Validation against NetLogo

This folder holds the NetLogo side of the validation: the experiments, the scripts that run NetLogo, and the NetLogo results of 100 runs per experiment. `gobees-fast` is compared with them in two places:

- The experiments of the full model (`step7`, `step7_badgers`, `step7_crop`, `step8_*`) are scenarios of [go/compare](../compare/README.md), which checks every step against the previous step, gobees and NetLogo. It writes the report [validation-report.md](validation-report.md) in this folder in every pull request and commits it to the branch, so it reaches `main` with the squash merge (details in [Results](#results)).
- The experiments of parts of the model (`step3`, `step4a`, `step6`) are tests in `go/sys`, because they build a part of the model from its systems.

Each experiment measures only quantities that depend on the parts already translated, so a deviation points to the steps that cause it.

The comparison starts with step 3. Steps 0 to 2 produce no dynamics that could be compared over time: step 0 is the schedule, step 1 reads parameters, step 2 builds the landscape and the species. Their results are checked by unit tests against the input files, and the landscape and species values are also compared with NetLogo at step 0 of the step 3 experiment. So the first experiment covers steps 1 to 3 together. From step 3 on, every step gets its own experiment with the quantities that become comparable with it (see the overview below).

Method (as in [beecs](https://github.com/mlange-42/beecs)): NetLogo and `gobees` each run the same experiment 100 times. Single runs cannot be compared, because NetLogo's random number sequence cannot be reproduced in Go. Instead, the distributions over the 100 runs are compared, on **every single day**.

## How every day is checked

An experiment records a set of metrics (e.g. the number of colonies, workers, the energy stores) after the setup (day 0) and after every simulated day (tick). For every metric and every day:

1. NetLogo: the 100 values of this day give the quantiles 5, 25, 50, 75 and 95 %. The range from 5 % to 95 % is the NetLogo band of this day: 90 % of the NetLogo runs lie within it.
2. `gobees`: the median of its 100 values of the same day.
3. The day passes if the `gobees` median lies within the NetLogo band of that day.

A metric passes only if it passes on every day of the experiment, e.g. on all 365 days (0 to 364) of step 6. One day outside the band makes the test fail and names the day. The report therefore says "365 of 365 days" for every metric, and shows the values of one day as an example (step 6: day 150, the middle of the colony season; at the end of the year nearly all values are 0). The test log also names the day on which the `gobees` median is farthest from the NetLogo median, relative to the width of the band.

For values that depend only on the landscape (step 4a), all NetLogo runs agree up to the last digits of float sums, so the band is widened by a relative tolerance of 1e-9. go/compare uses the same tolerance for all metrics.

The scenarios of go/compare are checked from day 1 on: the program writes no row for the setup (day 0).

This band check is weak: in the comparison with the reference commit, a 10 % higher foraging mortality left the median inside the band on every day ([go/compare/README.md](../compare/README.md)). It stays as the check that is enforced, and the report adds a stronger one.

## Run summaries and Kolmogorov-Smirnov tests

`scripts/aggregate_netlogo.py` reduces every NetLogo run to a few numbers per metric (`reference/netlogo_<experiment>_runs.csv`, read by `ParseRuns` in `summary.go`): the value at day 0 (setup), and for every year of 365 days the maximum and the mean (year 1: days 1 to 365, year 2: 366 to 730, ...; a shorter last year uses the days it has).

For every yearly summary of a scenario, go/compare compares the 100 values of gobees-fast with the 100 values of NetLogo by a two-sample Kolmogorov-Smirnov test. The values are rounded to 9 significant digits first, so that float sums in a different order count as equal. All tests of one scenario together have a significance level of 0.05 (Bonferroni: each test uses 0.05 divided by the number of tests).

The KS tests against NetLogo are reported in the report, not enforced: gobees itself differs from NetLogo in some of them.

## Parametrisation of the experiments

All experiments use the example farm of BEE-STEWARD, i.e. the default settings of the NetLogo interface
(`MyMap` = `_SYSTEM_Example_Farm.png`, `MyParametersFile` = `_SYSTEM_Example_Farm_Parameters.csv`, finding 13) with
these input files from `netlogo/`:


| Input             | File                                                                                                                    |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------- |
| parameters        | `_SYSTEM_Example_Farm_Parameters.csv`                                                                                   |
| landscape         | `_SYSTEM_Example_Farm_Foodsources.txt`: 73 flower patches, 10 m per NetLogo patch, world 300 x 210 patches (3 x 2.1 km) |
| flower species    | `_SYSTEM_Flowerspecies.csv`                                                                                             |
| habitats          | `_SYSTEM_Habitats.csv`                                                                                                  |
| bumblebee species | `_SYSTEM_BumbleSpecies_UK_01.csv`                                                                                       |

The values of the parameter file that matter for the model dynamics (the others only change the display):


| Parameter                                                        | Value                                           | Meaning                                                         |
| ---------------------------------------------------------------- | ----------------------------------------------- | --------------------------------------------------------------- |
| `BeeSpeciesInitialQueensListAsString`                            | `"B_terrestris 500"`                            | 500 hibernating B. terrestris queens at the start               |
| `InitialQueensModifierBehaviorSpace`                             | 1                                               | factor on the number of initial queens                          |
| `COLONIES_IBM`                                                   | 0                                               | all colonies are cohort-based (no individual-based colonies)    |
| `Weather`                                                        | `"Constant 8 hrs"`                              | 8 hours of foraging every day                                   |
| `ForagingMortalityModel`, `ForagingMortalityFactor`              | `"high"`, 1                                     | foraging mortality 1e-5 per second                              |
| `Lambda_detectProb`                                              | -0.005                                          | detection probability of flower patches by distance             |
| `MaxForagingRange_m`                                             | 10000                                           |                                                                 |
| `MaxHibernatingQueens`                                           | 10000                                           | limit of hibernating queens on 1 January (not reached)          |
| `MergeHedges?`, `MinSizeFoodSources?`, `RemoveEmptyFoodSources?` | true, true, true                                | landscape preparation                                           |
| `FoodSourceLimit`, `AbundanceBoost`, `MaxPatchRadius_m`          | 25, 1, 500                                      |                                                                 |
| `MapAreaIncluded`                                                | `"complete"`                                    |                                                                 |
| `WinterMortality?`                                               | true                                            | queens can die in winter                                        |
| `UnlimitedMales?`                                                | true                                            | a queen without a male of her species mates with a foreign male |
| `SexLocus?`                                                      | false                                           | no diploid males                                                |
| `N_Badgers`                                                      | 0                                               | no badgers                                                      |
| `CropRotationListAsString`                                       | `""`                                            | no crop rotation                                                |
| `KeepDeadColonies?`, `CumulVisitsOnlyLastYear?`                  | true, false                                     | output only                                                     |
| `StopExtinct?`                                                   | true, set to false by the experiments of step 7 | a run that dies out records zeros instead of stopping           |

Deviations per experiment: `step7_badgers.xml` uses `experiments/step7_badgers_Parameters.csv`, the same file with
`N_Badgers` = 2. `step7_crop.xml` sets the crop rotation `_SYSTEM_Example_Farm_Foodsources.txt Example_Farm_PlotTest_Foodsources.txt` after Setup. The step 8 experiments use parameter files
written by `scripts/parameter_files.py`, each the example farm file with a few values changed:

| Experiment | Changed values | Days |
| --- | --- | --- |
| `step8_ibm.xml` | `COLONIES_IBM` = 10: the first 10 colonies are individual-based | 730 |
| `step8_genetics.xml` | `SexLocus?` = true, `UnlimitedMales?` = false | 730 |
| `step8_species.xml` | `BeeSpeciesInitialQueensListAsString` = all 7 species: 150 queens of each Bombus species, 100 Psithyrus | 730 |
| `step8_weather.xml` | `Weather` = `"foragingHoursExample"`, `ForagingMortalityModel` = `"low"` | 730 |
| `step8_settings.xml` | `WinterMortality?`, `MergeHedges?`, `MinSizeFoodSources?`, `RemoveEmptyFoodSources?`, `KeepDeadColonies?` = false, `MaxHibernatingQueens` = 150 (takes effect on every 1 January, also on the first day of the run) | 730 |

NetLogo reads these files through `MyParametersFile` (set in the experiment, with `ProjectsOwnParameterFile?` =
false), go/compare runs the scenarios with the same files (embedded by `go/validation/reference.go`).

Random numbers: NetLogo runs without explicit seed (`RAND_SEED` = 0), gobees uses the seeds 1 to 100. The gobees tests read the
same input files, copied into `go/internal/examplefiles`.

## Steps and their validation

The translation proceeds in steps. Each step is annotated in the NetLogo code, reviewed, translated with unit tests, and, where possible, compared with NetLogo.


| Step | NetLogo procedures                                                                                                                                                                                                                                                                                                                                                                             | Status  | Comparison with NetLogo                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| ---- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 0    | `Go`: the schedule                                                                                                                                                                                                                                                                                                                                                                             | done    | none, no dynamics.`sys.Tick` is unit-tested against NetLogo's `ticks`.                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| 1    | parameters and time:`SetupWithoutClearingProc`, `ReadAllParametersProc`, `ParametersProc`, `QueensInitialProc`, `ListFromStringREP`, `AllCSOlistREP`, `Foraging_periodREP`                                                                                                                                                                                                                     | done    | none, no dynamics. Unit tests with the parameter and habitat files.                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| 2    | landscape and species:`CreateFoodsourcesProc`, `MergeHedgesProc`, `CreateLayersProc`, `CreateSpeciesProc`, `NestHabitatsListREP`                                                                                                                                                                                                                                                               | done    | at step 0 of the step 3 experiment: foodsources, masterpatches, nest sites, nest-site area                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| 3    | queens and colony foundation:`CreateInitialQueensProc`, `EmergenceNewQueensProc`, `NestSitesSearchingProc`, `NestSiteFoodSourceREP`, `PsithyrusNestSearchProc`, `CreateColoniesProc`, `PatchesInRangeProc`, `FoodsourcesInFlowerAndRangeProc`, `DieProc` (bees) and the reporters they use                                                                                                     | done    | `experiments/step3.xml`: queen traits at step 0, colonies founded and queens searching or hibernating in year 1                                                                                                                                                                                                                                                                                                                                                                                                             |
| 4a   | morning routine: rest of`UpdateMorning_Proc`, `UpdateFoodsourcesProc`, `UpdateColoniesProc`, `CompetitionPointDateREP`, `CheckNumbersProc`                                                                                                                                                                                                                                                     | done    | `experiments/step4a.xml`: day of the year, nectar and pollen available in the landscape per day, 400 days. The colony quantities need brood and foraging (steps 4b to 6).                                                                                                                                                                                                                                                                                                                                                   |
| 4b   | brood without foraging:`EgglayingProc`, `EggsParameterSettingProc`, `DevelopmentProc` and sub-procedures, `DetermineCaste_REP`, `MortalityBroodProc`, `FeedLarvaeProc`, `NeedNectarPollenLarvaeTodayProc`, `MaxWeightGainToday_mg_REP`, `QueenProductionDateProc`                                                                                                                              | done    | compared together with step 6 (`experiments/step6.xml`)                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| 5    | daily activity queue:`NextActiveBeeREP`, `ActivityProc`, the stimulus reporters, `BroodIncubationProc`                                                                                                                                                                                                                                                                                         | done    | compared together with step 6 (`experiments/step6.xml`)                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| 6    | foraging:`ForagingProc` and all `Foraging_...` procedures, `HandlingTime_s_REP`, `DetectionProbREP`                                                                                                                                                                                                                                                                                            | done    | `experiments/step6.xml`: colony dynamics of year 1 (364 days) for steps 4b to 6 together: colonies, brood, adults, stores, foraging trips, switch point, queen production, nectar and pollen left in the landscape. NetLogo runs `Go` without `QueensLeavingNestProc` and `UpdateSeasonalEventsProc` (step 7).                                                                                                                                                                                                              |
| 7    | year cycle and population:`UpdateSeasonalEventsProc`, `QueensLeavingNestProc`, winter, badgers, Psithyrus, `DieProc` (colonies, badgers), `OutputDailyProc`, `CheckNumbersProc`, crop rotation                                                                                                                                                                                                 | done    | `experiments/step7.xml`: the full model (NetLogo `Go`, gobees `model.Default`) over 3 years (1095 days): colonies, dead and founded colonies, brood, adults, young and hibernating queens, stores, foraging trips, nectar and pollen left in the landscape. `experiments/step7_badgers.xml`: the same with 2 badgers. `experiments/step7_crop.xml`: the same with a crop rotation of two maps (badgers and crop rotation do not run on the example farm). All three: every metric within the NetLogo band on all 1096 days. |
| 8 | model branches that the example farm does not use, each in its own experiment: individual-based colonies (`COLONIES_IBM`), diploid males (`SexLocus?`), queens without males (`UnlimitedMales?` = false), `WinterMortality?` = false, `MergeHedges?` = false, other `ForagingMortalityModel` values, the limit of hibernating queens (`MaxHibernatingQueens`), several species incl. Psithyrus | done | `experiments/step8_ibm.xml`, `step8_genetics.xml`, `step8_species.xml`, `step8_weather.xml`, `step8_settings.xml`: the full model over 2 years (730 days) with the metrics of step 7 plus a few per experiment (individual-based colonies and bee agents, diploid males, colonies per species and Psithyrus, foraging period, foodsources and removed queens). All five: every metric within the NetLogo band on all 731 days. |
| 9 | plots: per metric the NetLogo band (5-95 %), the NetLogo median and the gobees median over the days, as images in`validation-report.md` | done | one image per experiment below its table in `validation-report.md` (`plots/<experiment>.png`), drawn by the comparison tests with gonum/plot |
| 10 | run time: NetLogo against gobees on the example farm (6 runs of one year, setup and days timed separately), in pull requests with the label `benchmark`. Not part of the validation, see [go/benchmark/README.md](../benchmark/README.md) | done | – |

## Files

Everything of the comparison is in this folder, `go/validation/`. It is inside the Go module, because Go can embed
only files from its own folder (`//go:embed`).

```
go/validation/
  README.md                   this file
  experiments/                BehaviorSpace experiments, kept outside the .nlogo so that the model file stays unchanged
    step3.xml                 steps 1 to 3: setup, emergence, winter mortality, nest search, colony foundation (250 days)
    step4a.xml                step 4a: day of the year, nectar and pollen in the landscape (400 days)
    step6.xml                 steps 4b to 6: colony dynamics of year 1 (364 days). Its go commands are NetLogo's Go
                              without the two procedures of step 7 (generated from the .nlogo, see the comment in the file)
    step7.xml                 step 7: the full model over 3 years (1095 days), go command NetLogo's Go
    step7_badgers.xml         step 7: as step7.xml with 2 badgers
    step7_badgers_Parameters.csv  the parameter file of step7_badgers.xml: the example farm with N_Badgers = 2
    step7_crop.xml            step 7: as step7.xml with a crop rotation of two maps
    step8_*.xml               step 8: model branches the example farm does not use (2 years each), with their
    step8_*_Parameters.csv    parameter files, written by scripts/parameter_files.py
  scripts/
    run_netlogo_experiment.ps1  runs an experiment headless with NetLogo 5.3.1
    aggregate_netlogo.py        turns the BehaviorSpace table into the quantiles per day and metric and the run summaries
    parameter_files.py          writes the parameter files of the step 8 experiments
  reference/                  NetLogo reference: netlogo_<experiment>.csv (step, metric, n, q05, q25, q50, q75, q95)
                              and netlogo_<experiment>_runs.csv (run, metric, period, kind, value)
  reference.go                embeds the reference files and the parameter files of the experiments
  validation.go               reading the reference, median
  summary.go                  reading the run summaries
  validation-report.md        the report of the scenarios, written by go/compare in the CI
  plots/                      one plot per scenario (<scenario>.png), written with the report
```

The comparison tests of steps 3 to 6 are in `go/sys`, because they build a part of the model from its systems:
`validation_step3_test.go` (about 2 s), `validation_step4a_test.go` (about 4 s), `validation_step6_test.go` (about
50 s). They run with every `go test` and are skipped with `go test -short`. The experiments of steps 7 and 8 run the
full model and are scenarios of [go/compare](../compare/README.md) (`scenario.go`), which runs them as a program.

## Running an experiment

From the repository root, in PowerShell:

```powershell
.\go\validation\scripts\run_netlogo_experiment.ps1 -Experiment go\validation\experiments\step3.xml -Output step3_netlogo.csv
python go/validation/scripts/aggregate_netlogo.py step3_netlogo.csv step3
cd go; go test -count=1 -timeout 30m -run Validation -v ./sys
```

For an experiment of the full model, the last line is `go/compare/run_compare.sh` (Git Bash) instead.

The NetLogo table (`step3_netlogo.csv`) is not needed afterwards and is excluded by `.gitignore`. Commit the generated `go/validation/reference/netlogo_step3.csv`.

Notes on the NetLogo command line:

- NetLogo 5.3.1 is installed without `java.exe`. The script uses the Java on the PATH; Java 21 works.
- NetLogo finds its extensions (`profiler`, `csv`) only if the working directory is its `app` folder. The script changes into it.
- The classpath must be the one in `app\NetLogo.cfg`. Adding all jars of the folder fails, because `NetLogoLite.jar` contains conflicting classes.
- With 20 threads, 100 runs of 250 days take about 6 minutes (one run about 17 s).

## Results

The results are not written down by hand. The CI (workflow "Tests", `.github/workflows/tests.yml`) runs on every
pull request and on every push to `main`, in two jobs:

1. `test` runs all tests of gobees-fast, incl. the experiments of parts of the model (`go/sys`). A metric outside
   the NetLogo band fails the test.
2. `compare` runs [go/compare](../compare/README.md): every scenario against the previous step, gobees and NetLogo.
   It writes one comment with a table, one row per scenario and one column per comparison, to the pull request
   (updated on every run) and to the summary of the run. It writes [validation-report.md](validation-report.md)
   with one plot per scenario (`plots/<scenario>.png`: the NetLogo band and median and the gobees-fast median over
   the days) and the Kolmogorov-Smirnov tests against NetLogo. In a pull request, the job commits report and plots
   to the branch if all comparisons passed and something changed ("Update validation report and plots", made by
   github-actions). Pull the branch before you push to it again. On `main` it never commits.

The runs use fixed seeds (1 to 100), so the report changes only when the results change.

What each experiment measures and why is described in the experiment file (e.g. `experiments/step3.xml`), in the
comparison test (e.g. `go/sys/validation_step3_test.go`) and in the scenario (`go/compare/scenario.go`).

To create the report locally (not needed, the CI does it), run go/compare with `COMPARE_REPORT_DIR` set. A report in
this folder replaces the committed one; `git checkout validation/validation-report.md validation/plots` restores it.

```bash
COMPARE_REPORT_DIR="$PWD/go/validation" go/compare/run_compare.sh
```
