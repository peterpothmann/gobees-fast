// Command netlogo2gobees converts the NetLogo input files of BEE-STEWARD into the run files (YAML) of gobees-fast.
//
// gobees-fast reads only run files. The NetLogo readers live here, with the special cases of the NetLogo
// files (findings in trans/netlogo-findings.md). The values do not change, only the format: gobees-fast derives the
// same values from a run file that it derived from the NetLogo files before.
//
//	go run . convert -d ../../netlogo -p my_Parameters.csv -o out -n mine   # out/run.yaml (name: mine) and its landscape files
//	go run . generate                                                        # the run files in the repository
package main
