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

// NetLogoStep7 is the reference of experiments/step7.xml.
//
//go:embed reference/netlogo_step7.csv
var NetLogoStep7 string

// NetLogoStep7Badgers is the reference of experiments/step7_badgers.xml.
//
//go:embed reference/netlogo_step7_badgers.csv
var NetLogoStep7Badgers string

// NetLogoStep7Crop is the reference of experiments/step7_crop.xml.
//
//go:embed reference/netlogo_step7_crop.csv
var NetLogoStep7Crop string

// NetLogoStep8IBM is the reference of experiments/step8_ibm.xml.
//
//go:embed reference/netlogo_step8_ibm.csv
var NetLogoStep8IBM string

// NetLogoStep8Genetics is the reference of experiments/step8_genetics.xml.
//
//go:embed reference/netlogo_step8_genetics.csv
var NetLogoStep8Genetics string

// NetLogoStep8Species is the reference of experiments/step8_species.xml.
//
//go:embed reference/netlogo_step8_species.csv
var NetLogoStep8Species string

// NetLogoStep8Weather is the reference of experiments/step8_weather.xml.
//
//go:embed reference/netlogo_step8_weather.csv
var NetLogoStep8Weather string

// NetLogoStep8Settings is the reference of experiments/step8_settings.xml.
//
//go:embed reference/netlogo_step8_settings.csv
var NetLogoStep8Settings string

// ParameterFiles holds the parameter files of the experiments that change parameters, e.g.
// experiments/step8_ibm_Parameters.csv (written by scripts/parameter_files.py). NetLogo reads them through
// MyParametersFile, the comparison tests parse the same files.
//
//go:embed experiments/*_Parameters.csv
var ParameterFiles embed.FS
