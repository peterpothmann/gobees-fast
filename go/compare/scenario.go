package compare

import (
	"fmt"
	"strings"

	"gobees/internal/examplefiles"
	"gobees/validation"
)

// Scenario is a setting of the full model that every version runs: the example farm with the parameters of a
// NetLogo experiment in go/validation/experiments. Each scenario is compared with the reference commit, with gobees
// and with the NetLogo runs of its experiment.
type Scenario struct {
	Name        string // name of the NetLogo experiment, e.g. "step7_badgers"
	Title       string
	Description string
	Days        int
	// Parameters is the parameter file of the experiment in validation.ParameterFiles, empty for the example farm.
	Parameters string
	// Changes are further values of the parameter file, as they appear in the file (see ParameterFile).
	Changes map[string]string
}

// Scenarios are the scenarios of every comparison, in the order of the comment and the report.
var Scenarios = []Scenario{
	{Name: "step7", Title: "Example farm, 3 years", Days: 1095,
		Description: "500 B. terrestris queens on the example farm, the full model with the year cycle: young queens " +
			"leave the nest and hibernate, the season ends, the queens emerge again."},
	{Name: "step7_badgers", Title: "Badgers", Days: 1095, Parameters: "experiments/step7_badgers_Parameters.csv",
		Description: "As the example farm, with 2 badgers (N_Badgers = 2). They live on the two Scrub patches and dig " +
			"up each colony within 735 m with p = 0.019 per day."},
	{Name: "step7_crop", Title: "Crop rotation", Days: 1095,
		Changes: map[string]string{
			"CropRotationListAsString": `"""_SYSTEM_Example_Farm_Foodsources.txt Example_Farm_PlotTest_Foodsources.txt"""`,
		},
		Description: "As the example farm, with a crop rotation of two maps: on day 365 the landscape is replaced, " +
			"year 2 runs on `Example_Farm_PlotTest_Foodsources.txt` (36 flower patches), year 3 on the example farm " +
			"again. NetLogo sets CropRotationList after Setup, gobees-fast reads CropRotationListAsString, which gives " +
			"the same list [PAR-07]."},
	{Name: "step8_ibm", Title: "Individual-based colonies", Days: 730,
		Parameters:  "experiments/step8_ibm_Parameters.csv",
		Description: "COLONIES_IBM = 10: the first 10 colonies are individual-based [CCOL-05]."},
	{Name: "step8_genetics", Title: "Genetics", Days: 730, Parameters: "experiments/step8_genetics_Parameters.csv",
		Description: "SexLocus? = true (diploid males), UnlimitedMales? = false (a queen without a male dies)."},
	{Name: "step8_species", Title: "Several species", Days: 730, Parameters: "experiments/step8_species_Parameters.csv",
		Description: "All 7 species: 150 queens of each Bombus species and 100 Psithyrus queens."},
	{Name: "step8_weather", Title: "Weather and foraging mortality", Days: 730,
		Parameters:  "experiments/step8_weather_Parameters.csv",
		Description: `Weather = "foragingHoursExample" (foraging hours per day of the year), ForagingMortalityModel = "low".`},
	{Name: "step8_settings", Title: "Other settings", Days: 730, Parameters: "experiments/step8_settings_Parameters.csv",
		Description: "WinterMortality?, MergeHedges?, MinSizeFoodSources?, RemoveEmptyFoodSources?, KeepDeadColonies? " +
			"= false, MaxHibernatingQueens = 150 (the limit takes effect on every 1 January, also on the first day)."},
}

// ParameterFile returns the parameter file of the scenario: that of its experiment (or of the example farm) with
// its Changes and with StopExtinct? = false, as all experiments run [GO-20], so that every run has all days.
func (s Scenario) ParameterFile() (string, error) {
	text := examplefiles.ExampleFarmParameters
	if s.Parameters != "" {
		b, err := validation.ParameterFiles.ReadFile(s.Parameters)
		if err != nil {
			return "", err
		}
		text = string(b)
	}
	changes := map[string]string{"StopExtinct?": "false"}
	for name, value := range s.Changes {
		changes[name] = value
	}
	for name, value := range changes {
		var err error
		if text, err = setParameter(text, name, value); err != nil {
			return "", fmt.Errorf("%s: %w", s.Name, err)
		}
	}
	return text, nil
}

// setParameter sets the value of a parameter in the text of a parameter file: a header line and a line of values,
// separated by commas. The values contain no commas (scripts/parameter_files.py relies on the same).
func setParameter(text, name, value string) (string, error) {
	nl := "\n"
	if strings.Contains(text, "\r\n") {
		nl = "\r\n"
	}
	lines := strings.Split(text, nl)
	if len(lines) < 2 {
		return "", fmt.Errorf("a parameter file needs a header and a line of values")
	}
	header, values := strings.Split(lines[0], ","), strings.Split(lines[1], ",")
	if len(header) != len(values) {
		return "", fmt.Errorf("%d names and %d values", len(header), len(values))
	}
	for i, h := range header {
		if strings.TrimSpace(strings.Trim(h, `"`)) == name {
			values[i] = value
			lines[1] = strings.Join(values, ",")
			return strings.Join(lines, nl), nil
		}
	}
	return "", fmt.Errorf("no parameter %s", name)
}
