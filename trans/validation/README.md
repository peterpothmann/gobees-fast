# Validation against NetLogo

gobees is compared with the original NetLogo model as early as possible. Each experiment measures only quantities that depend on the parts already translated, so a deviation points to the steps that cause it.

The comparison starts with step 3. Steps 0 to 2 produce no dynamics that could be compared over time: step 0 is the schedule, step 1 reads parameters, step 2 builds the landscape and the species. Their results are checked by unit tests against the input files, and the landscape and species values are also compared with NetLogo at step 0 of the step 3 experiment. So the first experiment covers steps 1 to 3 together. From step 3 on, every step gets its own experiment with the quantities that become comparable with it (see the overview below).

Method (as in [beecs](https://https://github.com/mlange-42/beecs)): 100 NetLogo runs with BehaviorSpace give quantiles (5, 25, 50, 75, 95 %) per day and metric. A Go test runs gobees 100 times and requires the gobees median to lie within the NetLogo 5–95 % band. Single runs cannot be csompared, because NetLogo's random number sequence cannot be reproduced in Go.

## Steps and their validation

The translation proceeds in steps. Each step is annotated in the NetLogo code, reviewed, translated with unit tests, and, where possible, compared with NetLogo.


| Step | NetLogo procedures                                                                                                                                                                                                                                                                         | Status                                                  | Comparison with NetLogo                                                                                                                                                                                         |
| ---- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 0    | `Go`: the schedule                                                                                                                                                                                                                                                                         | done                                                    | none, no dynamics.`sys.Tick` is unit-tested against NetLogo's `ticks`.                                                                                                                                          |
| 1    | parameters and time:`SetupWithoutClearingProc`, `ReadAllParametersProc`, `ParametersProc`, `QueensInitialProc`, `ListFromStringREP`, `AllCSOlistREP`, `Foraging_periodREP`                                                                                                                 | done                                                    | none, no dynamics. Unit tests with the parameter and habitat files.                                                                                                                                             |
| 2    | landscape and species:`CreateFoodsourcesProc`, `MergeHedgesProc`, `CreateLayersProc`, `CreateSpeciesProc`, `NestHabitatsListREP`                                                                                                                                                           | done                                                    | at step 0 of the step 3 experiment: foodsources, masterpatches, nest sites, nest-site area                                                                                                                      |
| 3    | queens and colony foundation:`CreateInitialQueensProc`, `EmergenceNewQueensProc`, `NestSitesSearchingProc`, `NestSiteFoodSourceREP`, `PsithyrusNestSearchProc`, `CreateColoniesProc`, `PatchesInRangeProc`, `FoodsourcesInFlowerAndRangeProc`, `DieProc` (bees) and the reporters they use | done                                                    | `step3_experiment.xml`: queen traits at step 0, colonies founded and queens searching or hibernating in year 1                                                                                                  |
| 4    | brood without foraging: rest of`UpdateMorning_Proc`, `UpdateFoodsourcesProc`, `UpdateColoniesProc`, `EgglayingProc`, `DevelopmentProc` and sub-procedures, `MortalityBroodProc`, `FeedLarvaeProc`, `NeedNectarPollenLarvaeTodayProc`, `QueenProductionDateProc`                            | open                                                    | not yet decided. Brood and stores depend on foraging, so most colony quantities become comparable only with step 6. Candidates: nectar and pollen available in the landscape per day (`UpdateFoodsourcesProc`). |
| 5    | daily activity queue:`NextActiveBeeREP`, `ActivityProc`, the stimulus reporters, `BroodIncubationProc`                                                                                                                                                                                     | open                                                    | not yet decided, depends on foraging as well                                                                                                                                                                    |
| 6    | foraging:`ForagingProc` and all `Foraging_...` procedures, `HandlingTime_s_REP`, `DetectionProbREP`                                                                                                                                                                                        | open                                                    | colony dynamics of year 1: colony size, brood, stores, workers, foraging trips, young queens produced                                                                                                           |
| 7    | year cycle and population:`UpdateSeasonalEventsProc`, `QueensLeavingNestProc`, winter, badgers, Psithyrus, `DieProc` (colonies, badgers), `OutputDailyProc`, `CheckNumbersProc`, crop rotation                                                                                             | open                                                    | several years: colonies per year, hibernating queens, extinction                                                                                                                                                |
| 8    | full validation, benchmarks, CI                                                                                                                                                                                                                                                            | CI for tests is running (`.github/workflows/tests.yml`) | all outputs of the full model over several years                                                                                                                                                                |

## Files


| File                              | Purpose                                                                                                                                                                                    |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `step3_experiment.xml`            | BehaviorSpace experiment for steps 1 to 3 together (setup, emergence, winter mortality, nest search, colony foundation). Kept outside the`.nlogo`, so that the model file stays unchanged. |
| `run_netlogo_experiment.ps1`      | Runs an experiment headless with NetLogo 5.3.1.                                                                                                                                            |
| `aggregate_netlogo.py`            | Turns the BehaviorSpace table into quantiles and writes them into`go/internal/validation/netlogo_<name>.go`. A Go file and not a CSV file, because `.gitignore` excludes `*.csv`.          |
| `go/sys/validation_step3_test.go` | Compares gobees with the quantiles. Runs with every`go test` (about 2 s), skipped with `go test -short`.                                                                                   |

## Running an experiment

From the repository root, in PowerShell:

```powershell
.\trans\validation\run_netlogo_experiment.ps1 -Experiment trans\validation\step3_experiment.xml -Output step3_netlogo.csv
python trans/validation/aggregate_netlogo.py step3_netlogo.csv step3
cd go; go test ./sys -run Validation -v
```

The NetLogo table (`step3_netlogo.csv`) is not needed afterwards and is excluded by `.gitignore`. Commit the generated Go file.

Notes on the NetLogo command line:

- NetLogo 5.3.1 is installed without `java.exe`. The script uses the Java on the PATH; Java 21 works.
- NetLogo finds its extensions (`profiler`, `csv`) only if the working directory is its `app` folder. The script changes into it.
- The classpath must be the one in `app\NetLogo.cfg`. Adding all jars of the folder fails, because `NetLogoLite.jar` contains conflicting classes.
- With 20 threads, 100 runs of 250 days take about 6 minutes (one run about 17 s).

## Results

The results are not written down here, they are generated on every push. The summary of each run of the
workflow "Tests" on GitHub (Actions, then the run) shows

- all test results, and
- the comparison report: for every metric the gobees and NetLogo medians, the NetLogo 5–95 % band, and at how many
  steps gobees lies within the band.

Both are also attached to the run as artifact `reports`. What each experiment measures and why is described in the
experiment file (e.g. `step3_experiment.xml`) and in the comparison test (e.g. `go/sys/validation_step3_test.go`).

To create the comparison report locally, set `VALIDATION_REPORT` to an absolute path (`go test` runs the test in
the package folder `go/sys`, so a relative path would end up there).

PowerShell:

```powershell
cd go
$env:VALIDATION_REPORT = "$PWD\..\trans\validation\validation-report.md"
Remove-Item $env:VALIDATION_REPORT -ErrorAction SilentlyContinue
go test -count=1 -run Validation ./sys
Remove-Item Env:VALIDATION_REPORT
```

Git Bash:

```bash
cd go
rm -f ../trans/validation/validation-report.md
VALIDATION_REPORT="$PWD/../trans/validation/validation-report.md" go test -count=1 -run Validation ./sys
```

The report is written to `trans/validation/validation-report.md`, next to the experiments (open it in VS Code with
Ctrl+Shift+V for the preview). `.gitignore` excludes it. The
tests append to the file, because several comparison tests write into one report, so the commands delete the old
report first.
