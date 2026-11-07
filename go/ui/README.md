# Window of `gobees --ui`

`gobees --ui` shows one run in a window while it runs. Built with [Ebitengine](https://ebitengine.org), like
[beecs-ui](https://github.com/mlange-42/beecs-ui). It started as a copy of
[gobees-ui](https://github.com/peterpothmann/gobees-ui), the interface of gobees.

```shell
cd go
go run . --ui                                         # example farm
go run . --ui -c compare/scenarios/step8_species.yaml -o
```

## The run

The run is the normal run of the command line: the run file and the flags (`-c`, `-s`, `--days`, `-o`, `-f`,
`--columns`) define it, the output and the run info are written as without the window. The window only decides
when the next day runs (play, step, speed). It calls `Initialize`, one `Update` per day and `Finalize` of the
same app as `a.Run` does, so the output is byte for byte the same.

- One run only: one run file, `runs` 1, no `variations`. Otherwise `--ui` is an error.
- Closing the window ends the run on that day. The output then holds the days up to then, the run info was
  written before the run.
- Platforms: on Windows the window is always built in. On macOS it needs cgo (the Xcode command line tools). On
  Linux it is built only with the tag `ui` (`go run -tags ui . --ui`), because Ebitengine needs cgo and the X11
  and OpenGL headers there. Without the tag, gobees on Linux stays pure Go (CI, benchmark, cluster) and `--ui`
  fails with a message.

## What it shows

* **Colonies:** every colony as a honeycomb cell at its nest, the larger, the more bees (brood included, full
  size from 400 bees). New colonies get a ring. Point at a colony to see its foraging range and the flower patches
  in flower within it (`comp.ColonyRange`), lines to the patches its bees know for nectar (yellow) and pollen
  (blue) (`comp.Forager.KnownNectarPatches`, `KnownPollenPatches`), its brood, adults and stores, and its life so
  far. Click a colony to follow it: the plots then show this colony since its foundation. `Esc` or a click on the
  map returns to all colonies. With more than one species, each species has its own colour.

  The model keeps no colony numbers and no day of foundation. The window numbers the colonies in the order it
  first sees them and takes that day as the day of foundation. It reads the model at the end of every day, so
  this is the day the colony was founded.
* **Cuckoo bumblebees:** a colony with a queen of a parasitic species (`parasitic` in the species of the run file)
  in its nest has a violet ring and dot. Cuckoo bees in a colony are counted apart from the colony's own bees.
* **Badgers:** with the range in which they dig up colonies (`constants.badgers.range`).
* **Model output per day:** nectar and pollen left in the landscape at the end of the day; queens hibernating,
  queens searching a nest and colonies; eggs, larvae and pupae; workers, males and queens in colonies; energy and
  pollen stores of all colonies and the foraging trips of the day (`obs.Counters`). Vertical lines mark when the
  queens of each species emerge on average (E) and when their season ends (S, `seasonStop`).
* **Going back:** `←` shows the day before, held down it goes back, up to 365 days. `→`, Step or Play go forward
  through the kept days first and run the model again once the latest day is reached. The model itself cannot run
  backwards: the window keeps what it showed on each day. After a crop rotation the days before it can no longer
  be shown.
* **Map:** one circle per flower patch, of the area of the patch, filled with the colour of its habitat type
  (landscape files hold no colours: the NetLogo colours of the example farm, Sussex and the NetLogo habitats file,
  grey for other types). The yellow fill shows how many of its flower species are in flower today.
* **Test nest:** right-click the map to place a nest. The dashed circle is the foraging range
  (`constants.foraging.maxForagingRange`). Lines run to the edge of each reachable patch in flower: the stronger
  the line, the more likely a searching bee finds it (`exp(detectionLambda × distance to the edge)`). The window
  computes this itself, the test nest is not part of the model.
* **Patch info:** its flower species (`comp.Foodsource.Flower`) with their flowering periods, nectar and pollen
  left at the end of the day, and the colonies whose bees know the patch. The model counts no visits per patch.
* **How the model works:** the button opens a second window (`gobees --ui-explain`) with a schematic animation of
  the processes in BEE-STEWARD, "A year" and "A day". It does not run the model and names the NetLogo procedures.

| Key | Action |
| --- | --- |
| `Space` | play or pause |
| `→` | one day, held down: runs at the selected speed |
| `←` | one day back, held down: goes back |
| `↑`, `↓` | speed |
| `Esc` | clear selection and test nest |

Mouse: point at or click a patch, click a colony to follow it, right-click to place a test nest, double right-click
to remove it.

## Checks without clicking

`--ui-screenshot PATH` (hidden flag, with `--ui`) runs to the end at once, saves a screenshot of the window and
closes it. The tests (`go test ./ui/`, on Linux with `-tags ui`) step the example farm as the window does.
