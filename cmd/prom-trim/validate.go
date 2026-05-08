package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourorg/prom-trim/internal/rules"
)

// runValidate parses the --validate flag and validates the given rules file,
// printing a human-readable summary. Returns true if validation was handled.
func runValidate(args []string) bool {
	vfs := flag.NewFlagSet("validate", flag.ContinueOnError)
	var rulesFile string
	vfs.StringVar(&rulesFile, "rules", "", "Path to the Prometheus rules file to validate")

	if err := vfs.Parse(args); err != nil {
		return false
	}

	if rulesFile == "" {
		return false
	}

	fmt.Fprintf(os.Stdout, "Validating rules file: %s\n", rulesFile)

	if err := rules.ValidateFile(rulesFile); err != nil {
		fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
		os.Exit(2)
	}

	fmt.Fprintln(os.Stdout, "Validation passed: no issues found.")
	return true
}
