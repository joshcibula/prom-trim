package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/prom-trim/internal/rules"
)

func runAudit(logPath, format string) error {
	log, err := rules.LoadAudit(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No audit log found.")
			return nil
		}
		return fmt.Errorf("load audit log: %w", err)
	}

	if len(log.Events) == 0 {
		fmt.Println("Audit log is empty.")
		return nil
	}

	if format == "json" {
		return printAuditJSON(log.Events)
	}
	return printAuditTable(log.Events)
}

func printAuditTable(events []rules.AuditEvent) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tACTION\tRULE\tFILE\tREASON\tDRY_RUN")
	for _, e := range events {
		dryRun := "-"
		if e.DryRun {
			dryRun = "yes"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			e.Timestamp.Format(time.RFC3339),
			e.Action,
			e.RuleName,
			e.RuleFile,
			e.Reason,
			dryRun,
		)
	}
	return w.Flush()
}

func printAuditJSON(events []rules.AuditEvent) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(events)
}
