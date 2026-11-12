# Check of a step

gobees-fast starts as a copy of gobees and is changed in small steps. This folder checks every step: it runs the
scenarios of the full model with the current code and compares it with three versions:

| Compared with | Version | Why |
| --- | --- | --- |
| its previous step | the reference commit, an earlier commit of gobees-fast; it moves with the steps | did this step change the results? |
| gobees | the base commit, gobees with the output of the NetLogo metrics (see [Commits](#commits)); it stays fixed | differences too small to show in a single step but adding up over several |
| NetLogo | the 100 NetLogo runs of each scenario, embedded in [go/validation](../validation/README.md) | does it still match the original model? |

It also times the current code against gobees. The CI shows all results in one comment per pull request, a table
with one row per scenario and one column per comparison (see [The comment](#the-comment)).

## Phases

The steps are grouped into three phases, ordered by how strictly a step can be checked.

| Phase | What changes | Check against the previous step |
| --- | --- | --- |
| 1. Exact | Changes that keep the number and the order of all random draws, e.g. dead code, unused fields, fixes of findings in the NetLogo code that are not active in the default run | Same value on every day in every column, for every seed |
| 2. Statistically equal | Changes that alter the random draws or the order of agents, but should not change the dynamics, e.g. leaving out a shuffle, other Ark tables | Distributions over many seeds |
| 3. Expected effect | Fixes of findings in the NetLogo code that are active in the default run | Effect measured and reported, then decided |

gobees is deterministic for a given seed, so a change of phase 1 must give exactly the output of the reference.
Some changes without effect on the dynamics still change the random draws: a shuffle that is left out draws no
random numbers, and gobees shuffles before every `ask`, so a change of the order of the entities before a shuffle
(e.g. other Ark tables) changes its result. Such changes belong to phase 2.

## Scenarios

A scenario is a setting of the full model ([scenario.go](scenario.go)): the example farm with the parameters of a
NetLogo experiment in `go/validation/experiments`, and the same name. Its parameter file is that of the experiment
(or of the example farm) with `StopExtinct? = false`, as in all experiments, so that every run has all days. Its run
file in [scenarios](scenarios) is the conversion of this parameter file (`go run . generate` in `tools/netlogo2gobees`,
whose tests check that the run files are up to date).

| Scenario | Days | Setting |
| --- | ---: | --- |
| `step7` | 1095 | the example farm |
| `step7_badgers` | 1095 | 2 badgers |
| `step7_crop` | 1095 | a crop rotation of two maps |
| `step8_genetics` | 730 | diploid males, no unlimited males |
| `step8_species` | 730 | all 7 species |
| `step8_weather` | 730 | foraging hours per day of the year, low foraging mortality |
| `step8_settings` | 730 | further switches and a limit of hibernating queens |

The experiments of parts of the model (`step3`, `step4a`, `step6`) cannot run as a program. They are tests in
`go/sys` (`validation_step*_test.go`) and run with all other tests.

## The comparisons

Every version runs as a command line program (`go/main.go`) with `-s <seed> --days <days>`, the flag of its daily output (`-f <file>`, the current code `-o <folder>`) and its input:
gobees (`BASE_COMMIT`) the NetLogo files (`-d <folder> -p <parameter file>`), the current code the run file of the
scenario (`--config`, converted with `tools/netlogo2gobees` into [scenarios](scenarios)), the reference the format of
`REFERENCE_INPUT` in `commits.env`: with `yaml` the run files of its own commit (`COMPARE_REFERENCE_DIR`, set by
`run_compare.sh` to `go/compare` of the reference worktree), so the landscape files may change their format. The programs run with seeds 1 to 100, and is compared on its daily output (`obs.DailyOutputExtended`) from day 1 on (the
programs write no row for the setup). The columns are matched by name. Columns only the current code has are listed
in the comment and not compared. A column of the other version that the current code no longer has is an error, unless
it is listed in `REMOVED_COLUMNS`: a column is removed only together with its model branch.

gobees writes the same columns (its `obs.DailyOutput`) except `ForagerDeathsToday`, which it does not
count. The drift is therefore compared on all other columns, and the plots of the NetLogo report show the gobees
median. The fork point `40dc301` wrote only the 15 columns of `obs.DailyOutput` (`ticks` and 14 values).

| Comparison | Passes if |
| --- | --- |
| Exact (previous step, phase 1) | every value of every common column is equal, in every run |
| Statistical (previous step in phases 2 and 3, gobees always) | (1) the daily median of the current code lies within the 5 to 95 % band of the other version, for every day and column, and (2) no yearly summary (maximum and mean of each column in each year) differs in a two-sample Kolmogorov-Smirnov test at the level 0.05 divided by the number of tests (Bonferroni) |
| NetLogo | the daily median of each metric of the NetLogo experiment lies within the 5 to 95 % band of the NetLogo runs, for every day ([netlogo.go](netlogo.go)) |
| Speed | always: 10 runs of 365 days on the example farm per version, one after the other and alternating between the versions, without output; reports the median wall time per run |

The program writes the metrics of the NetLogo experiments as columns under their NetLogo names
(`obs.NetLogoMetrics`), so all three comparisons read the same output.

**Rounded columns.** The columns in `RoundedColumns` ([compare.go](compare.go)) are rounded to whole numbers in the
output of every version, before every comparison. At the moment this is only `foraging_period_s`. NetLogo, gobees
and older commits of gobees-fast compute the foraging period as hours × 3600, so on 15 of the 365 days of the
example weather year the value misses the whole second in the last bit (e.g. 2.2 × 3600 = 7920.000000000001).
gobees-fast reads whole seconds from the weather file (`step8_weather`). The model gives the same results, only this
column differs. Without the rounding, `step8_weather` would fail exactly against the previous step and, with a band
of width 0, against gobees.

**Last digits of float sums.** The statistical comparisons (previous step, gobees and NetLogo) allow a relative
tolerance of 1e-9 at the band (`tolerance` in [stats.go](stats.go)), and round the yearly summaries to 9 significant
digits before the Kolmogorov-Smirnov tests (`round9`). A sum of floats depends in its last digits (about 1e-15
relative) on the order of its terms, and the steps change such orders on purpose. Example: the landscape in metres
(2026-11-11) gives a map area of 2.485726977543863 km² instead of 2.485726977543862 in every run of
`step8_settings`, and the yearly maximum of `nectar_left_l` (the full stock of all layers, reached on days without
foraging) was summed in random order, now in file order. Without the tolerance, the band of width 0 of the map area
failed on all 730 days, and the Kolmogorov-Smirnov tests of the stocks gave D up to 0.89 in 4 scenarios. A real
change of the model moves the values by far more than 1e-9 relative.

**Why the Kolmogorov-Smirnov tests are needed.** The daily band alone is a weak check. Tested with the current
code against itself with a higher foraging mortality (`ForagingMortalityFactor` in the parameter file, 100 seeds,
3 years, 2026-11-01):

| `ForagingMortalityFactor` | Days and columns outside the band | Significant Kolmogorov-Smirnov tests |
| --- | --- | --- |
| 1.1 (10 % more forager deaths) | 0 of 26280 | 62 of 144 |
| 1.25 | 672 of 26280 | 112 of 144 |

Against NetLogo, the Kolmogorov-Smirnov tests are only reported (in the NetLogo report), not enforced: gobees itself
differs from NetLogo in some of them.

## Commits

[commits.env](commits.env) holds:

- `REFERENCE_COMMIT`: the previous step. It moves at the end of each phase, and in phase 3 after each fix.
- `BASE_COMMIT`: gobees. Not the fork point `40dc301`, but `ff06b97`, the later commit of gobees that writes the
  NetLogo metrics (its `obs.DailyOutput`). Between the two, only the output and the validation tests of gobees
  changed, not the model. The commit is not in the history of gobees-fast. It is in this repository as the tag
  `gobees-base`, which the CI gets with the full history (`fetch-depth: 0`). The tag is pushed once from the remote
  `upstream` (gobees): `git fetch upstream` and `git push origin ff06b97:refs/tags/gobees-base`. It stays fixed.
- `PHASE`: 1 compares exactly with the previous step, 2 and 3 statistically.
- `REMOVED_COLUMNS`: columns of the other versions that the current code removes on purpose, e.g. `TotalIBMColonies`
  and `ibm_colonies` with the individual-based colonies. They are not compared and are listed in the
  comment. gobees keeps them, so the list only grows.

## Running the comparisons

Linux or Git Bash, anywhere in the repository:

```shell
go/compare/run_compare.sh
```

The script checks the reference and the base commit out into temporary git worktrees, builds the command line
program of each and of the current working tree (incl. uncommitted changes), runs all scenarios and the speed
comparison (`TestScenarios`, `TestSpeedAgainstBase`) and writes the comment to `compare-comment.md` in the repository
root. The worktrees are removed at the end.

| Environment variable | Meaning |
| --- | --- |
| `COMPARE_THREADS` | runs at the same time, default: number of CPUs |
| `COMPARE_COMMENT` | file of the comment, default `compare-comment.md` in the repository root |
| `COMPARE_REPORT_DIR` | folder the NetLogo report `validation-report.md` and its plots are written to (the CI uses `go/validation`), none if empty |

Without the variables set by the script (`COMPARE_REFERENCE_BIN`, `COMPARE_CURRENT_BIN`, `COMPARE_BASE_BIN`),
`go test ./...` skips these tests and runs only the unit tests of the package.

## The comment

The CI ([.github/workflows/tests.yml](../../.github/workflows/tests.yml), job `compare`) runs the script on every
pull request and every push to `main` and posts the comment in the pull request (updated on every run) and in the
summary of the run (`Comment` in [report.go](report.go)):

```markdown
gobees-fast in this pull request, compared with:

| Scenario | its previous step (`40dc301`), exact | gobees (`40dc301`), statistical | NetLogo |
| --- | --- | --- | --- |
| Example farm, 3 years | ✅ identical | ✅ equal | ✅ 15 of 15 metrics in band |
| Badgers | ✅ identical | ✅ equal | ✅ 16 of 16 metrics in band |
| … | | | |

**Speed:** 1.02× as fast as gobees `40dc301` (…)
```

Below the table, the comment lists the runs and the compared columns and, collapsed, the details of every comparison
that failed. A scenario without a result (e.g. a program failed) is shown as failed. The CI also writes the NetLogo
report [go/validation/validation-report.md](../validation/validation-report.md) with one plot per scenario and, on
`main` after a merge, commits it if all comparisons passed (never to the branch of a pull request).
