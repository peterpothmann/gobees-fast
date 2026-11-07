<div align="center">

<img alt="gobees-fast" src="docs/logo/gobees-fast-logo-dark.svg" width="720">

[![Tests](https://github.com/peterpothmann/gobees-fast/actions/workflows/tests.yml/badge.svg)](https://github.com/peterpothmann/gobees-fast/actions/workflows/tests.yml)

</div>

`gobees-fast` is the optimized version of [gobees](https://github.com/peterpothmann/gobees), a port of [BEE-STEWARD](https://beehave-model.net/bee-steward-is-published/) ([Twiston-Davies et al. 2021](#ref-twiston-davies2021)) from `NetLogo` to `Go`, built on the [Ark](https://github.com/mlange-42/ark) entity component system. `BEE-STEWARD` is a research and decision support tool for land management that promotes bumblebee populations. It combines the bumblebee population model [Bumble-BEEHAVE](https://beehave-model.net/category/bumblebeehave/) ([Becher et al. 2018](#ref-becher2018)) with the landscape module [BEESCOUT](https://beehave-model.net/category/beescout/) ([Becher et al. 2016](#ref-becher2016)). It simulates colonies of six *Bombus* species and the cuckoo bumblebee *Psithyrus*, from queen emergence through foraging and brood rearing to hibernation, over several years. The bees forage in a nectar and pollen landscape built from maps, with the flowering period of each plant species. Agri-environment schemes, such as flower strips, and crop rotations can be applied to selected fields.

All credit for the scientific model, its development, and parameterisation belongs to the original authors and contributors of `BEE-STEWARD` and its component models.

> [!NOTE]
> gobees-fast is at its first step: apart from a longer daily output and the comparisons in [go/compare](go/compare/README.md), it is still the code of gobees (commit `40dc301`). The rest of this README describes gobees.

A graphical interface for gobees is available in [gobees-ui](https://github.com/peterpothmann/gobees-ui). It shows the landscape, the flowering calendar and the queens of a running simulation, and explains how the model works in a schematic animation.

## gobees and gobees-fast

- **gobees** stays as close as possible to the NetLogo code: same quirks, same variable and procedure names, same structure. It is meant for users who know BEE-STEWARD in NetLogo and want to move to Go, and who need to trace every result back to the NetLogo code.
- **gobees-fast** (this repository) starts as a copy of gobees, with its history. It is changed in small steps: dead code and unused data are removed, the data model is simplified, and the quirks of the NetLogo code ([trans/netlogo-findings.md](trans/netlogo-findings.md)) are fixed. It is meant for users who learn BEE-STEWARD from scratch, and for applications where performance matters most.

Every step is compared with an earlier commit of gobees-fast, first value by value, later by distributions over 100 seeds ([go/compare](go/compare/README.md)), and with NetLogo ([go/validation](go/validation/README.md)). Once a quirk that is active in the default run is fixed, the results differ from NetLogo, and gobees-fast is a modified model version, not BEE-STEWARD in Go.

`git diff upstream/main` shows every change against gobees, with the remote `upstream` set to [https://github.com/peterpothmann/gobees](https://github.com/peterpothmann/gobees).

## How the translation works

Each step follows the same order:

1. The NetLogo procedures of the step are annotated in [netlogo/__BEE-STEWARD_RC_ColourVersionMB.nlogo](netlogo/__BEE-STEWARD_RC_ColourVersionMB.nlogo). Annotations change comments only. Every procedure gets a header, and every step inside it gets a code such as `[UCOL-02]`.
2. Questionable behaviour of the NetLogo code is recorded in [trans/netlogo-findings.md](trans/netlogo-findings.md) and checked in NetLogo.
3. The step is translated to Go. Every system and function names the codes it translates, and every finding gets a test.
4. Where possible, gobees is compared with NetLogo (see [Validation](#validation)).

The reference is written in NetLogo 5.3.1.

## Repository layout


| Folder                 | Content                                                                                                             |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------- |
| `go/schedule`          | system order of setup and one simulated day                                                                         |
| `go/sys`               | 4 systems (Setup, Morning, Day, Evening), their parts one per NetLogo procedure, file names prefixed with the group |
| `go/comp`              | components of the entities (bees, colonies, foodsources, ...)                                                       |
| `go/globals`           | global state of the run (changes during the run)                                                                    |
| `go/params`            | parameters, the content of the input files and the values derived from them (read-only during the run)              |
| `go/input`             | reader of the run files (YAML), the built-in example farm                                                           |
| `go/obs`               | observers of the model output                                                                                       |
| `go/benchmark`         | run time of NetLogo, gobees and gobees-fast                                                                         |
| `go/compare`           | check of each step: scenarios against the previous step, gobees and NetLogo                                         |
| `go/enum`              | castes, stages and activities as Go constants                                                                       |
| `tools/netlogo2gobees` | converter of the NetLogo input files into run files (own Go module)                                                 |
| `netlogo`              | the annotated NetLogo model and its input files                                                                     |
| `trans`                | findings on the NetLogo code and the validation against NetLogo                                                     |

The layout follows [beecs](https://github.com/mlange-42/beecs).

## Usage

Requires Go 1.26 or later.

```shell
cd go
go run .
```

This runs the example farm for 365 days and prints a short summary. The example farm is built in
([go/input/examplefarm](go/input/examplefarm), converted from `netlogo/_SYSTEM_Example_Farm_Parameters.csv`).
Other scenarios need a run file (YAML) and its landscape files (see [Run files](#run-files)).

### Command-line arguments

The model (`go/`, `go run .`):


| Flag                 | Short | Default                                | Meaning                                                                                                                                |
| -------------------- | ----- | -------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `--config PATH`      | `-c`  | the example farm (built in)            | run file (YAML). Its landscape files are relative to its folder.                                                                       |
| `--seed N`           | `-s`  | `run.seed` of the run file, else `0`   | random seed (NetLogo:`RAND_SEED`). `0` draws a random seed, which the summary and the run info show.                                   |
| `--days N`           |       | `run.days` of the run file, else `365` | number of days to run.`0`: no limit, the run ends only when the population is extinct and `stopExtinct` is true.                       |
| `--output`           | `-o`  | off                                    | write the daily output to`daily_output.csv` in the working directory (`go/`), tab separated, and the [run info](#run-info) next to it. |
| `--output-file PATH` | `-f`  |                                        | write the daily output to`PATH` instead (implies `-o`). Missing folders are created.                                                   |
| `--help`             | `-h`  |                                        | show all flags.                                                                                                                        |

```shell
go run . -c path/to/run.yaml -s 3 --days 730   # own run file, seed 3, 2 years
go run . -o                                    # example farm, output in go/daily_output.csv
go run . -f path/to/output.csv                 # output to any path
go run . -c path/to/output_run_info.yaml -f x.csv  # repeat the run of path/to/output.csv
```

The converter of the NetLogo input files (`tools/netlogo2gobees`, `go run . <command>`, see
[NetLogo input files](#netlogo-input-files)):


| Command    | Flag                | Short | Default                    | Meaning                                                                                                                                      |
| ---------- | ------------------- | ----- | -------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `convert`  | `--parameters PATH` | `-p`  | (required)                 | NetLogo parameter file (`MyParametersFile`).                                                                                                 |
| `convert`  | `--dir PATH`        | `-d`  | `../../netlogo`            | folder of the NetLogo input files (the folder of the`.nlogo`). The file names in the parameter file are relative to it.                      |
| `convert`  | `--map NAME`        | `-m`  | `_SYSTEM_Example_Farm.png` | map (`MyMap`). The foodsources are read from `<NAME without .png>_Foodsources.txt` in `--dir`.                                               |
| `convert`  | `--out PATH`        | `-o`  | (required)                 | folder of the run file and its landscape files, created if missing.                                                                          |
| `convert`  | `--run NAME`        |       | `run.yaml`                 | name of the run file in`--out`.                                                                                                              |
| `generate` | `--root PATH`       |       | `../..`                    | root of the repository. Writes the run files of the repository: the example farm (`go/input/examplefarm`) and the scenarios of `go/compare`. |

### Run files

A run file has these sections. All keys are in lowerCamelCase, values are in the unit of the model (e.g. seconds).
[docs/names.md](docs/names.md) lists every key with its unit and the NetLogo name.


| Key                  | Content                                                                                                                                  | NetLogo                                                                         |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| `run`                | optional:`seed` and `days` of the run (the flags `-s` and `--days` override them), in a [run info](#run-info) also the record of the run | `RAND_SEED`                                                                     |
| `parameters`         | the model parameters                                                                                                                     | the columns of the parameter file                                               |
| `output`             | output options (`keepDeadColonies`), change no model value                                                                               | `KeepDeadColonies?`                                                             |
| `constants`          | the values that NetLogo hard-codes, every key needed (see below)                                                                         | `ParametersProc` and local values of the procedures                             |
| `initialQueens`      | species and number of initial queens                                                                                                     | `BeeSpeciesInitialQueensListAsString`                                           |
| `foragingPeriod`     | foraging period [s], 365 values per year, one or more years, the years repeat                                                            | `Weather`                                                                       |
| `stewardshipOptions` | names of the stewardship options                                                                                                         | from the`HabitatsFile`                                                          |
| `species`            | bee species, found by`name` (only those of `initialQueens` are used), `parasitic: true` for a cuckoo bumblebee                           | rows of the`BeespeciesFile` (columns `species` = `name`, Psithyrus by its name) |
| `flowers`            | flower species, without the key no flower values                                                                                         | rows of the`FlowerspeciesFile` (without `proteinPollenProp`)                    |
| `landscape`          | landscape file of the start                                                                                                              | foodsources file of`MyMap`                                                      |
| `cropRotation`       | landscape files of the crop rotation (fewer than 2: none)                                                                                | `CropRotationListAsString`                                                      |

A landscape file holds `metresPerPatch` and the flower patches (`id`, `patchType`, `xcor`, `ycor`, `area`, `flowers`),
the columns of the NetLogo foodsources file without the GUI columns. All files are read at the start (package `input`),
also the landscapes of a crop rotation. An unknown key is an error, a missing key keeps its zero value, except in
`constants`. Unlike NetLogo,
a missing file is an error at the start, also a landscape of the crop rotation (NetLogo: on its day 365).

The section `constants` holds the values that NetLogo sets in its code and that NetLogo's parameter files cannot
change: the values of `ParametersProc` and the local values of the procedures (e.g. the pause of a bee, the
probabilities of the badgers and the cuckoo bumblebees, the task thresholds). It is grouped by part of the model
(`activities`, `stimuli`, `foraging`, `colony`, `bees`, `psithyrus`, `badgers`, `landscape`, `initialQueens`,
`thresholds`) and must hold every key: a run file without a key of `constants` is an error, so that no value of a
run comes from the code unseen. The converter writes the NetLogo values (`params.DefaultConstants`) with unit and
source as a comment at each key. The coefficients of fitted equations (handling time, competition point, winter
survival) stay in the code.

### NetLogo input files

[tools/netlogo2gobees](tools/netlogo2gobees) converts a NetLogo parameter file and the files it names into a run file
(its own Go module, the model does not import it). The map (`MyMap`, a setting of the NetLogo interface) gives the
foodsources file of the start.

```shell
cd tools/netlogo2gobees
go run . convert -d ../../netlogo -p ../../netlogo/_SYSTEM_Sussex1_Parameters.csv -m _SYSTEM_Sussex1.png -o sussex
go run . generate
```

`Weather` is either one of NetLogo's constant options (`"Constant 8 hrs"`, ...: 365 equal values) or the name of a
weather file in the folder of the NetLogo files: a CSV table with the columns `Day` (1 to 365) and `ForagingPeriod_s`
(0 to 86400), and optionally `Year` (1, 2, ...). NetLogo's example year `"foragingHoursExample"`, a list in the NetLogo
code, is the file [netlogo/_SYSTEM_Weather_foragingHoursExample.csv](netlogo/_SYSTEM_Weather_foragingHoursExample.csv).
A missing parameter file is an error (NetLogo loads `_SYSTEM_Parameters.csv` instead). Also an error: `COLONIES_IBM`
other than 0, `MapAreaIncluded` other than `"complete"` (a part of a map is its own landscape file), and a species row
whose columns `species` and `name` differ. The converter prints every column it does not take over, with the
reason. NetLogo's parameter files set `MaxForagingRange_m` = 10000, which has no effect in NetLogo: NetLogo uses the
constant `ForagingRangeMax_m` = 758 m. The converter writes 758 and says so, in its output and in the comment at
`constants.foraging.maxForagingRange`. The tests of the converter check
that the run files of the repository are the conversion of the NetLogo files.

### Daily output

The daily output (`-o`, `-f`) is a tab-separated file. It holds the totals of `OutputDailyProc` (colonies, queens, males, mean weights, ...), followed by brood, workers, stores, foraging trips, forager deaths, queens produced and the mean competition point date (`obs.DailyOutputExtended`). The CSV reporter of ark-tools adds one row at the end of every tick. Its first column `t` counts from 0, the column `ticks` holds the NetLogo ticks (1 on the first day). The totals of `OutputDailyProc` keep their NetLogo names.

### Run info

With an output (`-o`, `-f`), gobees-fast writes the run info next to it, before the run: the name of the output
without extension plus `_run_info.yaml` (`daily_output.csv` → `daily_output_run_info.yaml`). The run info is a run
file that repeats the run: `go run . -c daily_output_run_info.yaml -f again.csv` writes the same daily output, byte
for byte (keyword: reproduceability). It holds all sections of the run file as read (the constants included) and the section `run`:


| Key                 | Content                                                                                                                                                                                                   |
| ------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `seed`              | the seed used. A seed`0` is drawn before the run and recorded here.                                                                                                                                       |
| `days`              | the number of days                                                                                                                                                                                        |
| `started`           | start time (RFC 3339)                                                                                                                                                                                     |
| `command`           | command line                                                                                                                                                                                              |
| `version`           | module, version, VCS revision, its time and`modified` (changes against the revision), Go version. The VCS keys are only known in a binary built with `go build` inside the repository, not with `go run`. |
| `config`            | the run file given with`-c`, or `built-in example farm`                                                                                                                                                   |
| `output`, `columns` | the output file and its columns                                                                                                                                                                           |
| `landscapes`        | per landscape file: the copy (`file`), the name in the run file (`source`) and the SHA-256                                                                                                                |

The landscape files are copied byte for byte next to the run info, with the name of the output in front
(`daily_output__SYSTEM_Example_Farm_Foodsources.yaml`). The keys `landscape` and `cropRotation` of the run info name
the copies, so the run info and its copies can be moved or archived together. Only `seed` and `days` of the section
`run` are read, the other keys are a record.

Tests:

```shell
cd go
go test -timeout 30m ./...   # all tests, incl. the comparison with NetLogo (about 10 minutes)
go test -short ./...         # without the comparison with NetLogo
```

## Validation

Single runs of gobees and NetLogo cannot be compared, because NetLogo's random number sequence cannot be reproduced in Go. Instead, 100 NetLogo runs give quantiles per day for each compared quantity, and a Go test requires the median of 100 gobees runs to lie within the NetLogo 5–95 % band. The NetLogo results are embedded in the repository, so NetLogo does not have to be installed to run the tests.

The experiments of the full model run as scenarios of the check of each step (below), which writes the report [go/validation/validation-report.md](go/validation/validation-report.md) with one plot per scenario and, in a pull request, commits it to the branch. The experiments of parts of the model are Go tests. Details on the experiments: [go/validation/README.md](go/validation/README.md).

## Check of each step

The CI checks every pull request and every push to `main`: it runs each scenario (the example farm and the settings of the NetLogo experiments) 100 times with gobees-fast and compares it with three versions: its previous step (a reference commit of gobees-fast, value by value while a step must not change any random draw, otherwise by distributions), gobees (by distributions, to catch differences that add up over several steps) and NetLogo (the band of the NetLogo runs). It also times gobees-fast against gobees. One comment in the pull request shows the results as a table, one row per scenario and one column per comparison. Locally: `go/compare/run_compare.sh`. Details: [go/compare/README.md](go/compare/README.md).

## Benchmarks

In pull requests with the label `benchmark`, the CI runs the example farm in NetLogo 5.3.1, in gobees (the fork point) and in gobees-fast on the same runner, with the same seeds and days, and times each run. It commits the report [go/benchmark/benchmark-report.md](go/benchmark/benchmark-report.md) to the branch. How the comparison is kept fair: [go/benchmark/README.md](go/benchmark/README.md).

## References

- <a id="ref-becher2014"></a>Becher MA, Grimm V, Thorbek P, Horn J, Kennedy PJ, Osborne JL (2014) BEEHAVE: a systems model of honeybee colony dynamics and foraging to explore multifactorial causes of colony failure. *Journal of Applied Ecology* 51: 470–482. [doi:10.1111/1365-2664.12222](https://doi.org/10.1111/1365-2664.12222)
- <a id="ref-becher2016"></a>Becher MA, Grimm V, Knapp J, Horn J, Twiston-Davies G, Osborne JL (2016) BEESCOUT: A model of bee scouting behaviour and a software tool for characterizing nectar/pollen landscapes for BEEHAVE. *Ecological Modelling* 340: 126–133. [doi:10.1016/j.ecolmodel.2016.09.013](https://doi.org/10.1016/j.ecolmodel.2016.09.013)
- <a id="ref-becher2018"></a>Becher MA, Twiston-Davies G, Penny TD, Goulson D, Rotheray EL, Osborne JL (2018) Bumble-BEEHAVE: a systems model for exploring multifactorial causes of bumblebee decline at individual, colony, population and community level. *Journal of Applied Ecology* 55: 2790–2801. [doi:10.1111/1365-2664.13165](https://doi.org/10.1111/1365-2664.13165)
- <a id="ref-twiston-davies2021"></a>Twiston-Davies G, Becher MA, Osborne JL (2021) BEE-STEWARD: A research and decision-support software for effective land management to promote bumblebee populations. *Methods in Ecology and Evolution* 12: 1809–1815. [doi:10.1111/2041-210X.13673](https://doi.org/10.1111/2041-210X.13673)

## Related projects

- [BEE-STEWARD](https://beehave-model.net/bee-steward-is-published/): the original NetLogo model.
- [beecs](https://github.com/mlange-42/beecs): the honeybee model [BEEHAVE](https://beehave-model.net/) ([Becher et al. 2014](#ref-becher2014)) translated to Go with Ark. gobees took its layout and several ideas from it. Worth a look.
