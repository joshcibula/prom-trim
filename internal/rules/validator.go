package rules

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidationError holds a list of issues found during rule file validation.
type ValidationError struct {
	Issues []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed with %d issue(s): %s", len(e.Issues), strings.Join(e.Issues, "; "))
}

// ValidateFile parses and validates a Prometheus rules YAML file,
// returning a ValidationError if any structural problems are found.
func ValidateFile(path string) error {
	groups, err := ParseFile(path)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	if len(groups) == 0 {
		return errors.New("rules file contains no groups")
	}

	var issues []string

	for _, g := range groups {
		if strings.TrimSpace(g.Name) == "" {
			issues = append(issues, "group has empty name")
		}
		for i, r := range g.Rules {
			name := ruleRecordName(r)
			if name == "" {
				issues = append(issues, fmt.Sprintf("group %q rule[%d] has no record name", g.Name, i))
			}
			var exprNode yaml.Node
			if err := r.Expr.Decode(&exprNode); err != nil || strings.TrimSpace(exprNode.Value) == "" {
				issues = append(issues, fmt.Sprintf("group %q rule %q has empty expr", g.Name, name))
			}
		}
	}

	if len(issues) > 0 {
		return &ValidationError{Issues: issues}
	}
	return nil
}
