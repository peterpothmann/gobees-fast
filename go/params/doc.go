// Package params contains the model parameters as ECS resources.
//
// Naming: fields keep the NetLogo names, so they can be found in the NetLogo code.
// Characters that are invalid in Go are replaced: "?" is dropped, "/" becomes "_per_".
// Numbers are float64 like in NetLogo, except whole numbers (counts, days, step F8).
//
// Everything in params is read-only during the run (step F14): the parameters of the run file (ParameterFile,
// Output), the content of the input files (Input), the values that package input derives from them (Derived,
// Weather, Species) and the hard-coded constants (Constants). Params.Apply adds them as resources. The state of
// the run is in package globals.
package params
