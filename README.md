<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/logo/gobees-logo-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="docs/logo/gobees-logo-light.svg">
  <img alt="gobees" src="docs/logo/gobees-logo-light.svg" width="560">
</picture>

[![Tests](https://github.com/peterpothmann/gobees/actions/workflows/tests.yml/badge.svg)](https://github.com/peterpothmann/gobees/actions/workflows/tests.yml)

</div>

`gobees` is a port of [BEE-STEWARD](https://beehave-model.net/bee-steward-is-published/) ([Twiston-Davies et al. 2021](#ref-twiston-davies2021)) from `NetLogo` to `Go`, built on the [Ark](https://github.com/mlange-42/ark) entity component system. `BEE-STEWARD` is a research and decision support tool for land management that promotes bumblebee populations. It combines the bumblebee population model [Bumble-BEEHAVE](https://beehave-model.net/category/bumblebeehave/) ([Becher et al. 2018](#ref-becher2018)) with the landscape module [BEESCOUT](https://beehave-model.net/category/beescout/) ([Becher et al. 2016](#ref-becher2016)). It simulates colonies of six *Bombus* species and the cuckoo bumblebee *Psithyrus*, from queen emergence through foraging and brood rearing to hibernation, over several years. The bees forage in a nectar and pollen landscape built from maps, with the flowering period of each plant species. Agri-environment schemes, such as flower strips, and crop rotations can be applied to selected fields.

All credit for the scientific model, its development, and parameterisation belongs to the original authors and contributors of `BEE-STEWARD` and its component models.

The goal is a faithful translation that preserves the behaviour of the original NetLogo implementation. Any unusual implementation details are documented and retained to maintain consistency with the original model. The goal is a faithful translation: gobees reproduces the NetLogo code, including its quirks, and does not change the model dynamics. Quirks are documented, not fixed

> [!NOTE]
> gobees is work in progress. The model is translated step by step and does not yet run the full bumblebee life cycle. The steps and their status are listed in [go/validation/README.md](go/validation/README.md#steps-and-their-validation).

A graphical interface for gobees is available in [gobees-ui](https://github.com/peterpothmann/gobees-ui). It shows the landscape, the flowering calendar and the queens of a running simulation, and explains how the model works in a schematic animation.

## Two versions

gobees will come in two versions:

- **This version** stays as close as possible to the NetLogo code: same quirks, same variable and procedure names, same structure. It is meant for users who know BEE-STEWARD in NetLogo and want to move to Go, and who need to trace every result back to the NetLogo code.
- **An optimized version** (planned, not started yet) will fix the quirks, rename variables, functions and systems, and slim down the model. It is meant for users who learn BEE-STEWARD from scratch, and for applications where performance matters most. Its results will differ from NetLogo wherever a quirk is fixed (to be checked).

If you do not depend on NetLogo, use the optimized version once it is available.

## How the translation works

Each step follows the same order:

1. The NetLogo procedures of the step are annotated in [netlogo/__BEE-STEWARD_RC_ColourVersionMB.nlogo](netlogo/__BEE-STEWARD_RC_ColourVersionMB.nlogo). Annotations change comments only. Every procedure gets a header, and every step inside it gets a code such as `[UCOL-02]`.
2. Questionable behaviour of the NetLogo code is recorded in [trans/netlogo-findings.md](trans/netlogo-findings.md) and checked in NetLogo.
3. The step is translated to Go. Every system and function names the codes it translates, and every finding gets a test.
4. Where possible, gobees is compared with NetLogo (see [Validation](#validation)).

The reference is written in NetLogo 5.3.1.

## Repository layout


| Folder       | Content                                                         |
| ------------ | --------------------------------------------------------------- |
| `go/model`   | system order of setup and one simulated day                     |
| `go/sys`     | systems, one per NetLogo procedure                              |
| `go/comp`    | components of the entities (bees, colonies, foodsources, ...)   |
| `go/globals` | global state of the model                                       |
| `go/params`  | parameters and the reader for NetLogo parameter files           |
| `go/util`    | functions that mirror NetLogo reporters                         |
| `go/enum`    | castes, stages and activities as Go constants                   |
| `netlogo`    | the annotated NetLogo model and its input files                 |
| `trans`      | findings on the NetLogo code and the validation against NetLogo |

The layout follows [beecs](https://github.com/mlange-42/beecs).

## Usage

Requires Go 1.26 or later.

```shell
cd go
go run .
```

This runs the example farm (`netlogo/_SYSTEM_Example_Farm_Parameters.csv`) with the parts translated so far and prints a short summary. To use a custom parameter file, edit the paths in the Go source code.

Tests:

```shell
cd go
go test ./...          # all tests, incl. the comparison with NetLogo
go test -short ./...   # without the comparison with NetLogo
```

## Validation

Single runs of gobees and NetLogo cannot be compared, because NetLogo's random number sequence cannot be reproduced in Go. Instead, 100 NetLogo runs give quantiles per day for each compared quantity, and a Go test requires the median of 100 gobees runs to lie within the NetLogo 5–95 % band. The NetLogo results are embedded in the repository, so NetLogo does not have to be installed to run the tests.

The CI runs all tests on every pull request and on every push to `main`. In a pull request it commits the comparison report [go/validation/validation-report.md](go/validation/validation-report.md) to the branch. Details on the experiments: [go/validation/README.md](go/validation/README.md).

## References

- <a id="ref-becher2014"></a>Becher MA, Grimm V, Thorbek P, Horn J, Kennedy PJ, Osborne JL (2014) BEEHAVE: a systems model of honeybee colony dynamics and foraging to explore multifactorial causes of colony failure. *Journal of Applied Ecology* 51: 470–482. [doi:10.1111/1365-2664.12222](https://doi.org/10.1111/1365-2664.12222)
- <a id="ref-becher2016"></a>Becher MA, Grimm V, Knapp J, Horn J, Twiston-Davies G, Osborne JL (2016) BEESCOUT: A model of bee scouting behaviour and a software tool for characterizing nectar/pollen landscapes for BEEHAVE. *Ecological Modelling* 340: 126–133. [doi:10.1016/j.ecolmodel.2016.09.013](https://doi.org/10.1016/j.ecolmodel.2016.09.013)
- <a id="ref-becher2018"></a>Becher MA, Twiston-Davies G, Penny TD, Goulson D, Rotheray EL, Osborne JL (2018) Bumble-BEEHAVE: a systems model for exploring multifactorial causes of bumblebee decline at individual, colony, population and community level. *Journal of Applied Ecology* 55: 2790–2801. [doi:10.1111/1365-2664.13165](https://doi.org/10.1111/1365-2664.13165)
- <a id="ref-twiston-davies2021"></a>Twiston-Davies G, Becher MA, Osborne JL (2021) BEE-STEWARD: A research and decision-support software for effective land management to promote bumblebee populations. *Methods in Ecology and Evolution* 12: 1809–1815. [doi:10.1111/2041-210X.13673](https://doi.org/10.1111/2041-210X.13673)

## Related projects

- [BEE-STEWARD](https://beehave-model.net/bee-steward-is-published/): the original NetLogo model.
- [beecs](https://github.com/mlange-42/beecs): the honeybee model [BEEHAVE](https://beehave-model.net/) ([Becher et al. 2014](#ref-becher2014)) translated to Go with Ark. gobees took its layout and several ideas from it. Worth a look.
