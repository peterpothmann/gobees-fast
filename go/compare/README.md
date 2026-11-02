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
| 1. Exact | Changes that keep the number and the order of all random draws, e.g. dead code, unused fields, fixes of quirks that are not active in the default run | Same value on every day in every column, for every seed |
| 2. Statistically equal | Changes that alter the random draws or the order of agents, but should not change the dynamics, e.g. leaving out a shuffle, other Ark tables | Distributions over many seeds |
| 3. Expected effect | Fixes of quirks that are active in the default run | Effect measured and reported, then decided |

gobees is deterministic for a given seed, so a change of phase 1 must give exactly the output of the reference.
Some changes without effect on the dynamics still change the random draws: a shuffle that is left out draws no
random numbers, and gobees shuffles before every `ask`, so a change of the order of the entities before a shuffle
(e.g. other Ark tables) changes its result. Such changes belong to phase 2.

## Scenarios

A scenario is a setting of the full model ([scenario.go](scenario.go)): the example farm with the parameters of a
NetLogo experiment in `go/validation/experiments`, and the same name. Its parameter file is that of the experiment
(or of the example farm) with `StopExtinct? = false`, as in all experiments, so that every run has all days.

| Scenario | Days | Setting |
| --- | ---: | --- |
| `step7` | 1095 | the example farm |
| `step7_badgers` | 1095 | 2 badgers |
| `step7_crop` | 1095 | a crop rotation of two maps |
| `step8_ibm` | 730 | 10 individual-based colonies |
| `step8_genetics` | 730 | diploid males, no unlimited males |
| `step8_species` | 730 | all 7 species |
| `step8_weather` | 730 | foraging hours per day of the year, low foraging mortality |
| `step8_settings` | 730 | further switches and a limit of hibernating queens |

The experiments of parts of the model (`step3`, `step4a`, `step6`) cannot run as a program. They are tests in
`go/sys` (`validation_step*_test.go`) and run with all other tests.

## The comparisons

Every version runs as a command line program (`go/main.go`) with `-s <seed> --days <days> -p <parameter file>
-f <file>`, with seeds 1 to 100, and is compared on its daily output (`obs.DailyOutputExtended`) from day 1 on (the
programs write no row for the setup). The columns are matched by name. Columns only the current code has are listed
in the comment and not compared. A column of the other version that the current code no longer has is an error: the
column names stay the same in all steps.

gobees writes the same columns (its `obs.DailyOutputExtended`) except `ForagerDeathsToday`, which it does not
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

- `REFERENCE_COMMIT`: the previous step. It moves at the end of each phase, and in phase 3 after each fix. Until the
  steps F0 are merged, it is the base commit, i.e. gobees, and shares its program. After the merge it is set to the
  merge commit.
- `BASE_COMMIT`: gobees. Not the fork point `40dc301`, but the later commit of gobees that writes the NetLogo
  metrics (`obs.DailyOutputExtended`). Between the two, only the output and the validation tests of gobees changed,
  not the model. The commit is not in the history of gobees-fast. It is in this repository as the tag `gobees-base`,
  which the CI gets with the full history (`fetch-depth: 0`). The tag is pushed once from the remote `upstream`
  (gobees): `git fetch upstream` and `git push origin <commit>:refs/tags/gobees-base`. It stays fixed.
- `PHASE`: 1 compares exactly with the previous step, 2 and 3 statistically.

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
report [go/validation/validation-report.md](../validation/validation-report.md) with one plot per scenario and, in a
pull request, commits it to the branch if all comparisons passed.
