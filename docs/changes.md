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

## Fixes

Fixes of edge cases that change no result of the scenarios, e.g. an unknown `ForagingMortalityModel` is an error instead of a warning: the lines "gobees-fast" in [trans/netlogo-findings.md](../trans/netlogo-findings.md). Fixes that change the results follow in phase 3, each with its measured effect here.
