# Usage of gobees-fast in detail

The short version is in the [README](../README.md#quick-start). This file is the reference: look up a section when you need it.


| Section                                           | Read it to ...                                              |
| ------------------------------------------------- | ----------------------------------------------------------- |
| [Command-line arguments](#command-line-arguments) | start a run and choose seed, days, runs and output          |
| [Run files](#run-files)                           | write or change a scenario (parameters, species, landscape) |
| [NetLogo input files](#netlogo-input-files)       | turn a NetLogo scenario into a gobees-fast run file         |
| [Daily output](#daily-output)                     | read the result file of a run                               |
| [Several runs](#several-runs)                     | run the model many times, e.g. for statistics               |
| [Experiments](#experiments)                       | vary parameters and run every combination                   |
| [Run info](#run-info)                             | repeat or document a run exactly                            |

## Command-line arguments

The flags control a single call of the model: which scenario runs, for how long, how often and where the output goes. Flags override the section `run` of the run file.

The flags of the model (`go/`, `go run .`). The flags of the converter are in [NetLogo input files](#netlogo-input-files).


| Flag                 | Short | Default                                    | Meaning                                                                                                                                                                                                                                                                                         |
| -------------------- | ----- | ------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--config PATH`      | `-c`  | the example farm (built in)                | run file (YAML). Its landscape files are relative to its folder. Repeat`-c` for several run files ([experiments](#experiments)).                                                                                                                                                                |
| `--seed N`           | `-s`  | `run.seed` of the run file                 | random seed of the first run, 1 or more (NetLogo:`RAND_SEED`).                                                                                                                                                                                                                                  |
| `--days N`           |       | `run.days` of the run file, else `365`     | number of days to run.`0`: no limit, the run ends only when the population is extinct and `stopExtinct` is true.                                                                                                                                                                                |
| `--runs N`           |       | `run.runs` of the run file                 | number of runs (of each combination of an [experiment](#experiments)), 1 or more, in one process. Run i has the seed `SEED + i - 1`. All runs of a run file go into one output file.                                                                                                           |
| `--threads T`        |       | `run.threads` of the run file              | number of runs at the same time, 1 or more. Each run uses one CPU core, so set it to the number of cores you want to use (on a cluster: the CPUs of the job, e.g.`--threads $SLURM_CPUS_PER_TASK`). `1`: all runs one after the other, also across run files. The output does not depend on it. |
| `--output`           | `-o`  | off                                        | write the daily output to`daily_output.csv` in the working directory (`go/`), and the [run info](#run-info) next to it.                                                                                                                                                                         |
| `--output-file PATH` | `-f`  |                                            | write the daily output to`PATH` instead (implies `-o`). Missing folders are created.                                                                                                                                                                                                            |
| `--columns A,B,C`    |       | `output.columns` of the run file, else all | columns of the daily output, in this order (needs`-o` or `-f`). An unknown name is an error that lists the columns.                                                                                                                                                                             |
| `--list-columns`     |       |                                            | print the columns of the daily output with a description and exit.                                                                                                                                                                                                                              |
| `--help`             | `-h`  |                                            | show all flags.                                                                                                                                                                                                                                                                                 |

```shell
go run . -c path/to/run.yaml -s 3 --days 730       # own run file, seed 3, 2 years
go run . -f path/to/output.csv                     # output to any path
go run . -c path/to/output_run_info.yaml -f x.csv  # repeat the run of path/to/output.csv
go run . -o --columns colonies,workers,eggs        # only these columns (and ticks)
go run . -c a.yaml -c b.yaml -f out/d.csv --threads 1  # two run files, one run after the other: out/d_a.csv, out/d_b.csv
```

## Run files

A run file (YAML) describes one scenario completely: parameters, species, flowers, landscape and run settings. The model reads nothing else, so the run file and its landscape files are all you need to run or share a scenario.

A run file has the sections below. All keys are in lowerCamelCase, values are in the unit of the model (e.g. seconds). [names.md](names.md) lists every key with its unit and the NetLogo name.


| Key                  | Content                                                                                                                                                                                                                                                                          | NetLogo                                                                         |
| -------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| `run`                | `seed`, `days`, `runs` and `threads` (the flags `-s`, `--days`, `--runs` and `--threads` override them). Needed: `seed`, `runs` and `threads`, each 1 or more. The converter writes seed 1, 365 days, 1 run and 1 thread. In a [run info](#run-info) also the record of the run. | `RAND_SEED`                                                                     |
| `variations`         | optional: keys and their values, every combination runs (an[experiment](#experiments))                                                                                                                                                                                           |                                                                                 |
| `parameters`         | the model parameters                                                                                                                                                                                                                                                             | the columns of the parameter file                                               |
| `output`             | output options (`keepDeadColonies`, `columns`), change no model value                                                                                                                                                                                                            | `KeepDeadColonies?`                                                             |
| `constants`          | the values that NetLogo hard-codes, every key needed (see below)                                                                                                                                                                                                                 | `ParametersProc` and local values of the procedures                             |
| `initialQueens`      | species and number of initial queens                                                                                                                                                                                                                                             | `BeeSpeciesInitialQueensListAsString`                                           |
| `foragingPeriod`     | foraging period [s], 365 values per year, one or more years, the years repeat                                                                                                                                                                                                    | `Weather`                                                                       |
| `stewardshipOptions` | names of the stewardship options                                                                                                                                                                                                                                                 | from the`HabitatsFile`                                                          |
| `species`            | bee species, found by`name` (only those of `initialQueens` are used), `parasitic: true` for a cuckoo bumblebee                                                                                                                                                                   | rows of the`BeespeciesFile` (columns `species` = `name`, Psithyrus by its name) |
| `flowers`            | flower species, without the key no flower values                                                                                                                                                                                                                                 | rows of the`FlowerspeciesFile` (without `proteinPollenProp`)                    |
| `landscape`          | landscape file of the start                                                                                                                                                                                                                                                      | foodsources file of`MyMap`                                                      |
| `cropRotation`       | landscape files of the crop rotation (fewer than 2: none)                                                                                                                                                                                                                        | `CropRotationListAsString`                                                      |

A landscape file holds `metresPerPatch` and the flower patches (`id`, `patchType`, `xcor`, `ycor`, `area`, `flowers`), the columns of the NetLogo foodsources file without the GUI columns. All files are read at the start (package `input`), also the landscapes of a crop rotation. An unknown key is an error. A missing key keeps its zero value, except in `constants`. Unlike NetLogo, a missing file is an error at the start, also a landscape of the crop rotation (NetLogo: on its day 365).

The section `constants` holds the values that NetLogo sets in its code and that NetLogo's parameter files cannot change: the values of `ParametersProc` and the local values of the procedures (e.g. the pause of a bee, the probabilities of the badgers and the cuckoo bumblebees, the task thresholds). It is grouped by part of the model (`activities`, `stimuli`, `foraging`, `colony`, `bees`, `psithyrus`, `badgers`, `landscape`, `initialQueens`, `thresholds`) and must hold every key. A run file without a key of `constants` is an error, so that no value of a run comes from the code unseen. The converter writes the NetLogo values (`params.DefaultConstants`) with unit and source as a comment at each key. The coefficients of fitted equations (handling time, competition point, winter survival) stay in the code.

## NetLogo input files

gobees-fast does not read the NetLogo input files. To run a scenario that exists for NetLogo, convert it once into a run file with [tools/netlogo2gobees](../tools/netlogo2gobees). The converter is its own Go module, the model does not import it.

The converter has two commands:

- `convert` turns one NetLogo scenario into a run file: the parameter file, the files it names, and the map (`MyMap`, a setting of the NetLogo interface), which gives the foodsources file of the start. This is the command for users.
- `generate` is for developers of gobees-fast. The repository contains run files made from NetLogo files: the built-in example farm (`go/input/examplefarm`) and the scenarios of the step check (`go/compare/scenarios`). `generate` writes them all again. Call it after a change of the NetLogo files, the scenarios or the run file format. A test fails if these run files are out of date.

The flags of the converter (`tools/netlogo2gobees`, `go run . <command>`):


| Command    | Flag                | Short | Default                    | Meaning                                                                                                                                |
| ---------- | ------------------- | ----- | -------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `convert`  | `--parameters PATH` | `-p`  | (required)                 | NetLogo parameter file (`MyParametersFile`).                                                                                           |
| `convert`  | `--dir PATH`        | `-d`  | `../../netlogo`            | folder of the NetLogo input files (the folder of the`.nlogo`). The file names in the parameter file are relative to it.                |
| `convert`  | `--map NAME`        | `-m`  | `_SYSTEM_Example_Farm.png` | map (`MyMap`). The foodsources are read from `<NAME without .png>_Foodsources.txt` in `--dir`.                                         |
| `convert`  | `--out PATH`        | `-o`  | (required)                 | folder of the run file and its landscape files, created if missing.                                                                    |
| `convert`  | `--run NAME`        |       | `run.yaml`                 | name of the run file in`--out`.                                                                                                        |
| `generate` | `--root PATH`       |       | `../..`                    | folder of the gobees-fast repository, where`generate` writes its files. The default fits when you call it from `tools/netlogo2gobees`. |

```shell
cd tools/netlogo2gobees
# example farm: parameter file and output folder are required, the map is the example farm by default
go run . convert -p ../../netlogo/_SYSTEM_Example_Farm_Parameters.csv -o farm
# other map: writes sussex/run.yaml and the landscape file next to it
go run . convert -p ../../netlogo/_SYSTEM_Sussex1_Parameters.csv -m _SYSTEM_Sussex1.png -o sussex
# own NetLogo files in another folder, writes mine/my_run.yaml
go run . convert -d path/to/netlogo -p path/to/netlogo/my_Parameters.csv -m my_map.png -o mine --run my_run.yaml
# all flags of a command
go run . convert --help
# developers: renew the run files of the repository
go run . generate
```

`-p` is a path from the folder you call the converter in. The files named in the parameter file and the map are looked up in `-d`. Run the converted scenario with the model:

```shell
cd ../../go
go run . -c ../tools/netlogo2gobees/sussex/run.yaml -o
```

What the converter does with the NetLogo values:

- `Weather` is either one of NetLogo's constant options (`"Constant 8 hrs"`, ...: 365 equal values) or the name of a weather file in the folder of the NetLogo files: a CSV table with the columns `Day` (1 to 365) and `ForagingPeriod_s` (0 to 86400), and optionally `Year` (1, 2, ...). NetLogo's example year `"foragingHoursExample"`, a list in the NetLogo code, is the file [netlogo/_SYSTEM_Weather_foragingHoursExample.csv](../netlogo/_SYSTEM_Weather_foragingHoursExample.csv).
- Errors: a missing parameter file (NetLogo loads `_SYSTEM_Parameters.csv` instead), `COLONIES_IBM` other than 0, `MapAreaIncluded` other than `"complete"` (a part of a map is its own landscape file), and a species row whose columns `species` and `name` differ.
- The converter prints every column it does not take over, with the reason.
- NetLogo's parameter files set `MaxForagingRange_m` = 10000, which has no effect in NetLogo: NetLogo uses the constant `ForagingRangeMax_m` = 758 m. The converter writes 758 and says so, in its output and in the comment at `constants.foraging.maxForagingRange`.

## Daily output

The daily output is the result of a run: the state of the bee population on every day, for analysis in R, Python or a spreadsheet.

The daily output (`-o`, `-f`) is a tab-separated file with one row per run and day, holding the state at the end of the day. It holds the totals of `OutputDailyProc` (colonies, queens, males, mean weights, ...), followed by brood, workers, stores, foraging trips, forager deaths, queens produced and the mean competition point date (`obs.DailyOutputExtended`).


| Column   | Content                                                 |
| -------- | ------------------------------------------------------- |
| `run`    | number of the run, from 1                               |
| `seed`   | seed of the run                                         |
| `t`      | days from 0, as the CSV reporter of ark-tools counts    |
| `ticks`  | NetLogo ticks, 1 on the first day                       |
| the rest | the totals of`OutputDailyProc` keep their NetLogo names |

`go run . --list-columns` prints all columns with a description. The key `output.columns` of the run file (a list of names) or the flag `--columns a,b,c` (which overrides the key) selects columns and their order. `ticks` is always the first column after `t`: it is added if missing and moved to the front. Without a selection, all columns are written. An unknown or repeated name is an error that lists the available columns, also when the run has no output. Every value is still computed, so the values of the selected columns do not depend on the selection.

## Several runs

The model is stochastic: two seeds give two different results. To see the spread of the results, run the same scenario with many seeds.

With `--runs N` (or `run.runs`), gobees-fast runs the model N times in one process, at most `--threads` runs at the same time (or `run.threads`). Run i has the seed `seed + i - 1`, so the same `seed` and `runs` always give the same runs. Seed, runs and threads have no special value: each must be 1 or more, nothing is drawn at random or taken from the machine. The output holds the runs one after the other, in the order of the runs, so the file does not depend on the number of threads. The runs share the input read at the start, the model does not change it.

## Experiments

An experiment answers questions such as "how does the number of badgers change the population?". It replaces BehaviorSpace of NetLogo.

The optional section `variations` makes a run file an experiment: each listed key takes each of its values, and every combination (full factorial) runs `run.runs` times.

```yaml
run:
  seed: 1
  days: 365
  runs: 10      # per combination
  threads: 8
variations:
  - key: parameters.badgers
    values: [0, 2, 4]
  - key: constants.foraging.maxForagingRange
    values: [500, 758, 1000]
parameters:
  ...
```

This gives 3 × 3 combinations × 10 runs = 90 runs. Rules:

- `key` is the path of a key in the run file, its parts separated by dots. A number selects an item of a list, e.g. `initialQueens.0.queens` (the first species) or `species.6.name`. A value is written as in the run file: a number, a text, a list.
- `values` lists every value. There are no ranges, so every value of an experiment is written out.
- Every combination is the run file with these values, read and checked like a run file: an unknown key, a fraction in a whole number and so on are errors at the start, naming the combination.
- Every combination runs with the same seeds (`seed` to `seed + runs - 1`), so combinations differ only in the varied values, not in the random numbers.
- The section `run` (seed, days, runs, threads), `output.columns`, `landscape` and `cropRotation` cannot vary. Other landscapes are other run files (see below).
- The daily output is one file: the columns `run`, `seed`, then one column per varied key (named by the key, holding the value), then `t` and the columns. The first variation changes slowest. The run info holds the variations, so it repeats the whole experiment.

Several run files run together with `-c a.yaml -c b.yaml`: their runs share one pool of `--threads` threads (the run files must ask for the same `threads`, or the flag sets it). Each run file writes its own output and run info, named after the run file: `-f out/daily.csv` gives `out/daily_a.csv` and `out/daily_b.csv`. Two run files with the same name are an error. The flags `-s`, `--days`, `--runs` and `--columns` apply to all run files.

## Run info

The run info records everything needed to repeat a run: the run file, the landscapes, the seed, the command and the program version. Keep it with the output, for example for a paper or to check a result later.

With an output (`-o`, `-f`), gobees-fast writes the run info next to it, before the run: the name of the output without extension plus `_run_info.yaml` (`daily_output.csv` → `daily_output_run_info.yaml`). The run info is a run file that repeats the run: `go run . -c daily_output_run_info.yaml -f again.csv` writes the same daily output, byte for byte. It holds all sections of the run file as read (the constants included) and the section `run`:


| Key               | Content                                                                                                                                                                                                   |
| ----------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `seed`            | the seed of the first run                                                                                                                                                                                 |
| `days`            | the number of days                                                                                                                                                                                        |
| `runs`, `threads` | the numbers of runs and threads                                                                                                                                                                           |
| `started`         | start time (RFC 3339)                                                                                                                                                                                     |
| `command`         | command line                                                                                                                                                                                              |
| `version`         | module, version, VCS revision, its time and`modified` (changes against the revision), Go version. The VCS keys are only known in a binary built with `go build` inside the repository, not with `go run`. |
| `config`          | the run file given with`-c`, or `built-in example farm`                                                                                                                                                   |
| `output`          | the output file. Its columns are the key`columns` of the section `output` of the run info, always written in full                                                                                         |
| `landscapes`      | per landscape file: the copy (`file`), the name in the run file (`source`) and the SHA-256                                                                                                                |

The landscape files are copied byte for byte next to the run info, with the name of the output in front (`daily_output__SYSTEM_Example_Farm_Foodsources.yaml`). The keys `landscape` and `cropRotation` of the run info name the copies, so the run info and its copies can be moved or archived together. Only `seed`, `days`, `runs` and `threads` of the section `run` are read, the other keys are a record.
