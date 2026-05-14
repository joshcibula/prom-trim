package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/yourorg/prom-trim/internal/rules"
)

func runSchedule(args []string) error {
	fs := flag.NewFlagSet("schedule", flag.ContinueOnError)
	schedPath := fs.String("schedule-file", "prom-trim-schedule.json", "Path to schedule file")
	graceDays := fs.Int("grace-days", 7, "Days to wait before pruning a stale rule")
	formatFlag := fs.String("format", "table", "Output format: table or json")
	rulesFile := fs.String("rules-file", "", "Rules file to scan for stale rules")
	configFile := fs.String("config", "config.yaml", "Config file path")

	if err := fs.Parse(args); err != nil {
		return err
	}

	_ = configFile // used in full integration; config loaded by caller

	sched, err := rules.LoadSchedule(*schedPath)
	if err != nil {
		return fmt.Errorf("load schedule: %w", err)
	}

	if *rulesFile != "" {
		groups, err := rules.ParseFile(*rulesFile)
		if err != nil {
			return fmt.Errorf("parse rules: %w", err)
		}
		names := rules.AllRecordNames(groups)
		now := time.Now().UTC()
		for _, name := range names {
			sched.AddEntry(rules.ScheduleEntry{
				RuleName:    name,
				ScheduledAt: now,
				PruneAfter:  now.Add(time.Duration(*graceDays) * 24 * time.Hour),
				Reason:      "stale candidate",
			})
		}
		if err := rules.SaveSchedule(*schedPath, sched); err != nil {
			return fmt.Errorf("save schedule: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Scheduled %d rule(s) in %s\n", len(names), *schedPath)
	}

	switch *formatFlag {
	case "json":
		return printScheduleJSON(sched)
	default:
		return printScheduleTable(sched)
	}
}

func printScheduleTable(s rules.Schedule) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "RULE\tSCHEDULED AT\tPRUNE AFTER\tREASON")
	for _, e := range s.Entries {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			e.RuleName,
			e.ScheduledAt.Format(time.RFC3339),
			e.PruneAfter.Format(time.RFC3339),
			e.Reason,
		)
	}
	return w.Flush()
}

func printScheduleJSON(s rules.Schedule) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}
