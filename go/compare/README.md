# Comparison with the reference commit

gobees-fast starts as a copy of gobees and is changed in small steps. Every step is checked against an earlier
commit of gobees-fast, the reference, and against gobees itself, the base. This folder holds these checks. The
comparison with NetLogo ([go/validation](../validation/README.md)) runs in every step as well. The CI shows the
results of all of them in one comment per pull request (see [The comment](#the-comment)).

## Phases

The steps are grouped into three phases, ordered by how strictly a step can be checked.

| Phase | What changes | Check against the reference |
| --- | --- | --- |
| 1. Exact | Changes that keep the number and the order of all random draws, e.g. dead code, unused fields, fixes of quirks that are not active in the default run | Same value on every day in every column, for every seed |
| 2. Statistically equal | Changes that alter the random draws or the order of agents, but should not change the dynamics, e.g. leaving out a shuffle, other Ark tables | Distributions over many seeds |
| 3. Expected effect | Fixes of quirks that are active in the default run | Effect measured and reported, then decided |

gobees is deterministic for a given seed, so a change of phase 1 must give exactly the output of the reference.
Some changes without effect on the dynamics still change the random draws: a shuffle that is left out draws no
random numbers, and gobees shuffles before every `ask`, so a change of the order of the entities before a shuffle
(e.g. other Ark tables) changes its result. Such changes belong to phase 2.

## The comparisons

All runs use the example farm, with the input files of `go/internal/examplefiles`. Both versions run as command
line programs (`go/main.go`) with `-s <seed> --days <days> -f <file>` and are compared on their daily output
(`obs.DailyOutputExtended`). The columns are matched by name. Columns only the current code has are reported and
not compared (gobees writes only the 15 columns of `obs.DailyOutput`: `ticks` and 14 values). A column of the other version that the
current code no longer has is an error: the column names stay the same in all steps.

| Test | Runs | Passes if |
| --- | --- | --- |
| `TestExactAgainstReference` | seeds 1 to 100, 1095 days (3 years), per version | every value of every common column is equal |
| `TestStatisticalAgainstReference` | the same runs | (1) the daily median of the current code lies within the 5 to 95 % band of the reference, for every day and column, and (2) no yearly summary (maximum and mean of each column in each year) differs in a two-sample Kolmogorov-Smirnov test at the level 0.05 divided by the number of tests (Bonferroni) |
| `TestDriftAgainstBase` | the same seeds and days, against the base commit (gobees) | like `TestStatisticalAgainstReference` |
| `TestSpeedAgainstBase` | seeds 1 to 10, 365 days, per version, one run after the other | always (reports the median wall time per run and the speed-up) |

The comparisons use the same runs of each program. The reference moves with the steps, the base stays fixed, so
the drift comparison shows differences that are too small to show in a single step but add up over several.

**Why the Kolmogorov-Smirnov tests are needed.** The daily band alone is a weak check. Tested with the current
code against itself with a higher foraging mortality (`ForagingMortalityFactor` in the parameter file, 100 seeds,
3 years, 2026-11-01):

| `ForagingMortalityFactor` | Days and columns outside the band | Significant Kolmogorov-Smirnov tests |
| --- | --- | --- |
| 1.1 (10 % more forager deaths) | 0 of 26280 | 62 of 144 |
| 1.25 | 672 of 26280 | 112 of 144 |

## Commits

[commits.env](commits.env) holds:

- `REFERENCE_COMMIT`: the reference of the exact and the statistical comparison. It moves at the end of each
  phase, and in phase 3 after each fix. Until step F0 is merged, it is the fork point `40dc301`, i.e. gobees,
  which writes only the 15 columns of `obs.DailyOutput`. After the merge of F0 it is set to the merge commit.
- `BASE_COMMIT`: the base of the drift and the speed comparison, the fork point `40dc301`, i.e. gobees. It stays
  fixed.
- `COMPARE_TESTS`: the comparisons of the current phase. Phase 1: `Exact|Drift|Speed`. In phases 2 and 3 the exact
  comparison fails by design and is replaced by the statistical one: `Statistical|Drift|Speed`.

## Running the comparisons

Linux or Git Bash, anywhere in the repository:

```shell
go/compare/run_compare.sh
```

The script checks the reference and the base commit out into temporary git worktrees, builds the command line
program of each and of the current working tree (incl. uncommitted changes), and runs the tests of
`COMPARE_TESTS`. The worktrees are removed at the end. Locally with 2 runs at the same time, everything takes
about 1.5 minutes.

| Environment variable | Meaning |
| --- | --- |
| `COMPARE_THREADS` | runs at the same time, default: number of CPUs |
| `COMPARE_CHECKS` | folder the results are written to, one JSON file per question of the comment (the CI sets it) |

Without the variables set by the script (`COMPARE_REFERENCE_BIN`, `COMPARE_CURRENT_BIN`, `COMPARE_BASE_BIN`),
`go test ./...` skips these tests and runs only the unit tests of the package.

`go test -v` shows the result of each comparison in its log.

## The comment

The CI ([.github/workflows/tests.yml](../../.github/workflows/tests.yml), job `compare`) runs the script on every
pull request and every push to `main`. The job `report` writes one comment with one row per question, in the pull
request and in the summary of the run (`compare.Comment`, written by `go/validation/cmd/report`):

| Question | Answered by |
| --- | --- |
| Did this step change the results? | `TestExactAgainstReference` (phase 1) or `TestStatisticalAgainstReference` (phases 2 and 3) |
| Has the model drifted away from gobees? | `TestDriftAgainstBase` |
| Does it still match NetLogo? | the comparison tests of [go/validation](../validation/README.md): every metric within the NetLogo 5-95 % band on all days |
| How fast is it? | `TestSpeedAgainstBase` |

Below the rows, the comment lists the runs of each question and, collapsed, the details. A question without a
result (e.g. a program did not build) is shown as failed.
