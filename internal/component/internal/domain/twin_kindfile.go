package domain

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var (
	kindNamePattern    = regexp.MustCompile(`^[a-z][a-z0-9_]{0,19}$`)
	metricNamePattern  = regexp.MustCompile(`^[a-z][a-z0-9_]{0,23}$`)
	commandNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

	kindMembers    = []string{"kind", "metrics", "commands"}
	metricMembers  = []string{"name", "data_type", "unit", "eu_range", "resolution", "writable"}
	metricRequired = []string{"name", "data_type", "resolution", "writable"}
	commandMembers = []string{"name", "args_schema", "risk_class", "safe_stop"}
)

type kindMetric struct {
	Name       string  `json:"name"`
	DataType   string  `json:"data_type"`
	Unit       string  `json:"unit"`
	EURange    *Range  `json:"eu_range"`
	Resolution float64 `json:"resolution"`
	Writable   bool    `json:"writable"`
}

type kindCommand struct {
	Name       string         `json:"name"`
	ArgsSchema map[string]any `json:"args_schema"`
	RiskClass  string         `json:"risk_class"`
	SafeStop   bool           `json:"safe_stop"`
}

type kindFile struct {
	Name     string        `json:"kind"`
	Metrics  []kindMetric  `json:"metrics"`
	Commands []kindCommand `json:"commands"`
}

func readKind(doc map[string]any) (*kindFile, []model.Issue) {
	if issues := TwinRequirements(doc); len(issues) > 0 {
		return nil, issues
	}
	var file kindFile
	if err := decodeKind(doc, &file); err != nil {
		return nil, []model.Issue{model.RuleIssue("", KeywordDecode, "", err.Error())}
	}
	return &file, nil
}

func decodeKind(doc map[string]any, file *kindFile) error {
	raw, err := json.Marshal(doc)
	if err == nil {
		err = model.DecodeBytes(raw, file)
	}
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}
