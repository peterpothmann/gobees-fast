# Run time of NetLogo, gobees and gobees-fast

Compares the run time of `BEE-STEWARD` (NetLogo 5.3.1), `gobees` and `gobees-fast` on the example farm. `gobees` is the base commit in [go/compare/commits.env](../compare/commits.env) (the fork point), `gobees-fast` the current code. Report: [benchmark-report.md](benchmark-report.md). To check whether the models have the same behaviour please look into the validation ([go/validation/README.md](../validation/README.md)) and the comparison with the reference commit ([go/compare/README.md](../compare/README.md)).

## What is measured

All models run the example farm with the seeds 1 to 6, 365 days each, 2 runs at the same time. Per run, one time:

- **NetLogo**: `Setup` and 365 × `Go`, timed inside NetLogo.
- **gobees and gobees-fast**: the command line program (`go/main.go` of each commit, without output), timed from the start of the program to its end. The code of the base commit cannot time its setup and days separately, so this is the only time that is measured the same way in both Go versions. It includes the start of the process and reading the input files.

## Fairness

- Same CI runner, NetLogo first, then the program of `gobees`, then the one of `gobees-fast` (all compiled before NetLogo starts).
- Same input files, seeds, days and runs at the same time. `StopExtinct?` is false in NetLogo, so every run simulates all days. The programs read `StopExtinct?` from the parameter file. A program run that ends before the last day is an error.
- The runs differ between NetLogo and Go (NetLogo's random numbers cannot be reproduced in Go), so the report lists the colonies founded per run as a check of the work done.
- Each program runs once, untimed, before its timed runs. Without it, the first runs of a newly built program took up to 1.7 times as long as the others, and `gobees`, which runs first, looked slower than it is.
- The Go programs include the start of the process (about 50 ms locally), NetLogo's time does not include the start of Java. This favours NetLogo.
- No output during a run: NetLogo records its metrics only at the end (`runMetricsEveryStep="false"`).
- No GUI: the `Show...?` switches are turned off after NetLogo's `Setup`. `ShowCohorts?` ([GO-19]) and `ShowPlots?` ([OUT-03]) run code every day.
- No start-up time: the start of Java and the compilation of the model are not timed.
- NetLogo runs with Java 21 (it ships with Java 8), which favours NetLogo, if anything.

The first NetLogo run per thread includes the warm-up of Java (longer `Setup`). However, after testing: the median over the runs is barely affected.

## When it runs

The workflow `.github/workflows/benchmark.yml` runs in pull requests with the label `benchmark` (when the label is added and on every push). It builds `gobees` from the base commit in a git worktree. It commits the report to the branch and adds it to the summary of the run. It runs in parallel to the tests.

The comparison in [go/compare](../compare/README.md) also times `gobees-fast` against the base commit, in every pull request, but without NetLogo.

## Running it locally

From the repository root (`./gobees`), in Git Bash or a Linux shell, with Java on the PATH:

```shell
git worktree add --detach ../gobees-base 40dc301 && go -C ../gobees-base/go build -o "$PWD/gobees.exe" . && git worktree remove --force ../gobees-base
go -C go build -o "$PWD/gobees-fast.exe" .
bash go/benchmark/run_netlogo.sh "/c/Program Files/NetLogo 5.3.1" 2 benchmark_netlogo.csv benchmark_netlogo_console.txt
BENCHMARK_NETLOGO_TABLE=../../benchmark_netlogo.csv BENCHMARK_NETLOGO_CONSOLE=../../benchmark_netlogo_console.txt BENCHMARK_BASE_BIN=../../gobees.exe BENCHMARK_CURRENT_BIN=../../gobees-fast.exe BENCHMARK_BASE_COMMIT=40dc301 BENCHMARK_THREADS=2 BENCHMARK_REPORT=benchmark-report-local.md go -C go/benchmark test -count=1 -run '^TestCompareWithNetLogo$' .
```

The report is `go/benchmark/benchmark-report-local.md`, next to the report of the CI. The paths are relative to `go/benchmark`, where the test runs. The runs at the same time (here 2) must match in both commands. Without these variables the test is skipped.
