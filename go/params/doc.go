// Package params contains the model parameters as ECS resources.
//
// Naming: Go names without units (step F15c). The unit stands in the doc comment, together with the NetLogo name
// where it differs (table in docs/names.md). The keys of the run file are the Go names in lowerCamelCase.
// Numbers are float64 like in NetLogo, except whole numbers (counts, days, step F8).
//
// Everything in params is read-only during the run (step F14): the parameters of the run file (Parameters,
// Output), the content of the input files (Input), the values that package input derives from them (Derived,
// ForagingPeriod, Species) and the hard-coded constants (Constants). Params.Apply adds them as resources. The state of
// the run is in package globals.
package params
