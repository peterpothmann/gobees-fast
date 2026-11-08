// Command gobees runs gobees-fast, the optimized Go version of the NetLogo model BEE-STEWARD.
//
// Without flags it runs the example farm (the run file built into package input). -c names another run file,
// -s the seed, --days the number of days, -o DIR writes the output into DIR/<name of the run file>: the daily
// output (package obs), the provenance and the copies of the landscape files. See README.md.
package main
