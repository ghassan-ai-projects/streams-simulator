package domain

import (
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"slices"

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

func notReproducible(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrKindNotReproducible, fmt.Sprintf(format, args...))
}

func readKind(doc map[string]any) (*kindFile, error) {
	if err := checkKindMembers(doc); err != nil {
		return nil, err
	}
	var file kindFile
	if err := decodeKind(doc, &file); err != nil {
		return nil, err
	}
	return &file, file.checkReproducible()
}

func decodeKind(doc map[string]any, file *kindFile) error {
	raw, err := json.Marshal(doc)
	if err == nil {
		err = model.DecodeBytes(raw, file)
	}
	if err != nil {
		return notReproducible("%v", err)
	}
	return nil
}

func checkKindMembers(doc map[string]any) error {
	if extra := extraMember(doc, kindMembers); extra != "" {
		return notReproducible("member %q is not part of the plant wire interface a twin reproduces", extra)
	}
	for _, section := range []struct {
		name            string
		allowed, needed []string
	}{{"metrics", metricMembers, metricRequired}, {"commands", commandMembers, commandMembers}} {
		if err := checkSection(doc, section.name, section.allowed, section.needed); err != nil {
			return err
		}
	}
	return nil
}

func checkSection(doc map[string]any, section string, allowed, needed []string) error {
	for i, item := range members(doc, section) {
		if err := checkItem(item, allowed, needed, pointer(section, indexPart(i))); err != nil {
			return err
		}
	}
	return nil
}

func checkItem(item map[string]any, allowed, needed []string, at string) error {
	if extra := extraMember(item, allowed); extra != "" {
		return notReproducible("%s/%s is not part of the interface a twin reproduces", at, extra)
	}
	for _, name := range needed {
		if _, present := item[name]; !present {
			return notReproducible("%s needs %q: the interface a twin reproduces always states it", at, name)
		}
	}
	return nil
}

func extraMember(item map[string]any, allowed []string) string {
	for _, key := range slices.Sorted(maps.Keys(item)) {
		if !slices.Contains(allowed, key) {
			return key
		}
	}
	return ""
}

func (f *kindFile) checkReproducible() error {
	if !kindNamePattern.MatchString(f.Name) {
		return notReproducible("kind name %q must match %s", f.Name, kindNamePattern)
	}
	if len(f.Metrics) == 0 {
		return notReproducible("kind %q has no metric: a twin needs a reading to observe", f.Name)
	}
	return f.checkParts()
}

func (f *kindFile) checkParts() error {
	for _, metric := range f.Metrics {
		if err := metric.checkReproducible(); err != nil {
			return err
		}
	}
	for _, command := range f.Commands {
		if err := command.checkReproducible(); err != nil {
			return err
		}
	}
	return nil
}

func (m kindMetric) checkReproducible() error {
	switch {
	case !metricNamePattern.MatchString(m.Name):
		return notReproducible("metric name %q must match %s", m.Name, metricNamePattern)
	case m.DataType == "enum" || m.DataType == "string":
		return notReproducible("metric %q is %s: a twin reproduces number, integer and boolean metrics", m.Name, m.DataType)
	case (m.DataType == "boolean") != (m.Unit == ""):
		return notReproducible("metric %q: a boolean metric has no unit and a number or integer metric has one", m.Name)
	case m.EURange != nil && m.EURange.Min >= m.EURange.Max:
		return notReproducible("metric %q: eu_range min %v must be below max %v", m.Name, m.EURange.Min, m.EURange.Max)
	}
	return nil
}

func (c kindCommand) checkReproducible() error {
	switch {
	case !commandNamePattern.MatchString(c.Name):
		return notReproducible("command name %q must match %s", c.Name, commandNamePattern)
	case !slices.Contains([]string{"low", "medium", "high"}, c.RiskClass):
		return notReproducible("command %q: risk_class %q is not low, medium or high, the classes a domain effector holds", c.Name, c.RiskClass)
	}
	return nil
}
