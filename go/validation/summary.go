package validation

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
)

// Summary identifies a summary of one metric of a NetLogo run: the value at step 0 (Period 0, Kind "value"), or the
// maximum or mean over the steps of a year (Period 1: steps 1 to 365, 2: 366 to 730, ...; Kind "max" or "mean"). A
// last year with fewer steps uses the steps it has. scripts/aggregate_netlogo.py writes them.
type Summary struct {
	Metric string
	Period int
	Kind   string
}

func (s Summary) String() string {
	if s.Period == 0 {
		return s.Metric + ", setup"
	}
	return fmt.Sprintf("%s, year %d, %s", s.Metric, s.Period, s.Kind)
}

// ParseRuns reads a table of run summaries (scripts/aggregate_netlogo.py): Summaries[summary][i], in the order of the
// rows.
func ParseRuns(table string) (map[Summary][]float64, error) {
	rows, err := csv.NewReader(strings.NewReader(table)).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 || len(rows[0]) != 5 || rows[0][0] != "run" {
		return nil, fmt.Errorf("expected the header run,metric,period,kind,value")
	}
	result := map[Summary][]float64{}
	for i, r := range rows[1:] {
		period, err := strconv.Atoi(r[2])
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i+2, err)
		}
		x, err := strconv.ParseFloat(r[4], 64)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i+2, err)
		}
		k := Summary{r[1], period, r[3]}
		result[k] = append(result[k], x)
	}
	return result, nil
}
