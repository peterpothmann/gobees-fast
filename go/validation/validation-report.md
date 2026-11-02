# Validation of gobees-fast against NetLogo

Written by go/compare for every pull request (go/compare/README.md). Each scenario runs 100 times in gobees-fast and in NetLogo (the experiment of the same name in go/validation/experiments) and is compared from day 1 on:

- Band (enforced): on every day, the median of the gobees-fast runs lies within the 5-95 % band of the NetLogo runs.
- KS (reported only): two-sample Kolmogorov-Smirnov test of the yearly maximum and mean of each metric, Bonferroni-corrected per scenario (family alpha 0.05).

The experiments of parts of the model (step3, step4a, step6) are tests in go/sys and not part of this report.

| Scenario | Band: metrics inside on all days | KS: significant tests |
| --- | --- | --- |
| Example farm, 3 years (`step7`) | ✅ 15 of 15 | ✅ 0 of 90 |
| Badgers (`step7_badgers`) | ✅ 16 of 16 | ✅ 0 of 96 |
| Crop rotation (`step7_crop`) | ✅ 16 of 16 | ✅ 0 of 96 |
| Individual-based colonies (`step8_ibm`) | ✅ 17 of 17 | ✅ 0 of 0 |
| Genetics (`step8_genetics`) | ✅ 16 of 16 | ✅ 0 of 64 |
| Several species (`step8_species`) | ✅ 22 of 22 | ✅ 0 of 88 |
| Weather and foraging mortality (`step8_weather`) | ✅ 16 of 16 | ✅ 0 of 64 |
| Other settings (`step8_settings`) | ✅ 17 of 17 | ✅ 0 of 68 |

## Example farm, 3 years (`step7`)

500 B. terrestris queens on the example farm, the full model with the year cycle: young queens leave the nest and hibernate, the season ends, the queens emerge again. 1095 days.

| Metric | Band: days inside | KS |
| --- | --- | --- |
| colonies | ✅ 1095 of 1095 | ✅ 0 of 6 |
| dead_colonies | ✅ 1095 of 1095 | ✅ 0 of 6 |
| colonies_founded | ✅ 1095 of 1095 | ✅ 0 of 6 |
| eggs | ✅ 1095 of 1095 | ✅ 0 of 6 |
| larvae | ✅ 1095 of 1095 | ✅ 0 of 6 |
| pupae | ✅ 1095 of 1095 | ✅ 0 of 6 |
| workers | ✅ 1095 of 1095 | ✅ 0 of 6 |
| males | ✅ 1095 of 1095 | ✅ 0 of 6 |
| young_queens | ✅ 1095 of 1095 | ✅ 0 of 6 |
| hibernating_queens | ✅ 1095 of 1095 | ✅ 0 of 6 |
| energy_kJ | ✅ 1095 of 1095 | ✅ 0 of 6 |
| pollen_g | ✅ 1095 of 1095 | ✅ 0 of 6 |
| trips | ✅ 1095 of 1095 | ✅ 0 of 6 |
| nectar_left_l | ✅ 1095 of 1095 | ✅ 0 of 6 |
| pollen_left_kg | ✅ 1095 of 1095 | ✅ 0 of 6 |

![step7](plots/step7.png)

## Badgers (`step7_badgers`)

As the example farm, with 2 badgers (N_Badgers = 2). They live on the two Scrub patches and dig up each colony within 735 m with p = 0.019 per day. 1095 days.

| Metric | Band: days inside | KS |
| --- | --- | --- |
| colonies | ✅ 1095 of 1095 | ✅ 0 of 6 |
| dead_colonies | ✅ 1095 of 1095 | ✅ 0 of 6 |
| colonies_founded | ✅ 1095 of 1095 | ✅ 0 of 6 |
| eggs | ✅ 1095 of 1095 | ✅ 0 of 6 |
| larvae | ✅ 1095 of 1095 | ✅ 0 of 6 |
| pupae | ✅ 1095 of 1095 | ✅ 0 of 6 |
| workers | ✅ 1095 of 1095 | ✅ 0 of 6 |
| males | ✅ 1095 of 1095 | ✅ 0 of 6 |
| young_queens | ✅ 1095 of 1095 | ✅ 0 of 6 |
| hibernating_queens | ✅ 1095 of 1095 | ✅ 0 of 6 |
| energy_kJ | ✅ 1095 of 1095 | ✅ 0 of 6 |
| pollen_g | ✅ 1095 of 1095 | ✅ 0 of 6 |
| trips | ✅ 1095 of 1095 | ✅ 0 of 6 |
| nectar_left_l | ✅ 1095 of 1095 | ✅ 0 of 6 |
| pollen_left_kg | ✅ 1095 of 1095 | ✅ 0 of 6 |
| badgers | ✅ 1095 of 1095 | ✅ 0 of 6 |

![step7_badgers](plots/step7_badgers.png)

## Crop rotation (`step7_crop`)

As the example farm, with a crop rotation of two maps: on day 365 the landscape is replaced, year 2 runs on `Example_Farm_PlotTest_Foodsources.txt` (36 flower patches), year 3 on the example farm again. NetLogo sets CropRotationList after Setup, gobees-fast reads CropRotationListAsString, which gives the same list [PAR-07]. 1095 days.

| Metric | Band: days inside | KS |
| --- | --- | --- |
| colonies | ✅ 1095 of 1095 | ✅ 0 of 6 |
| dead_colonies | ✅ 1095 of 1095 | ✅ 0 of 6 |
| colonies_founded | ✅ 1095 of 1095 | ✅ 0 of 6 |
| eggs | ✅ 1095 of 1095 | ✅ 0 of 6 |
| larvae | ✅ 1095 of 1095 | ✅ 0 of 6 |
| pupae | ✅ 1095 of 1095 | ✅ 0 of 6 |
| workers | ✅ 1095 of 1095 | ✅ 0 of 6 |
| males | ✅ 1095 of 1095 | ✅ 0 of 6 |
| young_queens | ✅ 1095 of 1095 | ✅ 0 of 6 |
| hibernating_queens | ✅ 1095 of 1095 | ✅ 0 of 6 |
| energy_kJ | ✅ 1095 of 1095 | ✅ 0 of 6 |
| pollen_g | ✅ 1095 of 1095 | ✅ 0 of 6 |
| trips | ✅ 1095 of 1095 | ✅ 0 of 6 |
| nectar_left_l | ✅ 1095 of 1095 | ✅ 0 of 6 |
| pollen_left_kg | ✅ 1095 of 1095 | ✅ 0 of 6 |
| foodsources | ✅ 1095 of 1095 | ✅ 0 of 6 |

![step7_crop](plots/step7_crop.png)

## Individual-based colonies (`step8_ibm`)

COLONIES_IBM = 10: the first 10 colonies are individual-based [CCOL-05]. 730 days.

| Metric | Band: days inside | KS |
| --- | --- | --- |
| colonies | ✅ 730 of 730 | – no data |
| dead_colonies | ✅ 730 of 730 | – no data |
| colonies_founded | ✅ 730 of 730 | – no data |
| eggs | ✅ 730 of 730 | – no data |
| larvae | ✅ 730 of 730 | – no data |
| pupae | ✅ 730 of 730 | – no data |
| workers | ✅ 730 of 730 | – no data |
| males | ✅ 730 of 730 | – no data |
| young_queens | ✅ 730 of 730 | – no data |
| hibernating_queens | ✅ 730 of 730 | – no data |
| energy_kJ | ✅ 730 of 730 | – no data |
| pollen_g | ✅ 730 of 730 | – no data |
| trips | ✅ 730 of 730 | – no data |
| nectar_left_l | ✅ 730 of 730 | – no data |
| pollen_left_kg | ✅ 730 of 730 | – no data |
| ibm_colonies | ✅ 730 of 730 | – no data |
| bee_agents | ✅ 730 of 730 | – no data |

![step8_ibm](plots/step8_ibm.png)

## Genetics (`step8_genetics`)

SexLocus? = true (diploid males), UnlimitedMales? = false (a queen without a male dies). 730 days.

| Metric | Band: days inside | KS |
| --- | --- | --- |
| colonies | ✅ 730 of 730 | ✅ 0 of 4 |
| dead_colonies | ✅ 730 of 730 | ✅ 0 of 4 |
| colonies_founded | ✅ 730 of 730 | ✅ 0 of 4 |
| eggs | ✅ 730 of 730 | ✅ 0 of 4 |
| larvae | ✅ 730 of 730 | ✅ 0 of 4 |
| pupae | ✅ 730 of 730 | ✅ 0 of 4 |
| workers | ✅ 730 of 730 | ✅ 0 of 4 |
| males | ✅ 730 of 730 | ✅ 0 of 4 |
| young_queens | ✅ 730 of 730 | ✅ 0 of 4 |
| hibernating_queens | ✅ 730 of 730 | ✅ 0 of 4 |
| energy_kJ | ✅ 730 of 730 | ✅ 0 of 4 |
| pollen_g | ✅ 730 of 730 | ✅ 0 of 4 |
| trips | ✅ 730 of 730 | ✅ 0 of 4 |
| nectar_left_l | ✅ 730 of 730 | ✅ 0 of 4 |
| pollen_left_kg | ✅ 730 of 730 | ✅ 0 of 4 |
| diploid_males | ✅ 730 of 730 | ✅ 0 of 4 |

![step8_genetics](plots/step8_genetics.png)

## Several species (`step8_species`)

All 7 species: 150 queens of each Bombus species and 100 Psithyrus queens. 730 days.

| Metric | Band: days inside | KS |
| --- | --- | --- |
| colonies | ✅ 730 of 730 | ✅ 0 of 4 |
| dead_colonies | ✅ 730 of 730 | ✅ 0 of 4 |
| colonies_founded | ✅ 730 of 730 | ✅ 0 of 4 |
| eggs | ✅ 730 of 730 | ✅ 0 of 4 |
| larvae | ✅ 730 of 730 | ✅ 0 of 4 |
| pupae | ✅ 730 of 730 | ✅ 0 of 4 |
| workers | ✅ 730 of 730 | ✅ 0 of 4 |
| males | ✅ 730 of 730 | ✅ 0 of 4 |
| young_queens | ✅ 730 of 730 | ✅ 0 of 4 |
| hibernating_queens | ✅ 730 of 730 | ✅ 0 of 4 |
| energy_kJ | ✅ 730 of 730 | ✅ 0 of 4 |
| pollen_g | ✅ 730 of 730 | ✅ 0 of 4 |
| trips | ✅ 730 of 730 | ✅ 0 of 4 |
| nectar_left_l | ✅ 730 of 730 | ✅ 0 of 4 |
| pollen_left_kg | ✅ 730 of 730 | ✅ 0 of 4 |
| colonies_B_terrestris | ✅ 730 of 730 | ✅ 0 of 4 |
| colonies_B_pascuorum | ✅ 730 of 730 | ✅ 0 of 4 |
| colonies_B_lapidarius | ✅ 730 of 730 | ✅ 0 of 4 |
| colonies_B_hortorum | ✅ 730 of 730 | ✅ 0 of 4 |
| colonies_B_pratorum | ✅ 730 of 730 | ✅ 0 of 4 |
| colonies_B_hypnorum | ✅ 730 of 730 | ✅ 0 of 4 |
| psithyrus | ✅ 730 of 730 | ✅ 0 of 4 |

![step8_species](plots/step8_species.png)

## Weather and foraging mortality (`step8_weather`)

Weather = "foragingHoursExample" (foraging hours per day of the year), ForagingMortalityModel = "low". 730 days.

| Metric | Band: days inside | KS |
| --- | --- | --- |
| colonies | ✅ 730 of 730 | ✅ 0 of 4 |
| dead_colonies | ✅ 730 of 730 | ✅ 0 of 4 |
| colonies_founded | ✅ 730 of 730 | ✅ 0 of 4 |
| eggs | ✅ 730 of 730 | ✅ 0 of 4 |
| larvae | ✅ 730 of 730 | ✅ 0 of 4 |
| pupae | ✅ 730 of 730 | ✅ 0 of 4 |
| workers | ✅ 730 of 730 | ✅ 0 of 4 |
| males | ✅ 730 of 730 | ✅ 0 of 4 |
| young_queens | ✅ 730 of 730 | ✅ 0 of 4 |
| hibernating_queens | ✅ 730 of 730 | ✅ 0 of 4 |
| energy_kJ | ✅ 730 of 730 | ✅ 0 of 4 |
| pollen_g | ✅ 730 of 730 | ✅ 0 of 4 |
| trips | ✅ 730 of 730 | ✅ 0 of 4 |
| nectar_left_l | ✅ 730 of 730 | ✅ 0 of 4 |
| pollen_left_kg | ✅ 730 of 730 | ✅ 0 of 4 |
| foraging_period_s | ✅ 730 of 730 | ✅ 0 of 4 |

![step8_weather](plots/step8_weather.png)

## Other settings (`step8_settings`)

WinterMortality?, MergeHedges?, MinSizeFoodSources?, RemoveEmptyFoodSources?, KeepDeadColonies? = false, MaxHibernatingQueens = 150 (the limit takes effect on every 1 January, also on the first day). 730 days.

| Metric | Band: days inside | KS |
| --- | --- | --- |
| colonies | ✅ 730 of 730 | ✅ 0 of 4 |
| dead_colonies | ✅ 730 of 730 | ✅ 0 of 4 |
| colonies_founded | ✅ 730 of 730 | ✅ 0 of 4 |
| eggs | ✅ 730 of 730 | ✅ 0 of 4 |
| larvae | ✅ 730 of 730 | ✅ 0 of 4 |
| pupae | ✅ 730 of 730 | ✅ 0 of 4 |
| workers | ✅ 730 of 730 | ✅ 0 of 4 |
| males | ✅ 730 of 730 | ✅ 0 of 4 |
| young_queens | ✅ 730 of 730 | ✅ 0 of 4 |
| hibernating_queens | ✅ 730 of 730 | ✅ 0 of 4 |
| energy_kJ | ✅ 730 of 730 | ✅ 0 of 4 |
| pollen_g | ✅ 730 of 730 | ✅ 0 of 4 |
| trips | ✅ 730 of 730 | ✅ 0 of 4 |
| nectar_left_l | ✅ 730 of 730 | ✅ 0 of 4 |
| pollen_left_kg | ✅ 730 of 730 | ✅ 0 of 4 |
| foodsources | ✅ 730 of 730 | ✅ 0 of 4 |
| queens_removed | ✅ 730 of 730 | ✅ 0 of 4 |

![step8_settings](plots/step8_settings.png)

Plots: blue, NetLogo median and 5-95 % band; red, gobees median; grey dashed, gobees-fast median; x axis: day.
