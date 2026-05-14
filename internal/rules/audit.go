package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// AuditEvent records a single pruning or scheduling action.
type AuditEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	RuleName  string    `json:"rule_name"`
	RuleFile  string    `json:"rule_file"`
	Reason    string    `json:"reason"`
	DryRun    bool      `json:"dry_run"`
}

// AuditLog is a collection of audit events.
type AuditLog struct {
	Events []AuditEvent `json:"events"`
}

// AppendAudit appends one or more events to the audit log file.
func AppendAudit(path string, events []AuditEvent) error {
	log, err := LoadAudit(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("load audit log: %w", err)
	}
	log.Events = append(log.Events, events...)

	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal audit log: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// LoadAudit reads the audit log from disk.
func LoadAudit(path string) (AuditLog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AuditLog{}, err
	}
	var log AuditLog
	if err := json.Unmarshal(data, &log); err != nil {
		return AuditLog{}, fmt.Errorf("unmarshal audit log: %w", err)
	}
	return log, nil
}
