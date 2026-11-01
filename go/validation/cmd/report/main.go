// Command report compares the recorded validation results of gobees-fast with NetLogo and with gobees and writes
// the validation report, its plots and a short comment for a pull request (validation.Compare). See
// go/validation/README.md.
//
// Usage (in go/):
//
//	go run ./validation/cmd/report -fast DIR [-gobees DIR] [-out validation] [-comment FILE] [-url URL]
//
// DIR is a folder written by the comparison tests with VALIDATION_DUMP_DIR set.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gobees/validation"
)

func main() {
	fast := flag.String("fast", "", "recorded results of gobees-fast (required)")
	gobees := flag.String("gobees", "", "recorded results of gobees, may be empty or missing")
	out := flag.String("out", "validation", "folder of validation-report.md and plots/")
	comment := flag.String("comment", "", "file for the pull request comment, none if empty")
	url := flag.String("url", "", "link to the full report in the comment")
	flag.Parse()
	if *fast == "" {
		flag.Usage()
		os.Exit(2)
	}
	gobeesDir := *gobees
	if gobeesDir != "" {
		if _, err := os.Stat(gobeesDir); err != nil {
			log.Printf("no gobees results (%v), compared with NetLogo only", err)
			gobeesDir = ""
		}
	}

	results, err := validation.Compare(*fast, gobeesDir)
	if err != nil {
		log.Fatal(err)
	}
	plots, err := validation.WritePlots(*out, results)
	if err != nil {
		log.Fatal(err)
	}
	report := filepath.Join(*out, "validation-report.md")
	if err := os.WriteFile(report, []byte(validation.Report(results, plots)), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Println("report written to", report)
	if *comment != "" {
		if err := os.WriteFile(*comment, []byte(validation.Comment(results, *url)), 0o644); err != nil {
			log.Fatal(err)
		}
	}
}
