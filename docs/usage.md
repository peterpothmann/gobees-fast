# Usage of gobees-fast in detail

The short version is in the [README](../README.md#quick-start). This file is the reference: look up a section when you need it.


| Section                                           | Read it to ...                                              |
| ------------------------------------------------- | ----------------------------------------------------------- |
| [Command-line arguments](#command-line-arguments) | start a run and choose seed, days, runs and output          |
| [Run files](#run-files)                           | write or change a run file (parameters, species, landscape) |
| [NetLogo input files](#netlogo-input-files)       | turn NetLogo input files into a gobees-fast run file        |
| [Daily output](#daily-output)                     | read the result file of a run                               |
| [Several runs](#several-runs)                     | run the model many times, e.g. for statistics               |
| [Experiments](#experiments)                       | vary parameters and run every combination                   |
| [Parts on a cluster](#parts-on-a-cluster)         | split the runs over the jobs of an array                    |
| [Output](#output)                                 | find the output files of a run                              |
| [Provenance](#provenance)                         | repeat or document a run exactly                            |

## Command-line arguments

The flags control a single call of the model: which run file runs, for how long, how often and where the output goes. Flags override the section `run` of the run file.

The flags of the model (`go/`, `go run .`). The flags of the converter are in [NetLogo input files](#netlogo-input-files).


| Flag                 | Short | Default                                    | Meaning                                                                                                                                                                                                                                                                                         |
| -------------------- | ----- | ------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--config PATH`      | `-c`  | the example farm (built in)                | run file (YAML). Its landscape files are relative to its folder. Repeat`-c` for several run files ([experiments](#experiments)).                                                                                                                                                                |
| `--seed N`           | `-s`  | `run.seed` of the run file                 | random seed of the first repetition, 1 or more (NetLogo:`RAND_SEED`).                                                                                                                                                                                                                           |
| `--days N`           |       | `run.days` of the run file, else `365`     | number of days to run.`0`: no limit, the run ends only when the population is extinct and `stopExtinct` is true.                                                                                                                                                                                |
| `--repetitions N`    |       | `run.repetitions` of the run file          | number of runs of each combination of an [experiment](#experiments) (of the run file without variations), 1 or more, in one process. Repetition i has the seed `SEED + i - 1`. All runs of a run file go into one output file. |
| `--threads T`        |       | `run.threads` of the run file              | number of runs at the same time, 1 or more. Each run uses one CPU core, so set it to the number of cores you want to use (on a cluster: the CPUs of the job, e.g.`--threads $SLURM_CPUS_PER_TASK`). `1`: all runs one after the other, also across run files. The output does not depend on it. |
| `--part I/N`         |       | `run.part` of the run file, else all runs  | run only the part I of N of the runs, 1 <= I <= N, e.g. `--part $SLURM_ARRAY_TASK_ID/50` ([parts on a cluster](#parts-on-a-cluster)). The output goes into `DIR/<name>/part_I`. |
| `--output DIR`       | `-o`  | no output                                  | write the [output](#output) of each run file into the folder `DIR/<name>` (`name`: the key of the run file): the daily output, the [provenance](#provenance) and the copies of the landscape files. Missing folders are created.                                                               |
| `--columns A,B,C`    |       | `output.columns` of the run file, else all | columns of the daily output, in this order (needs`-o`). An unknown name is an error that lists the columns.                                                                                                                                                                                     |
| `--list-columns`     |       |                                            | print the columns of the daily output with a description and exit.                                                                                                                                                                                                                              |
| `--help`             | `-h`  |                                            | show all flags.                                                                                                                                                                                                                                                                                 |

```shell
go run . -c path/to/run.yaml -s 3 --days 730       # own run file, seed 3, 2 years
go run . -o out                                    # output to out/example_farm/daily.csv
go run . -c out/example_farm/provenance.yaml -o again  # repeat the run of out/example_farm
go run . -o out --columns colonies,workers,eggs    # only these columns (and ticks)
go run . -c a.yaml -c b.yaml -o out --threads 1    # two run files, one run after the other: out/<name of a>, out/<name of b>
```

## Run files

A run file (YAML) describes a run completely: parameters, species, flowers, landscape and run settings. The model reads nothing else, so the run file and its landscape files are all you need to run or share it.

A run file has the sections below. All keys are in lowerCamelCase, values are in the unit of the model (e.g. seconds). [names.md](names.md) lists every key with its unit and the NetLogo name.


| Key                  | Content                                                                                                                                                                                                                                                                          | NetLogo                                                                         |
| -------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| `name`               | name of the run file, the folder of its [output](#output). Needed, only letters, digits, `-` and `_`. The converter writes the name of its flag `--name`. | |
| `run`                | `seed`, `days`, `repetitions` and `threads` (the flags `-s`, `--days`, `--repetitions` and `--threads` override them). Needed: `seed`, `repetitions` and `threads`, each 1 or more. Optional `part` (`I/N`, the flag `--part` overrides it). The converter writes seed 1, 365 days, 1 repetition and 1 thread. In a [provenance](#provenance) also the record of the run. | `RAND_SEED`                                                                     |
| `variations`         | optional: keys and their values, or the [sites](#sites), every combination runs (an [experiment](#experiments))                                                                                                                                                                  |                                                                                 |
| `parameters`         | the model parameters                                                                                                                                                                                                                                                             | the columns of the parameter file                                               |
| `output`             | output options (`keepDeadColonies`, `columns`, `visits`), change no model value                                                                                                                                                                                                  | `KeepDeadColonies?`                                                             |
| `constants`          | the values that NetLogo hard-codes, every key needed (see below)                                                                                                                                                                                                                 | `ParametersProc` and local values of the procedures                             |
| `initialQueens`      | species and number of initial queens                                                                                                                                                                                                                                             | `BeeSpeciesInitialQueensListAsString`                                           |
| `foragingPeriod`     | foraging period [s], 365 values per year, one or more years, the years repeat                                                                                                                                                                                                    | `Weather`                                                                       |
| `stewardshipOptions` | names of the stewardship options                                                                                                                                                                                                                                                 | from the`HabitatsFile`                                                          |
| `species`            | bee species, found by`name` (only those of `initialQueens` are used), `parasitic: true` for a cuckoo bumblebee                                                                                                                                                                   | rows of the`BeespeciesFile` (columns `species` = `name`, Psithyrus by its name) |
| `flowers`            | flower species, without the key no flower values                                                                                                                                                                                                                                 | rows of the`FlowerspeciesFile` (without `proteinPollenProp`)                    |
| `landscape`          | landscape file of the start                                                                                                                                                                                                                                                      | foodsources file of`MyMap`                                                      |
| `cropRotation`       | landscape files of the crop rotation (fewer than 2: none)                                                                                                                                                                                                                        | `CropRotationListAsString`                                                      |

The crop rotation changes the landscape in the morning of day 365 of each year (as NetLogo, [finding 50](../trans/netlogo-findings.md#finding-50)): on day 365 of year 1 to the second file of `cropRotation`, on day 365 of year 2 to the third, and so on. The first file comes again only after one full cycle. The first year runs on `landscape`. So the cycle is right only if `landscape` is the first file of `cropRotation`:

| `landscape: a`, `cropRotation: [a, b, c]` | landscape |
| --- | --- |
| year 1, days 1 to 364 | `a` |
| from day 365 of year 1 | `b` |
| from day 365 of year 2 | `c` |
| from day 365 of year 3 | `a`, then again `b`, `c`, ... |

With `landscape: x` and the same list, the run has `x, b, c, a, b, c, ...`: `x` never comes back. gobees-fast does not check this.

A landscape file (YAML) holds the names of its tables (CSV), relative to the folder of the landscape file. All files are read at the start (package `input`), also the landscapes of a crop rotation. An unknown key or column is an error. A missing key keeps its zero value, except in `constants`. Unlike NetLogo, a missing file is an error at the start, also a landscape of the crop rotation (NetLogo: on its day 365).

| Key            | Table                                                       | Columns                                    |
| -------------- | ----------------------------------------------------------- | ------------------------------------------ |
| `patches`      | the flower patches (needed)                                 | `id,patchType,x,y,area`                    |
| `habitats`     | the flower list of each patch type                          | `patchType,flower,abundance`               |
| `patchFlowers` | own flower lists of single patches (optional)               | `id,flower,abundance`                      |

- `patches`: one row per flower patch, the columns of the NetLogo foodsources file without `flowerSpeciesList` and the GUI columns. `id` an integer, `x`, `y` the centre in metres, `area` in m². The landscape has no edge: the map area of the output is the rectangle around all flower patch circles.
- `habitats`: the flower list of a patch type are its rows, in their order (`abundance`: the relative abundance of `flowerSpeciesList`). One habitat table can serve many landscapes, e.g. `habitats: ../habitats.csv`.
- `patchFlowers`: if the table has rows of a patch (`id` of `patches`, unique), these rows are its flower list instead of the list of its patch type.
- Each table starts with a header of all its columns, in any order. From R: `write.csv(patches, "patches.csv", row.names = FALSE)`, without `row.names = FALSE` the extra first column is an error. Numbers with `.` as decimal point.

```yaml
# _SYSTEM_Example_Farm_Foodsources.yaml
patches: _SYSTEM_Example_Farm_Foodsources_patches.csv
habitats: _SYSTEM_Example_Farm_Foodsources_habitats.csv
```

The converter multiplies `xcor`, `ycor` by the metres per patch of the foodsources file (line 1) and rounds to millimetres. It writes the habitat list of a patch type as the most frequent list of its patches and every other list into `patchFlowers`. It leaves out the stewardship options of `flowerSpeciesList` (items without a value, e.g. `["Plot"]`). In NetLogo they mark the flower species after them as growing in the option, which only the GUI reads (map symbols, areas of the options). An option must be in the `stewardshipOptions` of the HabitatsFile, as in NetLogo. A flower patch with only options is left out, as NetLogo removes it at the setup, and the converter reports it.

The section `constants` holds the values that NetLogo sets in its code and that NetLogo's parameter files cannot change: the values of `ParametersProc` and the local values of the procedures (e.g. the pause of a bee, the probabilities of the badgers and the cuckoo bumblebees, the task thresholds). It is grouped by part of the model (`activities`, `stimuli`, `foraging`, `colony`, `bees`, `psithyrus`, `badgers`, `landscape`, `initialQueens`, `thresholds`) and must hold every key. A run file without a key of `constants` is an error, so that no value of a run comes from the code unseen. The converter writes the NetLogo values (`params.DefaultConstants`) with unit and source as a comment at each key. The coefficients of fitted equations (handling time, competition point, winter survival) stay in the code.

## NetLogo input files

gobees-fast does not read the NetLogo input files. To run NetLogo input files, convert them once into a run file with [tools/netlogo2gobees](../tools/netlogo2gobees). The converter is its own Go module, the model does not import it.

The converter has two commands:

- `convert` turns NetLogo input files into a run file: the parameter file, the files it names, and the map (`MyMap`, a setting of the NetLogo interface), which gives the foodsources file of the start. This is the command for users.
- `generate` is for developers of gobees-fast. The repository contains run files made from NetLogo files: the built-in example farm (`go/input/examplefarm`) and the scenarios of the step check (`go/compare/scenarios`). `generate` writes them all again. Call it after a change of the NetLogo files, the scenarios or the run file format. A test fails if these run files are out of date.

The flags of the converter (`tools/netlogo2gobees`, `go run . <command>`):


| Command    | Flag                | Short | Default                    | Meaning                                                                                                                                |
| ---------- | ------------------- | ----- | -------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `convert`  | `--parameters PATH` | `-p`  | (required)                 | NetLogo parameter file (`MyParametersFile`).                                                                                           |
| `convert`  | `--dir PATH`        | `-d`  | `../../netlogo`            | folder of the NetLogo input files (the folder of the`.nlogo`). The file names in the parameter file are relative to it.                |
| `convert`  | `--map NAME`        | `-m`  | `_SYSTEM_Example_Farm.png` | map (`MyMap`). The foodsources are read from `<NAME without .png>_Foodsources.txt` in `--dir`.                                         |
| `convert`  | `--out PATH`        | `-o`  | (required)                 | folder of the run file and its landscape files with their tables, created if missing.                                                  |
| `convert`  | `--name NAME`       | `-n`  | (required)                 | key`name` of the run file, the folder of its output in gobees-fast. Letters, digits, `-` and `_`.                                      |
| `convert`  | `--run NAME`        |       | `run.yaml`                 | name of the run file in`--out`.                                                                                                        |
| `generate` | `--root PATH`       |       | `../..`                    | folder of the gobees-fast repository, where`generate` writes its files. The default fits when you call it from `tools/netlogo2gobees`. |

```shell
cd tools/netlogo2gobees
# example farm: parameter file and output folder are required, the map is the example farm by default
go run . convert -p ../../netlogo/_SYSTEM_Example_Farm_Parameters.csv -o farm -n farm
# other map: writes sussex/run.yaml and the landscape file with its tables next to it
go run . convert -p ../../netlogo/_SYSTEM_Sussex1_Parameters.csv -m _SYSTEM_Sussex1.png -o sussex -n sussex
# own NetLogo files in another folder, writes mine/my_run.yaml
go run . convert -d path/to/netlogo -p path/to/netlogo/my_Parameters.csv -m my_map.png -o mine -n mine --run my_run.yaml
# all flags of a command
go run . convert --help
# developers: renew the run files of the repository
go run . generate
```

`-p` is a path from the folder you call the converter in. The files named in the parameter file and the map are looked up in `-d`. Run the converted run file with the model:

```shell
cd ../../go
go run . -c ../tools/netlogo2gobees/sussex/run.yaml -o output  # writes output/sussex/
```

What the converter does with the NetLogo values:

- `Weather` is either one of NetLogo's constant options (`"Constant 8 hrs"`, ...: 365 equal values) or the name of a weather file in the folder of the NetLogo files: a CSV table with the columns `Day` (1 to 365) and `ForagingPeriod_s` (0 to 86400), and optionally `Year` (1, 2, ...). NetLogo's example year `"foragingHoursExample"`, a list in the NetLogo code, is the file [netlogo/_SYSTEM_Weather_foragingHoursExample.csv](../netlogo/_SYSTEM_Weather_foragingHoursExample.csv).
- Errors: a missing parameter file (NetLogo loads `_SYSTEM_Parameters.csv` instead), `COLONIES_IBM` other than 0, `MapAreaIncluded` other than `"complete"` (a part of a map is its own landscape file), and a species row whose columns `species` and `name` differ.
- The converter prints every column it does not take over, with the reason.
- NetLogo's parameter files set `MaxForagingRange_m` = 10000, which has no effect in NetLogo: NetLogo uses the constant `ForagingRangeMax_m` = 758 m. The converter writes 758 and says so, in its output and in the comment at `constants.foraging.maxForagingRange`.

## Daily output

The daily output is the result of a run: the state of the bee population on every day, for analysis in R, Python or a spreadsheet.

The daily output (`daily.csv` in the [output](#output) folder) is a tab-separated file with one row per run and day, holding the state at the end of the day. It holds the totals of `OutputDailyProc` (colonies, queens, males, mean weights, ...), followed by brood, workers, stores, foraging trips, forager deaths, queens produced, the mean competition point date and the adults produced (`obs.DailyOutputExtended`). The adult queens are `young_queens + TotalMatedQueens`: every mated queen is an adult.


| Column   | Content                                                 |
| -------- | ------------------------------------------------------- |
| `run`    | number of the run, from 1 over all combinations of an [experiment](#experiments) |
| `seed`   | seed of the run                                         |
| `t`      | days from 0, as the CSV reporter of ark-tools counts    |
| `ticks`  | NetLogo ticks, 1 on the first day                       |
| the rest | the totals of`OutputDailyProc` keep their NetLogo names |

`go run . --list-columns` prints all columns with a description. The key `output.columns` of the run file (a list of names) or the flag `--columns a,b,c` (which overrides the key) selects columns and their order. `ticks` is always the first column after `t`: it is added if missing and moved to the front. Without a selection, all columns are written. An unknown or repeated name is an error that lists the available columns, also when the run has no output. Every value is still computed, so the values of the selected columns do not depend on the selection.

## Visits output

The visits output tells which foodsources the foragers use: the visits per foodsource and day. It replaces the visits file of NetLogo (`BBCT_WriteNPvisitsProc`).

With the key `visits: true` in the section `output` of the run file, gobees-fast writes `visits.csv` into the [output](#output) folder, next to the daily output. It is a tab-separated file with one row per run, day and foodsource with at least one visit. A foodsource without visits on a day has no row.

| Column          | Content                                                                                                   |
| --------------- | --------------------------------------------------------------------------------------------------------- |
| `run`, `seed`, the varied keys, `t` | as in the [daily output](#daily-output)                                               |
| `ticks`         | NetLogo ticks, 1 on the first day                                                                         |
| `landscape`     | the landscape file of the foodsource, named as its copy in the output folder. With [sites](#sites), which are not copied, its name relative to the folder of the run file |
| `id`            | the id of its flower patch in the landscape file. Merged hedges keep the id of the largest hedge          |
| `flowerSpecies` | its flower species                                                                                        |
| `nectarVisits`  | visits that took nectar on this day                                                                       |
| `pollenVisits`  | visits that took pollen on this day                                                                       |

A foodsource is one flower species on one flower patch, so `landscape`, `id` and `flowerSpecies` together name it. A visit counts every bee of a cohort, as in NetLogo. A trip to an empty foodsource is no visit. Unlike NetLogo, the visits are counted per day, not summed since the start, so `CumulVisitsOnlyLastYear?` is not needed: the sum over days gives NetLogo's numbers. Area, position and patch type of a foodsource are in the landscape file, joined by `id`. The rows of a day are sorted by landscape, id and flower species.

## Several runs

The model is stochastic: two seeds give two different results. To see the spread of the results, run the same run file with many seeds.

With `--repetitions N` (or `run.repetitions`), gobees-fast runs the model N times in one process, at most `--threads` runs at the same time (or `run.threads`). Repetition i has the seed `seed + i - 1`, so the same `seed` and `repetitions` always give the same runs. Seed, repetitions and threads have no special value: each must be 1 or more, nothing is drawn at random or taken from the machine. The output holds the runs one after the other, in the order of the runs, so the file does not depend on the number of threads. The runs share the input read at the start, the model does not change it.

## Experiments

An experiment answers questions such as "how does the number of badgers change the population?". It replaces BehaviorSpace of NetLogo.

The optional section `variations` makes a run file an experiment: each listed key takes each of its values, and every combination (full factorial) runs `run.repetitions` times.

```yaml
run:
  seed: 1
  days: 365
  repetitions: 10  # per combination
  threads: 8
variations:
  - key: parameters.badgers
    values: [0, 2, 4]
  - key: constants.foraging.maxForagingRange
    values: [500, 758, 1000]
parameters:
  ...
```

This gives 3 × 3 combinations × 10 repetitions = 90 runs. Rules:

- `key` is the path of a key in the run file, its parts separated by dots. A number selects an item of a list, e.g. `initialQueens.0.queens` (the first species) or `species.6.name`. A value is written as in the run file: a number, a text, a list.
- `values` lists every value. There are no ranges, so every value of an experiment is written out.
- Every combination is the run file with these values, read and checked like a run file: an unknown key, a fraction in a whole number and so on are errors at the start, naming the combination. After the check, gobees-fast keeps only the values of each combination. A combination is read again when its first run starts, and dropped when its last run ends. So the memory depends on `--threads`, not on the number of combinations.
- Every combination runs with the same seeds (`seed` to `seed + repetitions - 1`), so combinations differ only in the varied values, not in the random numbers.
- `name`, the section `run` (seed, days, repetitions, threads), `output.columns` and `output.visits` cannot vary. `landscape` and `cropRotation` cannot vary as keys: a landscape and its crop rotation vary together, as [sites](#sites).
- The daily output is one file: the columns `run`, `seed`, then one column per varied key (named by the key, holding the value), then `t` and the columns. `run` counts from 1 over all combinations, so each run has its own number: in the example above runs 1 to 10 are the first combination, 11 to 20 the second. The first variation changes slowest. The provenance holds the variations, so it repeats the whole experiment.

### Sites

To run the same parameters on many landscapes, each with its own crop rotation, give each landscape a folder, a site, and list the sites in the run file. The run file then stays as short as with one landscape.

```
experiment.yaml
sites/
  L0001/
    site.yaml
    start.yaml
    year2.yaml
  L0002/
    site.yaml
    start.yaml
```

```yaml
# sites/L0001/site.yaml, paths relative to the folder of the site file
landscape: start.yaml
cropRotation: [start.yaml, year2.yaml]
```

```yaml
# experiment.yaml
name: many_landscapes
variations:
  - sites: sites/*/site.yaml
  - key: parameters.badgers
    values: [0, 2]
parameters:
  ...
```

- A site file sets `landscape` (needed) and `cropRotation` (optional, fewer than 2 files: no crop rotation), as the keys of a run file, relative to its own folder. Any other key is an error. The files of `cropRotation` run in the order of the list, and `landscape` should be its first file (see the rule of the [crop rotation](#run-files)).
- `sites` is a pattern (`*` matches any part of a name, `?` one character, as [path/filepath.Match](https://pkg.go.dev/path/filepath#Match)) or a list of patterns and files, relative to the folder of the run file. The sites run in the order of the list, the files of one pattern sorted by name. A pattern without a file and a file matched twice are errors.
- The name of a site is the name of its folder (`L0001`): letters, digits, `-` and `_`, each name once. The daily output holds it in the column `site`.
- The run file has no `landscape` and no `cropRotation` of its own. One `sites` per run file. It combines with the other variations like a key: the example above gives 2 sites × 2 values = 4 combinations, each run `run.repetitions` times.
- All site files are read at the start, and every combination is checked with the landscape files of its site, so an error in a site stops the call before the first run. A combination reads its landscape files again when its first run starts. A landscape file that changed in between is an error.

How the landscapes and crop rotations of the sites are made is up to you (a script, a GIS, the [converter](#netlogo-input-files) per map): gobees-fast reads only the site files and the landscape files.

Several run files run together with `-c a.yaml -c b.yaml`: their runs share one pool of `--threads` threads (the run files must ask for the same `threads`, or the flag sets it). Each run file writes its own [output](#output) into the folder of its key `name`: with the names `farm` and `sussex`, `-o out` gives `out/farm/` and `out/sussex/`. Two run files with the same name are an error, also when only the case differs. The flags `-s`, `--days`, `--repetitions` and `--columns` apply to all run files.

## Parts on a cluster

On a cluster, the runs of a large experiment are split over the jobs of an array: each job runs one part. All jobs start the same call with the same run file, only `--part` differs.

`--part I/N` splits the runs of all combinations (combinations × repetitions, in the order of the [daily output](#daily-output)) into N blocks of consecutive runs and runs only block I. The sizes of the blocks differ by one run at most. Each run keeps its number and seed, so a part is a piece of the call without `--part`:

- 2000 combinations × 10 repetitions = 20000 runs, `--part 3/50`: the runs 801 to 1200.
- The output goes into `DIR/<name>/part_I`, I with as many digits as N (`part_03` of 50), so that all jobs can write into the same `-o DIR` and the folders sort in the order of the parts.
- Every part checks all combinations at the start, as without `--part`. With many sites this takes some time per job (less than 30 ms per site with the landscapes of the example farm).
- The provenance of a part holds `run.part`, so it repeats the part. Without the key, it repeats all runs.
- N can be at most the number of runs, so that no part is empty.

A SLURM array of 50 jobs with 8 cores each (build the program once with `go build -o gobees .` in `go/`):

```shell
#!/bin/bash
#SBATCH --array=1-50
#SBATCH --cpus-per-task=8
./gobees -c experiment.yaml -o out --threads $SLURM_CPUS_PER_TASK --part $SLURM_ARRAY_TASK_ID/50
```

The parts appended in the order of their folders give the daily output of the call without `--part`, byte for byte. Each part has the header, keep it once:

```shell
cd out/many_landscapes
head -n 1 part_01/daily.csv > daily.csv
for f in part_*/daily.csv; do tail -n +2 "$f" >> daily.csv; done
```

## Output

With `-o DIR`, gobees-fast writes the output of each run file into the folder `DIR/<name>`, where `name` is the key `name` of the run file (the example farm: `example_farm`). Missing folders are created, files of an earlier run of the same name are overwritten.

| File               | Content                                                                                                     |
| ------------------ | ----------------------------------------------------------------------------------------------------------- |
| `daily.csv`        | the [daily output](#daily-output): all runs of the run file, one row per run and day                        |
| `visits.csv`       | the [visits output](#visits-output), only with `output.visits: true`                                        |
| `provenance.yaml`  | the [provenance](#provenance): the run file as run, it repeats the run                                      |
| landscape files    | copies of the landscape files of the run and their tables, with their names (see [Provenance](#provenance)). Not for [sites](#sites). |

The folder holds everything needed to repeat the run and can be moved or archived as a whole.

## Provenance

The provenance records everything needed to repeat a run: the run file, the landscapes, the seed, the command and the program version. Keep it with the output, for example for a paper or to check a result later.

With an output (`-o DIR`), gobees-fast writes the provenance `provenance.yaml` into the output folder, before the run. The provenance is a run file that repeats the run: `go run . -c out/example_farm/provenance.yaml -o again` writes the same daily output into `again/example_farm/daily.csv`, byte for byte. It holds all sections of the run file as read (`name` and the constants included) and the section `run`:


| Key                      | Content                                                                                                                                                                                                   |
| ------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `seed`                   | the seed of the first repetition                                                                                                                                                                          |
| `days`                   | the number of days                                                                                                                                                                                        |
| `repetitions`, `threads` | the numbers of repetitions and threads                                                                                                                                                                    |
| `part`                   | with [`--part`](#parts-on-a-cluster): the part of the runs, `I/N`                                                                                                                                        |
| `started`                | start time (RFC 3339)                                                                                                                                                                                     |
| `command`                | command line                                                                                                                                                                                              |
| `version`                | module, version, VCS revision, its time and`modified` (changes against the revision), Go version. The VCS keys are only known in a binary built with `go build` inside the repository, not with `go run`. |
| `config`                 | the run file given with`-c`, or `built-in example farm`                                                                                                                                                   |
| `output`                 | the daily output file. Its columns are the key`columns` of the section `output` of the provenance, always written in full                                                                                 |
| `landscapes`             | per landscape file and table: the copy (`file`), the name in the run file or, for a table, relative to the folder of the run file (`source`) and the SHA-256 of the source                               |
| `sites`                  | with [sites](#sites): per site its name (`site`), its site file (`file`) and per landscape file the name in the run file (`source`) and the SHA-256 at the start                                          |

The landscape files and their tables are copied into the output folder and keep their names (`_SYSTEM_Example_Farm_Foodsources.yaml`). A second file of the same name from another folder gets `_2` in front of the extension, as does a file named `daily.csv`, `visits.csv` or `provenance.yaml`. A table of several landscape files is copied once. The tables are copied byte for byte. A landscape file is copied byte for byte if its tables keep their names, else it is written again with the names of the copies (its SHA-256 is the one of the source). The keys `landscape` and `cropRotation` of the provenance name the copies.

With sites, the provenance copies no landscape files: 1000 sites would copy 1000 landscapes into every output. Its key `sites` lists the site files that ran, relative to the output folder, instead of the pattern of the run file. So it repeats the same sites, also when the folder of the sites gets more files later, as long as the sites stay in their place. The SHA-256 in `run.sites` shows whether a landscape file changed since the run.

Only `seed`, `days`, `repetitions`, `threads` and `part` of the section `run` are read, the other keys are a record.
