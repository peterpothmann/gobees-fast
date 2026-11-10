# Changes against NetLogo

gobees-fast started as an exact copy of gobees, the faithful port of BEE-STEWARD. This file lists where gobees-fast now differs from the NetLogo model, why, and what that does to the results. It describes the current state. The history of the steps is in the pull requests.

There are two kinds of changes:

| Kind | Effect on the results | How it is checked |
| --- | --- | --- |
| **Statistically equal** | other random numbers, the same distribution of the results | statistical comparison with the end of phase 1 (100 seeds per scenario), with gobees and with the NetLogo band, in every pull request ([go/compare/README.md](../go/compare/README.md)) |
| **Fix** | the results change on purpose | effect on the default run and position in the NetLogo band, recorded in the pull request, with the decision and its reason |

A run of gobees-fast with the same seed gives other numbers than NetLogo and gobees since the first statistically equal change.

## Statistically equal changes

### Random order

Changed in [#47](https://github.com/peterpothmann/gobees-fast/pull/47).

NetLogo asks agents (`ask bees`, `ask colonies`, ...) in random order, so every `ask` shuffles. gobees-fast shuffles only where the result depends on the order. Everywhere else it goes through the agents in storage order.

A shuffle is left out when each agent changes only itself (or adds to a sum) and either draws no random number or draws its own. The draws are independent and identically distributed, so which draw goes to which agent does not change the distribution. Sums in another order differ only in the last bits.

Left out:

| Part | Agents |
| --- | --- |
| brood incubation (`activities.incubateBrood`) | the brood of the colony: each gets the same heat |
| night rest | adult females |
| badgers (evening) | badgers, colonies in range, bees of a colony dug up: one draw per badger and colony, a colony dug up twice loses nothing more |
| brood mortality, development | all bees: own state, own draws |
| larvae feeding, larvae needs | colonies, larvae: own stores, sums |
| queens leaving the nest | young queens: each draws her male, males are not used up |
| queen production date | colonies |
| colony update | colonies, and the bees that die in a colony (all of them) |
| colony foundation | queens in nest construction (each founds her own colony), foodsources in range (the order only reaches the search, which stays random) |
| crop rotation | bees, colonies, species: each clears its lists |
| flower refill | foodsources |
| queen emergence | emerging cohorts, emerged queens, victims of a cuckoo bee, counts |
| nest site choice | sites: one uniform number, each site with probability area / total area in any order |
| season end, males, limit of hibernating queens | species and bees: all die, or uniform draws from the list |
| setup | badgers (identical until placed), habitat types of the nest habitats (only used with `member?`) |

Kept, because the order matters:

| Part | Why |
| --- | --- |
| wake-up | ties of the personal time: the order of the active list decides which bee empties a foodsource first |
| queen emergence, searching queens | a cuckoo bee can kill another one in the same host |
| merging of hedges | ties of the area: the first hedge of a field becomes the master |
| splitting of flower patches into layers | the creation order of the layers reaches several places |

Replaced by an equal draw: the search of a forager. NetLogo draws once per flower patch in random order and takes the last one detected, that is one of the detected patches, each with the same probability. gobees-fast draws once per patch in list order and lets the k-th detected patch replace the choice with probability 1/k (reservoir sampling).

### Order of the activity queue

Changed in [#PR](https://github.com/peterpothmann/gobees-fast/pull/PR).

During the day the bees act in the order of their personal time. After her action a bee goes back into the queue with her new time (NetLogo: `NextActiveBeeREP`). NetLogo puts her behind all bees with a time up to hers, with one exception that comes from the rounding of its binary search: if her time equals the time of the last bee of the queue, she goes before that bee. gobees-fast puts her behind all bees with the same time in every case.

Why: the exception has no meaning in the model. Keeping it needed the last bee of the queue at all times, a search for it on every tie, and a new search after the last bee died.

Effect: the order changes only among bees with the same time at the end of the queue. In the example farm (3 seeds × 10 runs × 730 days) the exception occurs in 22,395 of 16,359,703 re-insertions (0.14 %), about 750 times per run. The bee that acts first can, for example, empty a foodsource first, so single trips change. The statistical comparison of the pull request checks that the distribution of the results stays the same.

### Survival of a foraging trip

Changed in [#PR](https://github.com/peterpothmann/gobees-fast/pull/PR).

A forager survives a trip of t seconds with the probability s^t, with s = 1 − `MortalityForager_per_s` · `ForagingMortalityFactor` (NetLogo: `Foraging_costs&choiceProc`). gobees-fast computes exp(t · ln s), with ln s computed once per run, instead of the power.

Why: the power took about 3 % of the run time (profile of the example farm). The exponential with the stored logarithm skips the splitting into integer and fraction and the repeated squaring of the power.

Effect: both values agree up to the last bits. A trip ends with death if a uniform random number is below 1 − s^t, so a different last bit changes the outcome only if the random number falls between the two values. Counted on the example farm and three scenarios (5,887,293 trips): the last bits differ in 99.97 % of the trips, the outcome in none. The output of the 5 scenarios (example farm, step7_crop, step8_species, step7_badgers, step8_genetics, seed 3, 3 runs, 730 days) is byte for byte the same as with the power. The change is listed here because the equality is not guaranteed for every input.

### Incubation per colony

Changed in [#PR](https://github.com/peterpothmann/gobees-fast/pull/PR).

A nurse heats the brood of her colony for 48 minutes. Every egg, larva and pupa of the colony gets the same share of her heat, the heat divided by the brood of the morning counts (NetLogo: `BroodIncubationProc`). NetLogo adds the share to each brood agent at every incubation. gobees-fast adds the shares of the day to one sum per colony and adds that sum to each brood agent once, at the end of the day. An egg laid during the day gets only the shares of the incubations after its laying.

Why: the loop over all brood agents at every incubation took about 6 % of the run time (profile of the example farm). The received heat is read only in the evening, by the brood mortality and the development.

Effect: one sum per day rounds differently from one addition per incubation. Counted on the example farm and two scenarios (10 runs × 730 days each, 758,523 brood agents and days): the received heat differs in the last bits in 81 % to 83 % of the cases, by at most 7.7 · 10⁻¹⁵ of its value. The brood dies or develops when the received heat passes a threshold, so a different last bit changes the outcome only if the heat lies within that distance of a threshold. The output of the 5 scenarios (example farm, step7_crop, step8_species, step7_badgers, step8_genetics, seed 3, 3 runs, 730 days) and of 40 runs of the example farm (seed 1, 730 days) is byte for byte the same as before. The change is listed here because the equality is not guaranteed for every input.

## Fixes

Fixes of edge cases that change no result of the scenarios, e.g. an unknown `ForagingMortalityModel` is an error instead of a warning: the lines "gobees-fast" in [trans/netlogo-findings.md](../trans/netlogo-findings.md). Fixes that change the results follow in phase 3, each with its measured effect here.
