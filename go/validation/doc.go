// Package validation holds the results of NetLogo reference runs and helpers to read them.
//
// The folder holds the NetLogo side of the comparison: the BehaviorSpace experiments (experiments/), the scripts
// that run NetLogo and aggregate its output (scripts/), the NetLogo reference data (reference/, embedded in
// reference.go) and this package. go/compare runs the scenarios of the full model against them and writes
// validation-report.md, the tests in go/sys (validation_step*_test.go) compare parts of the model. See README.md.
package validation
