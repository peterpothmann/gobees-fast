// Command report compares the recorded validation results of gobees-fast with NetLogo and writes the validation
// report and its plots (validation.Compare). With -comment it also writes the pull request comment of the step: the
// checks of go/compare (-checks) and the check against NetLogo (compare.Comment). See go/validation/README.md.
//
// Usage (in go/):
//
//	go run ./validation/cmd/report -fast DIR [-out validation] [-checks DIR] [-comment FILE] [-url URL] [-run URL]
//
// -fast is a folder written by the comparison tests with VALIDATION_DUMP_DIR set, -checks one written by
// go/compare/run_compare.sh with COMPARE_CHECKS set.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gobees/compare"
	"gobees/validation"
)

func main() {
	fast := flag.String("fast", "", "recorded results of gobees-fast (required)")
	out := flag.String("out", "validation", "folder of validation-report.md and plots/")
	checksDir := flag.String("checks", "", "results of go/compare for the comment, may be empty or missing")
	comment := flag.String("comment", "", "file for the pull request comment, none if empty")
	url := flag.String("url", "", "link to the full report in the comment")
	run := flag.String("run", "", "link to the workflow run in the comment")
	flag.Parse()
	if *fast == "" {
		flag.Usage()
		os.Exit(2)
	}

	checks := []compare.Check{}
	if *checksDir != "" {
		var err error
		if checks, err = compare.ReadChecks(*checksDir); err != nil {
			log.Fatal(err)
		}
	}
	// Without validation results the comment is still written: it shows the NetLogo check as failed.
	failed := false
	if results, err := validation.Compare(*fast); err != nil {
		log.Print(err)
		failed = true
	} else {
		if err := writeReport(*out, results); err != nil {
			log.Fatal(err)
		}
		checks = append(checks, validation.NetLogoCheck(results, *url))
	}
	if *comment != "" {
		if err := os.WriteFile(*comment, []byte(compare.Comment(checks, *run)), 0o644); err != nil {
			log.Fatal(err)
		}
	}
	if failed {
		os.Exit(1)
	}
}

func writeReport(out string, results []validation.ExperimentResult) error {
	plots, err := validation.WritePlots(out, results)
	if err != nil {
		return err
	}
	report := filepath.Join(out, "validation-report.md")
	if err := os.WriteFile(report, []byte(validation.Report(results, plots)), 0o644); err != nil {
		return err
	}
	fmt.Println("report written to", report)
	return nil
}
