// Package compare checks a step of gobees-fast: it runs the scenarios (scenario.go) with the current code, the
// reference commit and the base commit (gobees) and compares the current code with each of them and with NetLogo.
//
// All versions run as command line programs (go/main.go) with the same input files, parameter file, seeds and days,
// and write their daily output (-f, or -o for the run files of the current code). The tables are compared on the columns both versions have, matched by name:
//
//   - Exact: every value of every common column must be equal (phase 1).
//   - Statistical: distributions over many seeds (phases 2 and 3, and always against gobees), see stats.go.
//   - NetLogo: the daily median of each metric within the band of the NetLogo runs, see netlogo.go.
//   - Speed: run times of the current code and of the base commit, reported only.
//
// The programs are built by run_compare.sh, the tests read their paths from the environment (reference_test.go)
// and write the pull request comment and the NetLogo report (report.go). See README.md.
package compare
