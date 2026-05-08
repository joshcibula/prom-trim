package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/prom-trim/internal/config"
	"github.com/prom-trim/internal/prometheus"
	"github.com/prom-trim/internal/report"
	"github.com/prom-trim/internal/rules"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config file")
	dryRun := flag.Bool("dry-run", false, "preview changes without writing")
	outputFmt := flag.String("output", "table", "output format: table or json")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	groups, err := rules.ParseFile(cfg.RulesFile)
	if err != nil {
		log.Fatalf("failed to parse rules file: %v", err)
	}

	ruleNames := rules.AllRecordNames(groups)
	if len(ruleNames) == 0 {
		fmt.Println("no recording rules found — nothing to do")
		os.Exit(0)
	}

	client, err := prometheus.NewClient(cfg)
	if err != nil {
		log.Fatalf("failed to create prometheus client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSeconds)*time.Second)
	defer cancel()

	usage, err := prometheus.FetchRuleUsage(ctx, client, ruleNames, cfg.LookbackDays)
	if err != nil {
		log.Fatalf("failed to fetch rule usage: %v", err)
	}

	rows := report.FromUsage(ruleNames, usage)
	summary := report.BuildSummary(rows)

	switch *outputFmt {
	case "json":
		out, fmtErr := report.FormatJSON(rows, *dryRun)
		if fmtErr != nil {
			log.Fatalf("failed to format json: %v", fmtErr)
		}
		fmt.Println(out)
	default:
		fmt.Print(report.FormatTable(rows, *dryRun))
	}

	report.Write(os.Stdout, rows, summary, *dryRun)

	if !*dryRun {
		if err := rules.Prune(cfg.RulesFile, groups, usage); err != nil {
			log.Fatalf("failed to prune rules: %v", err)
		}
		fmt.Printf("pruned %d stale rule(s) from %s\n", summary.StaleCount, cfg.RulesFile)
	} else {
		fmt.Printf("dry-run: %d stale rule(s) would be removed\n", summary.StaleCount)
	}
}
