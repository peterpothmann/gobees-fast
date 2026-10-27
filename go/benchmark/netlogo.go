package benchmark

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"
)

// NetLogo is the result of the NetLogo experiment netlogo.xml.
type NetLogo struct {
	Runs []Run // sorted by run number
	Days int   // simulated days per run
}

// setupPrefix starts the line that the setup commands of netlogo.xml print for each run, followed by the run
// number and the setup time in seconds, e.g. "gobees-benchmark setup 2 1.058".
const setupPrefix = "gobees-benchmark setup "

// ParseNetLogo reads the BehaviorSpace table (--table) and the console output of the experiment netlogo.xml. The
// table has one row per run with the metrics at the end of the run: the time of the days and the colonies founded.
// The setup times come from the console output.
func ParseNetLogo(table, console io.Reader) (NetLogo, error) {
	r := csv.NewReader(table)
	r.FieldsPerRecord = -1 // the lines before the header have fewer fields
	records, err := r.ReadAll()
	if err != nil {
		return NetLogo{}, err
	}
	header := slices.IndexFunc(records, func(rec []string) bool { return len(rec) > 0 && rec[0] == "[run number]" })
	if header < 0 {
		return NetLogo{}, fmt.Errorf("no header line with [run number] in the table")
	}
	column := map[string]int{}
	for i, name := range records[header] {
		column[name] = i
	}
	for _, name := range []string{"RAND_SEED", "[step]", "timer", "TotalColoniesEverProduced"} {
		if _, ok := column[name]; !ok {
			return NetLogo{}, fmt.Errorf("no column %s in the table", name)
		}
	}

	var result NetLogo
	for _, rec := range records[header+1:] {
		values := map[string]float64{}
		for _, name := range []string{"[run number]", "RAND_SEED", "[step]", "timer", "TotalColoniesEverProduced"} {
			v, err := strconv.ParseFloat(rec[column[name]], 64)
			if err != nil {
				return NetLogo{}, fmt.Errorf("column %s: %w", name, err)
			}
			values[name] = v
		}
		days := int(values["[step]"])
		if result.Days != 0 && days != result.Days {
			return NetLogo{}, fmt.Errorf("run %v has %d days, an earlier run %d", values["[run number]"], days, result.Days)
		}
		result.Days = days
		result.Runs = append(result.Runs, Run{
			Number:          int(values["[run number]"]),
			Seed:            values["RAND_SEED"],
			Days:            seconds(values["timer"]),
			ColoniesFounded: values["TotalColoniesEverProduced"],
		})
	}
	if len(result.Runs) == 0 {
		return NetLogo{}, fmt.Errorf("no runs in the table")
	}
	slices.SortFunc(result.Runs, func(a, b Run) int { return a.Number - b.Number })

	setups := map[int]time.Duration{}
	scanner := bufio.NewScanner(console)
	for scanner.Scan() {
		line, ok := strings.CutPrefix(strings.TrimSpace(scanner.Text()), setupPrefix)
		if !ok {
			continue
		}
		var number int
		var s float64
		if _, err := fmt.Sscanf(line, "%d %g", &number, &s); err != nil {
			return NetLogo{}, fmt.Errorf("setup line %q: %w", line, err)
		}
		setups[number] = seconds(s)
	}
	if err := scanner.Err(); err != nil {
		return NetLogo{}, err
	}
	for i, run := range result.Runs {
		s, ok := setups[run.Number]
		if !ok {
			return NetLogo{}, fmt.Errorf("no setup time of run %d in the console output", run.Number)
		}
		result.Runs[i].Setup = s
	}
	return result, nil
}

// seconds converts NetLogo's timer (seconds, millisecond resolution) to a duration.
func seconds(s float64) time.Duration {
	return time.Duration(s * float64(time.Second))
}
