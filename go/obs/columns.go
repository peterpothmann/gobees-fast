package obs

import (
	"fmt"
	"slices"
	"strings"
)

// extraColumns are the columns of DailyOutputExtended after the NetLogo metrics.
var extraColumns = []string{"ForagerDeathsToday", "TotalAdultQueensEverProduced", "MeanCompetitionPointDate"}

// timeColumn is the first column of every selection: the NetLogo ticks (1 on the first day).
const timeColumn = "ticks"

// AllColumns returns the names of all columns of DailyOutputExtended, in the order of its output.
func AllColumns() []string {
	h := (&DailyOutput{}).Header()
	h = append(h, NetLogoMetrics...)
	return append(h, extraColumns...)
}

// columnDocs describe the columns. TestColumnDocs checks that every column has a description and that none is
// left over.
var columnDocs = map[string]string{
	// the totals of DailyOutput (OutputDailyProc)
	"ticks":                  "day of the run (NetLogo ticks), 1 on the first day",
	"TotalQueens":            "queens of all stages, queen brood included",
	"TotalMatedQueens":       "mated queens",
	"TotalUnmatedQueens":     "unmated queens",
	"TotalHibernatingQueens": "bees that hibernate",
	"TotalColonies":          "living colonies",
	"TotalBeeAgents":         "bee agents, a cohort counts once",
	"TotalMales":             "males of all stages",
	"TotalActiveBees":        "bees in the activity queue after the activity loop",
	"MeanWorkerWeight_mg":    "mean weight of the adult workers (mean over agents), 0 without",
	"MeanQueenWeight_mg":     "mean weight of the adult queens (mean over agents), 0 without",
	"MeanAdultWeight_mg":     "mean weight of all adults (mean over agents), 0 without",
	"TotalMapArea_km2":       "area of the map, set once",
	"ColonyDensity_km2":      "living colonies per km2 of the map",

	// the NetLogo metrics
	"colonies":              "living colonies",
	"dead_colonies":         "dead colonies since the last 1 January, or since the start with output.keepDeadColonies",
	"colonies_founded":      "colonies founded since the start",
	"eggs":                  "eggs",
	"larvae":                "larvae",
	"pupae":                 "pupae",
	"workers":               "adult workers",
	"males":                 "adult males",
	"young_queens":          "adult queens that are not mated yet",
	"hibernating_queens":    "bees that hibernate",
	"energy_kJ":             "energy store of all colonies",
	"pollen_g":              "pollen store of all colonies",
	"trips":                 "foraging trips of the day",
	"nectar_left_l":         "nectar in all foodsources",
	"pollen_left_kg":        "pollen in all foodsources",
	"badgers":               "badgers",
	"foodsources":           "foodsources",
	"bee_agents":            "bee agents, a cohort counts once",
	"diploid_males":         "adult and brood males with 2 alleles",
	"colonies_B_terrestris": "living colonies of B_terrestris",
	"colonies_B_pascuorum":  "living colonies of B_pascuorum",
	"colonies_B_lapidarius": "living colonies of B_lapidarius",
	"colonies_B_hortorum":   "living colonies of B_hortorum",
	"colonies_B_pratorum":   "living colonies of B_pratorum",
	"colonies_B_hypnorum":   "living colonies of B_hypnorum",
	"psithyrus":             "bees of the parasitic species (Psithyrus)",
	"foraging_period_s":     "foraging period of the day",
	"queens_removed":        "hibernating queens removed since the start (maxHibernatingQueens)",

	// further values
	"ForagerDeathsToday":           "foragers that died on a foraging trip today",
	"TotalAdultQueensEverProduced": "queens that emerged from pupae since the start",
	"MeanCompetitionPointDate":     "mean competition point date (ticks) of the living colonies that have one, 0 if none",
}

// ColumnDoc returns the description of the column name, "" for an unknown name.
func ColumnDoc(name string) string {
	return columnDocs[name]
}

// SelectColumns returns the columns of the output for the requested names, in the order requested. nil or empty
// means all columns. The column ticks is always the first: it is added if missing and moved to the front. An unknown
// or repeated name is an error that lists the available columns.
func SelectColumns(requested []string) ([]string, error) {
	all := AllColumns()
	if len(requested) == 0 {
		return all, nil
	}
	selected := []string{timeColumn}
	for _, name := range requested {
		if !slices.Contains(all, name) {
			return nil, fmt.Errorf("unknown output column %q, the columns are: %s", name, strings.Join(all, ", "))
		}
		if name == timeColumn {
			continue
		}
		if slices.Contains(selected, name) {
			return nil, fmt.Errorf("output column %q is listed twice", name)
		}
		selected = append(selected, name)
	}
	return selected, nil
}
