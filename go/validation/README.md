# Validation against NetLogo

`gobees` is compared with the original NetLogo model as early as possible. The final report is [validation-report.md](validation-report.md) in this folder. The CI writes it in every pull request and commits it to the branch of the pull request, so it reaches `main` with the squash merge (details in [Results](#results)). Each experiment measures only quantities that depend on the parts already translated, so a deviation points to the steps that cause it.

The comparison starts with step 3. Steps 0 to 2 produce no dynamics that could be compared over time: step 0 is the schedule, step 1 reads parameters, step 2 builds the landscape and the species. Their results are checked by unit tests against the input files, and the landscape and species values are also compared with NetLogo at step 0 of the step 3 experiment. So the first experiment covers steps 1 to 3 together. From step 3 on, every step gets its own experiment with the quantities that become comparable with it (see the overview below).

Method (as in [beecs](https://github.com/mlange-42/beecs)): NetLogo and `gobees` each run the same experiment 100 times. Single runs cannot be compared, because NetLogo's random number sequence cannot be reproduced in Go. Instead, the distributions over the 100 runs are compared, on **every single day**.

## How every day is checked

An experiment records a set of metrics (e.g. the number of colonies, workers, the energy stores) after the setup (day 0) and after every simulated day (tick). For every metric and every day:

1. NetLogo: the 100 values of this day give the quantiles 5, 25, 50, 75 and 95 %. The range from 5 % to 95 % is the NetLogo band of this day: 90 % of the NetLogo runs lie within it.
2. `gobees`: the median of its 100 values of the same day.
3. The day passes if the `gobees` median lies within the NetLogo band of that day.

A metric passes only if it passes on every day of the experiment, e.g. on all 365 days (0 to 364) of step 6. One day outside the band makes the test fail and names the day. The report therefore says "365 of 365 days" for every metric, and shows the values of one day as an example (step 6: day 150, the middle of the colony season; at the end of the year nearly all values are 0). The test log also names the day on which the `gobees` median is farthest from the NetLogo median, relative to the width of the band.

For values that depend only on the landscape (step 4a), all NetLogo runs agree up to the last digits of float sums, so the band is widened by a relative tolerance of 1e-9.

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
`N_Badgers` = 2. `step7_crop.xml` sets the crop rotation `_SYSTEM_Example_Farm_Foodsources.txt Example_Farm_PlotTest_Foodsources.txt` after Setup. Step 8 will add experiments for the other values (see the table
below).

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
| 8    | model branches that the example farm does not use, each in its own experiment: individual-based colonies (`COLONIES_IBM`), diploid males (`SexLocus?`), queens without males (`UnlimitedMales?` = false), `WinterMortality?` = false, `MergeHedges?` = false, other `ForagingMortalityModel` values, the limit of hibernating queens (`MaxHibernatingQueens`), several species incl. Psithyrus | planned | one experiment per group of branches, as in step 7                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| 9    | plots: per metric the NetLogo band (5-95 %), the NetLogo median and the gobees median over the days, as images in`validation-report.md`                                                                                                                                                                                                                                                        | planned | –                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| 10   | benchmarks: runtime of gobees (CI, benchstat) and of NetLogo against gobees, e.g. for different numbers of initial queens. Not part of the validation, see the README in the repository root                                                                                                                                                                                                   | planned | –                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |

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
  scripts/
    run_netlogo_experiment.ps1  runs an experiment headless with NetLogo 5.3.1
    aggregate_netlogo.py        turns the BehaviorSpace table into the quantiles per day and metric
  reference/                  NetLogo reference: netlogo_<experiment>.csv (step, metric, n, q05, q25, q50, q75, q95)
  reference.go                embeds the reference files
  validation.go, report.go    reading the reference, median, report
```

The comparison tests of steps 3 to 6 are in `go/sys`, because they build a part of the model from its systems:
`validation_step3_test.go` (about 2 s), `validation_step4a_test.go` (about 4 s), `validation_step6_test.go` (about
50 s). The tests of step 7 compare the full model and are in `go/model`, because they use `model.Default`:
`validation_step7_test.go` and `validation_step7_badgers_crop_test.go` (about 60 s per experiment), with the shared comparison in
`validation_test.go`. They run with every `go test` and are skipped with `go test -short`.

## Running an experiment

From the repository root, in PowerShell:

```powershell
.\go\validation\scripts\run_netlogo_experiment.ps1 -Experiment go\validation\experiments\step3.xml -Output step3_netlogo.csv
python go/validation/scripts/aggregate_netlogo.py step3_netlogo.csv step3
cd go; go test -count=1 -run Validation -v ./sys ./model
```

The NetLogo table (`step3_netlogo.csv`) is not needed afterwards and is excluded by `.gitignore`. Commit the generated `go/validation/reference/netlogo_step3.csv`.

Notes on the NetLogo command line:

- NetLogo 5.3.1 is installed without `java.exe`. The script uses the Java on the PATH; Java 21 works.
- NetLogo finds its extensions (`profiler`, `csv`) only if the working directory is its `app` folder. The script changes into it.
- The classpath must be the one in `app\NetLogo.cfg`. Adding all jars of the folder fails, because `NetLogoLite.jar` contains conflicting classes.
- With 20 threads, 100 runs of 250 days take about 6 minutes (one run about 17 s).

## Results

The results are not written down by hand. The CI (workflow "Tests", `.github/workflows/tests.yml`) runs all tests
on every pull request and on every push to `main`. The comparison tests write the report
[validation-report.md](validation-report.md): for every metric, on how many days the gobees median lies within the
NetLogo band (see "How every day is checked"), and the values of one example day.

- In a pull request, the CI commits the report to the branch of the pull request ("Update validation report", made
  by github-actions), but only if all tests passed and the report changed. Pull the branch before you push to it
  again. With a squash merge, code and report reach `main` in one commit.
- On `main`, the CI only runs the tests and never commits.
- The summary of each run (Actions, then the run) shows all test results and the report, which are also attached to
  the run as artifact `reports`.

The gobees runs use fixed seeds (1 to 100), so the report changes only when the results change.

What each experiment measures and why is described in the experiment file (e.g. `experiments/step3.xml`) and in the
comparison test (e.g. `go/sys/validation_step3_test.go`).

To create the comparison report locally (not needed, the CI does it), set `VALIDATION_REPORT` to an absolute path
(`go test` runs the test in the package folder, e.g. `go/sys`, so a relative path would end up there). Each test
adds its section to the file or replaces its old section, and writes the report sorted by step
(`validation.Section.Order`): the packages `go/sys` and `go/model` run at the same time and finish in any order. A
lock file (`validation-report.md.lock`) keeps them from writing at the same time; a test that is killed while
writing can leave it behind, then delete it. The commands delete the old report first, so that it holds only the
sections of this run. A local report replaces the committed one; `git checkout validation/validation-report.md`
restores it.

PowerShell:

```powershell
cd go
$env:VALIDATION_REPORT = "$PWD\validation\validation-report.md"
Remove-Item $env:VALIDATION_REPORT -ErrorAction SilentlyContinue
go test -count=1 -run Validation ./sys ./model
Remove-Item Env:VALIDATION_REPORT
```

Git Bash:

```bash
cd go
rm -f validation/validation-report.md
VALIDATION_REPORT="$PWD/validation/validation-report.md" go test -count=1 -run Validation ./sys ./model
```
