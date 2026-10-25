## Steps 1 to 3: colony foundation in year 1

Experiment `go/validation/experiments/step3.xml`: example farm, 500 B. terrestris queens, 250 days. Setup values and queen traits are compared at step 0 only (later, NetLogo colonies die with their queen, which is not yet translated). Colony foundation and queen counts are compared at every step.

100 runs each. Every metric is checked on each of its days (all days of the experiment, unless the description above says otherwise): the median of the gobees runs on a day must lie within the 5-95 % band of the NetLogo runs on the same day. "Days inside the band" counts the days that pass. A metric passes (✅) only if all days pass. The last three columns show the values of one example day.

| | Metric | Days inside the band | Example day | gobees median | NetLogo median [5-95 %] |
| --- | --- | --- | ---: | ---: | --- |
| ✅ | foodsources | 1 of 1 days | 0 | 200 | 200 [200, 200] |
| ✅ | masterpatches | 1 of 1 days | 0 | 36 | 36 [36, 36] |
| ✅ | nestsites | 1 of 1 days | 0 | 23 | 23 [23, 23] |
| ✅ | nestsite_area | 1 of 1 days | 0 | 306300 | 306300 [306300, 306300] |
| ✅ | mean_queen_weight | 1 of 1 days | 0 | 785.204 | 785.257 [777.529, 792.012] |
| ✅ | mean_emerging_date | 1 of 1 days | 0 | 91.132 | 91.011 [89.2796, 93.1467] |
| ✅ | colonies_founded | 251 of 251 days | 250 | 135.5 | 134 [119, 151] |
| ✅ | queens_searching | 251 of 251 days | 250 | 0 | 0 [0, 0] |
| ✅ | queens_hibernating | 251 of 251 days | 250 | 0 | 0 [0, 0] |

## Step 4a: morning routine

Experiment `go/validation/experiments/step4a.xml`: example farm, 400 days. Day of the year, nectar and pollen available in the landscape. These values depend only on the landscape, so the NetLogo band is widened by a relative tolerance of 1e-9 (float sums in random order).

100 runs each. Every metric is checked on each of its days (all days of the experiment, unless the description above says otherwise): the median of the gobees runs on a day must lie within the 5-95 % band of the NetLogo runs on the same day. "Days inside the band" counts the days that pass. A metric passes (✅) only if all days pass. The last three columns show the values of one example day.

| | Metric | Days inside the band | Example day | gobees median | NetLogo median [5-95 %] |
| --- | --- | --- | ---: | ---: | --- |
| ✅ | day | 401 of 401 days | 400 | 35 | 35 [35, 35] |
| ✅ | nectar_l | 401 of 401 days | 400 | 0.0112685 | 0.0112685 [0.0112685, 0.0112685] |
| ✅ | pollen_kg | 401 of 401 days | 400 | 0.0155049 | 0.0155049 [0.0155049, 0.0155049] |

## Steps 4b to 6: colony dynamics in year 1

Experiment `go/validation/experiments/step6.xml`: example farm, 500 B. terrestris queens, 364 days, with brood, the daily activity queue and foraging. Both models run without QueensLeavingNestProc and UpdateSeasonalEventsProc (step 7). Compared at every step: the gobees median must lie in the NetLogo 5-95 % band. The table shows day 150, the middle of the colony season.

100 runs each. Every metric is checked on each of its days (all days of the experiment, unless the description above says otherwise): the median of the gobees runs on a day must lie within the 5-95 % band of the NetLogo runs on the same day. "Days inside the band" counts the days that pass. A metric passes (✅) only if all days pass. The last three columns show the values of one example day.

| | Metric | Days inside the band | Example day | gobees median | NetLogo median [5-95 %] |
| --- | --- | --- | ---: | ---: | --- |
| ✅ | colonies | 365 of 365 days | 150 | 33 | 32.5 [25.95, 41] |
| ✅ | dead_colonies | 365 of 365 days | 150 | 99 | 101 [87.85, 116.05] |
| ✅ | eggs | 365 of 365 days | 150 | 264 | 264 [96, 409.8] |
| ✅ | larvae | 365 of 365 days | 150 | 444 | 444 [251.4, 684] |
| ✅ | pupae | 365 of 365 days | 150 | 510 | 504 [288, 708.6] |
| ✅ | workers | 365 of 365 days | 150 | 414 | 426 [214.8, 720.6] |
| ✅ | males | 365 of 365 days | 150 | 168 | 180 [84, 301.2] |
| ✅ | young_queens | 365 of 365 days | 150 | 36 | 36 [0, 96.6] |
| ✅ | energy_kJ | 365 of 365 days | 150 | 11267.4 | 10997.5 [7181.8, 15971.6] |
| ✅ | pollen_g | 365 of 365 days | 150 | 188.651 | 188.026 [128.883, 295.564] |
| ✅ | trips | 365 of 365 days | 150 | 3172 | 3250.5 [1030.8, 6598.65] |
| ✅ | switch_point | 365 of 365 days | 150 | 19 | 18 [12, 24.05] |
| ✅ | queen_production | 365 of 365 days | 150 | 16 | 15 [9, 21.05] |
| ✅ | nectar_left_l | 365 of 365 days | 150 | 1.57155 | 1.57687 [1.42276, 1.65923] |
| ✅ | pollen_left_kg | 365 of 365 days | 150 | 0.580268 | 0.581087 [0.559733, 0.590312] |

## Step 7: the full model over 3 years

Experiment `go/validation/experiments/step7.xml`: example farm, 500 B. terrestris queens, 1095 days, the full model (NetLogo `Go`, gobees `model.Default`) with the year cycle: young queens leave the nest and hibernate, the season ends, the queens emerge again. Both run with StopExtinct? = false. Compared at every step: the gobees median must lie in the NetLogo 5-95 % band. The table shows day 515 (day 150 of year 2).

100 runs each. Every metric is checked on each of its days (all days of the experiment, unless the description above says otherwise): the median of the gobees runs on a day must lie within the 5-95 % band of the NetLogo runs on the same day. "Days inside the band" counts the days that pass. A metric passes (✅) only if all days pass. The last three columns show the values of one example day.

| | Metric | Days inside the band | Example day | gobees median | NetLogo median [5-95 %] |
| --- | --- | --- | ---: | ---: | --- |
| ✅ | colonies | 1096 of 1096 days | 515 | 25 | 26 [14, 40.15] |
| ✅ | dead_colonies | 1096 of 1096 days | 515 | 215 | 214.5 [169.55, 269.4] |
| ✅ | colonies_founded | 1096 of 1096 days | 515 | 241 | 238.5 [182.7, 306.35] |
| ✅ | eggs | 1096 of 1096 days | 515 | 210 | 228 [35.4, 444.6] |
| ✅ | larvae | 1096 of 1096 days | 515 | 360 | 336 [142.8, 710.4] |
| ✅ | pupae | 1096 of 1096 days | 515 | 432 | 450 [215.4, 805.2] |
| ✅ | workers | 1096 of 1096 days | 515 | 384 | 396 [130.8, 638.4] |
| ✅ | males | 1096 of 1096 days | 515 | 180 | 162 [72, 348.6] |
| ✅ | young_queens | 1096 of 1096 days | 515 | 0 | 0 [0, 24] |
| ✅ | hibernating_queens | 1096 of 1096 days | 515 | 84 | 96 [12, 168] |
| ✅ | energy_kJ | 1096 of 1096 days | 515 | 8946.62 | 9019.46 [4455.02, 14132.8] |
| ✅ | pollen_g | 1096 of 1096 days | 515 | 161.356 | 157.115 [78.6733, 245.487] |
| ✅ | trips | 1096 of 1096 days | 515 | 2622 | 2471 [541.4, 5926.95] |
| ✅ | nectar_left_l | 1096 of 1096 days | 515 | 1.6026 | 1.61199 [1.45076, 1.68875] |
| ✅ | pollen_left_kg | 1096 of 1096 days | 515 | 0.583593 | 0.584464 [0.560194, 0.593013] |

