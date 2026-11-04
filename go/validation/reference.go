package validation

import "embed"

// The NetLogo reference data: quantiles of 100 NetLogo runs per step and metric, written by
// scripts/aggregate_netlogo.py from the BehaviorSpace runs of the experiments in experiments/ (see README.md).

// NetLogoStep3 is the reference of experiments/step3.xml.
//
//go:embed reference/netlogo_step3.csv
var NetLogoStep3 string

// NetLogoStep4a is the reference of experiments/step4a.xml.
//
//go:embed reference/netlogo_step4a.csv
var NetLogoStep4a string

// NetLogoStep6 is the reference of experiments/step6.xml.
//
//go:embed reference/netlogo_step6.csv
var NetLogoStep6 string

// ParameterFiles holds the parameter files of the experiments that change parameters, e.g.
// experiments/step8_genetics_Parameters.csv (written by scripts/parameter_files.py). NetLogo reads them through
// MyParametersFile, go/compare runs the scenarios with the same files.
//
//go:embed experiments/*_Parameters.csv
var ParameterFiles embed.FS

// referenceFS holds all NetLogo reference files: netlogo_<experiment>.csv (quantiles per step and metric) and
// netlogo_<experiment>_runs.csv (run summaries, see RunSummaries).
//
//go:embed reference
var referenceFS embed.FS

// NetLogoReference returns the NetLogo quantiles of an experiment and its run summaries (ParseRuns). runs is empty
// if the experiment has no run summaries.
func NetLogoReference(experiment string) (quantiles, runs string, err error) {
	q, err := referenceFS.ReadFile("reference/netlogo_" + experiment + ".csv")
	if err != nil {
		return "", "", err
	}
	r, err := referenceFS.ReadFile("reference/netlogo_" + experiment + "_runs.csv")
	if err != nil {
		return string(q), "", nil
	}
	return string(q), string(r), nil
}
