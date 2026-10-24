# Validation against NetLogo

gobees is compared with the original NetLogo model as early as possible. Each experiment measures only quantities that depend on the parts already translated, so a deviation points to the steps that cause it.

The comparison starts with step 3. Steps 0 to 2 produce no dynamics that could be compared over time: step 0 is the schedule, step 1 reads parameters, step 2 builds the landscape and the species. Their results are checked by unit tests against the input files, and the landscape and species values are also compared with NetLogo at step 0 of the step 3 experiment. So the first experiment covers steps 1 to 3 together. From step 3 on, every step gets its own experiment with the quantities that become comparable with it (see the overview below).

Method (as in [beecs](https://github.com/mlange-42/beecs)): NetLogo and gobees each run the same experiment 100 times. Single runs cannot be compared, because NetLogo's random number sequence cannot be reproduced in Go. Instead, the distributions over the 100 runs are compared, on every single day.

## How every day is checked

An experiment records a set of metrics (e.g. the number of colonies, workers, the energy stores) after the setup (day 0) and after every simulated day (tick). For every metric and every day:

1. NetLogo: the 100 values of this day give the quantiles 5, 25, 50, 75 and 95 %. The range from 5 % to 95 % is the NetLogo band of this day: 90 % of the NetLogo runs lie within it.
2. gobees: the median of its 100 values of the same day.
3. The day passes if the gobees median lies within the NetLogo band of that day.

A metric passes only if it passes on every day of the experiment, e.g. on all 365 days (0 to 364) of step 6. One day outside the band makes the test fail and names the day. The report therefore says "365 of 365 days" for every metric, and shows the values of one day as an example (step 6: day 150, the middle of the colony season; at the end of the year nearly all values are 0). The test log also names the day on which the gobees median is farthest from the NetLogo median, relative to the width of the band.

For values that depend only on the landscape (step 4a), all NetLogo runs agree up to the last digits of float sums, so the band is widened by a relative tolerance of 1e-9.

## Steps and their validation

The translation proceeds in steps. Each step is annotated in the NetLogo code, reviewed, translated with unit tests, and, where possible, compared with NetLogo.


| Step | NetLogo procedures                                                                                                                                                                                                                                                                         | Status                                                  | Comparison with NetLogo                                                                                                                                                           |
| ---- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 0    | `Go`: the schedule                                                                                                                                                                                                                                                                         | done                                                    | none, no dynamics.`sys.Tick` is unit-tested against NetLogo's `ticks`.                                                                                                            |
| 1    | parameters and time:`SetupWithoutClearingProc`, `ReadAllParametersProc`, `ParametersProc`, `QueensInitialProc`, `ListFromStringREP`, `AllCSOlistREP`, `Foraging_periodREP`                                                                                                                 | done                                                    | none, no dynamics. Unit tests with the parameter and habitat files.                                                                                                               |
| 2    | landscape and species:`CreateFoodsourcesProc`, `MergeHedgesProc`, `CreateLayersProc`, `CreateSpeciesProc`, `NestHabitatsListREP`                                                                                                                                                           | done                                                    | at step 0 of the step 3 experiment: foodsources, masterpatches, nest sites, nest-site area                                                                                        |
| 3    | queens and colony foundation:`CreateInitialQueensProc`, `EmergenceNewQueensProc`, `NestSitesSearchingProc`, `NestSiteFoodSourceREP`, `PsithyrusNestSearchProc`, `CreateColoniesProc`, `PatchesInRangeProc`, `FoodsourcesInFlowerAndRangeProc`, `DieProc` (bees) and the reporters they use | done                                                    | `experiments/step3.xml`: queen traits at step 0, colonies founded and queens searching or hibernating in year 1                                                                    |
| 4a   | morning routine: rest of`UpdateMorning_Proc`, `UpdateFoodsourcesProc`, `UpdateColoniesProc`, `CompetitionPointDateREP`, `CheckNumbersProc`                                                                                                                                                 | done                                                    | `experiments/step4a.xml`: day of the year, nectar and pollen available in the landscape per day, 400 days. The colony quantities need brood and foraging (steps 4b to 6).          |
| 4b   | brood without foraging:`EgglayingProc`, `EggsParameterSettingProc`, `DevelopmentProc` and sub-procedures, `DetermineCaste_REP`, `MortalityBroodProc`, `FeedLarvaeProc`, `NeedNectarPollenLarvaeTodayProc`, `MaxWeightGainToday_mg_REP`, `QueenProductionDateProc`                          | done                                                    | compared together with step 6 (`experiments/step6.xml`) |
| 5    | daily activity queue:`NextActiveBeeREP`, `ActivityProc`, the stimulus reporters, `BroodIncubationProc`                                                                                                                                                                                     | done                                                    | compared together with step 6 (`experiments/step6.xml`) |
| 6    | foraging:`ForagingProc` and all `Foraging_...` procedures, `HandlingTime_s_REP`, `DetectionProbREP`                                                                                                                                                                                        | done                                                    | `experiments/step6.xml`: colony dynamics of year 1 (364 days) for steps 4b to 6 together: colonies, brood, adults, stores, foraging trips, switch point, queen production, nectar and pollen left in the landscape. NetLogo runs `Go` without `QueensLeavingNestProc` and `UpdateSeasonalEventsProc` (step 7). |
| 7    | year cycle and population:`UpdateSeasonalEventsProc`, `QueensLeavingNestProc`, winter, badgers, Psithyrus, `DieProc` (colonies, badgers), `OutputDailyProc`, `CheckNumbersProc`, crop rotation                                                                                             | NetLogo annotated, in review                            | several years: colonies per year, hibernating queens, extinction                                                                                                                  |
| 8    | full validation, benchmarks, CI                                                                                                                                                                                                                                                            | CI for tests is running (`.github/workflows/tests.yml`) | all outputs of the full model over several years                                                                                                                                  |

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
  scripts/
    run_netlogo_experiment.ps1  runs an experiment headless with NetLogo 5.3.1
    aggregate_netlogo.py        turns the BehaviorSpace table into the quantiles per day and metric
  reference/                  NetLogo reference: netlogo_<experiment>.csv (step, metric, n, q05, q25, q50, q75, q95)
  reference.go                embeds the reference files
  validation.go, report.go    reading the reference, median, report
```

The comparison tests are in `go/sys`, because they build the model from its systems:
`validation_step3_test.go` (about 2 s), `validation_step4a_test.go` (about 4 s), `validation_step6_test.go` (about
50 s). They run with every `go test` and are skipped with `go test -short`.

## Running an experiment

From the repository root, in PowerShell:

```powershell
.\go\validation\scripts\run_netlogo_experiment.ps1 -Experiment go\validation\experiments\step3.xml -Output step3_netlogo.csv
python go/validation/scripts/aggregate_netlogo.py step3_netlogo.csv step3
cd go; go test ./sys -run Validation -v
```

The NetLogo table (`step3_netlogo.csv`) is not needed afterwards and is excluded by `.gitignore`. Commit the generated `go/validation/reference/netlogo_step3.csv`.

Notes on the NetLogo command line:

- NetLogo 5.3.1 is installed without `java.exe`. The script uses the Java on the PATH; Java 21 works.
- NetLogo finds its extensions (`profiler`, `csv`) only if the working directory is its `app` folder. The script changes into it.
- The classpath must be the one in `app\NetLogo.cfg`. Adding all jars of the folder fails, because `NetLogoLite.jar` contains conflicting classes.
- With 20 threads, 100 runs of 250 days take about 6 minutes (one run about 17 s).

## Results

The results are not written down here, they are generated on every push. The summary of each run of the
workflow "Tests" on GitHub (Actions, then the run) shows

- all test results, and
- the comparison report: for every metric, on how many days the gobees median lies within the NetLogo band (see
  "How every day is checked"), and the values of one example day.

Both are also attached to the run as artifact `reports`. What each experiment measures and why is described in the
experiment file (e.g. `experiments/step3.xml`) and in the comparison test (e.g. `go/sys/validation_step3_test.go`).

To create the comparison report locally, set `VALIDATION_REPORT` to an absolute path (`go test` runs the test in
the package folder `go/sys`, so a relative path would end up there).

PowerShell:

```powershell
cd go
$env:VALIDATION_REPORT = "$PWD\validation\validation-report.md"
Remove-Item $env:VALIDATION_REPORT -ErrorAction SilentlyContinue
go test -count=1 -run Validation ./sys
Remove-Item Env:VALIDATION_REPORT
```

Git Bash:

```bash
cd go
rm -f validation/validation-report.md
VALIDATION_REPORT="$PWD/validation/validation-report.md" go test -count=1 -run Validation ./sys
```

The report is written to `go/validation/validation-report.md` (open it in VS Code with
Ctrl+Shift+V for the preview). `.gitignore` excludes it. The
tests append to the file, because several comparison tests write into one report, so the commands delete the old
report first.
