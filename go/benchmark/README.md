# Run time of gobees and NetLogo

Compares the run time of `gobees` with `BEE-STEWARD` (NetLogo 5.3.1) on the example farm. Report: [benchmark-report.md](benchmark-report.md). To check whether both models have the same behaviour please look into the validation ([go/validation/README.md](../validation/README.md)).

## What is measured

Both models run the example farm with the seeds 1 to 6, 365 days each, 2 runs at the same time. Per run, two times:

- **Setup**: NetLogo's `Setup`, in `gobees` reading the parameter file, `model.Default` and `Initialize`.
- **Days**: 365 × NetLogo's `Go`, 365 × `Update` in `gobees`.

## Fairness

- Same CI runner, NetLogo first, then `gobees` (compiled before NetLogo starts).
- Same input files, seeds, days and runs at the same time. `StopExtinct?` is false, so every run simulates all days. The runs differ between the models (NetLogo's random numbers cannot be reproduced in Go), so the report lists the colonies founded per run as a check of the work done.
- No output during a run: NetLogo records its metrics only at the end (`runMetricsEveryStep="false"`).
- No GUI: the `Show...?` switches are turned off after NetLogo's `Setup`. `ShowCohorts?` ([GO-19]) and `ShowPlots?` ([OUT-03]) run code every day.
- No start-up time: the start of Java and the compilation of the model are not timed.
- NetLogo runs with Java 21 (it ships with Java 8), which favours NetLogo, if anything.

The first NetLogo run per thread includes the warm-up of Java (longer `Setup`). However, after testing: the median over the runs is barely affected.

## When it runs

The workflow `.github/workflows/benchmark.yml` runs in pull requests with the label `benchmark` (when the label is added and on every push). It commits the report to the branch and adds it to the summary of the run. It runs in parallel to the tests.

## Running it locally

From the repository root (`./gobees`), in Git Bash or a Linux shell, with Java on the PATH:

```shell
bash go/benchmark/run_netlogo.sh "/c/Program Files/NetLogo 5.3.1" 2 benchmark_netlogo.csv benchmark_netlogo_console.txt
BENCHMARK_NETLOGO_TABLE=../../benchmark_netlogo.csv BENCHMARK_NETLOGO_CONSOLE=../../benchmark_netlogo_console.txt BENCHMARK_THREADS=2 BENCHMARK_REPORT=benchmark-report-local.md go -C go/benchmark test -count=1 -run '^TestCompareWithNetLogo$' .
```

The report is `go/benchmark/benchmark-report-local.md`, next to the report of the CI. The paths are relative to `go/benchmark`, where the test runs. The runs at the same time (here 2) must match in both commands. Without these variables the test is skipped.
