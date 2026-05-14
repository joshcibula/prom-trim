package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/yourorg/prom-trim/internal/rules"
)

func runGroupStats(args []string) error {
	fs := flag.NewFlagSet("groupstats", flag.ContinueOnError)
	format := fs.String("format", "table", "Output format: table or json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: prom-trim groupstats [flags] <rules-file>")
	}
	file := fs.Arg(0)

	// Load usage from snapshot if available; otherwise use empty map.
	usage := map[string]float64{}
	if snap, err := rules.LoadSnapshot(file + ".snapshot.json"); err == nil {
		for _, e := range snap {
			usage[e.Record] = float64(e.Count)
		}
	}

	stats, err := rules.BuildGroupStats(file, usage)
	if err != nil {
		return err
	}

	switch *format {
	case "json":
		return printGroupStatsJSON(stats)
	default:
		printGroupStatsTable(stats)
		return nil
	}
}

func printGroupStatsTable(stats []rules.GroupStat) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "GROUP\tRULES\tSTALE\tCOVERAGE")
	for _, s := range stats {
		fmt.Fprintf(w, "%s\t%d\t%d\t%.1f%%\n",
			s.GroupName, s.RuleCount, s.StaleCount, s.Coverage)
	}
	w.Flush()
}

func printGroupStatsJSON(stats []rules.GroupStat) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(stats)
}
