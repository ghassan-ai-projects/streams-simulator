package domain

import (
	"errors"
	"maps"
	"slices"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	suggestedVersion = "0.1.0"
	maxSuggestedID   = 64
	maxTitleBytes    = 120
)

var ErrNoKindFiles = errors.New("component: no kind file to suggest a twin for")

func SuggestTwinChoices(lib *Library, files map[string]map[string]any) (map[string]any, error) {
	return lib.SuggestChoices(files)
}

func (l *Library) SuggestChoices(files map[string]map[string]any) (map[string]any, error) {
	names := slices.Sorted(maps.Keys(files))
	if len(names) == 0 {
		return nil, ErrNoKindFiles
	}
	kinds := make([]any, 0, len(names))
	for _, name := range names {
		file, issues := readKind(files[name])
		if len(issues) > 0 {
			return nil, violation(model.ContractKindFile, name, issues)
		}
		kinds = append(kinds, l.suggestKind(file))
	}
	return suggestedDocument(names, kinds), nil
}

func suggestedDocument(names []string, kinds []any) map[string]any {
	title := "Twin of " + strings.Join(names, ", ")
	return map[string]any{"schema_version": "0.1", "id": suggestedID(names), "version": suggestedVersion,
		"title": title[:min(len(title), maxTitleBytes)], "kinds": kinds}
}

func suggestedID(names []string) string {
	id := strings.ReplaceAll(strings.Join(names, "-"), "_", "-") + "-twin"
	if len(id) > maxSuggestedID {
		return "board-twin"
	}
	return id
}

func (l *Library) suggestKind(file *kindFile) map[string]any {
	choice := map[string]any{"kind": file.Name, "instance": strings.ReplaceAll(file.Name, "_", "-") + "-01", "metrics": l.suggestMetrics(file)}
	if commands := suggestCommands(file); len(commands) > 0 {
		choice["commands"] = commands
	}
	return choice
}

func (l *Library) suggestMetrics(file *kindFile) map[string]any {
	metrics := map[string]any{}
	for _, metric := range file.Metrics {
		metrics[metric.Name] = map[string]any{"instrument": l.instrumentChoice(metric)}
	}
	return metrics
}

func (l *Library) instrumentChoice(metric kindMetric) string {
	fitting := l.fittingInstruments(metric)
	if len(fitting) == 0 {
		return instrumentNone
	}
	index := slices.IndexFunc(fitting, func(candidate Summary) bool { return namedAfter(candidate.ID, metric.Name) })
	return reference(fitting[max(index, 0)].ID, fitting[max(index, 0)].Version)
}

func (l *Library) fittingInstruments(metric kindMetric) []Summary {
	return slices.DeleteFunc(l.InstrumentsForUnit(metric.Unit), func(candidate Summary) bool {
		return l.fit(l.instruments[reference(candidate.ID, candidate.Version)], metric) != nil
	})
}

func namedAfter(id, metric string) bool {
	_, name, _ := strings.Cut(id, "/")
	return strings.Contains(metric, strings.ReplaceAll(name, "-", "_"))
}

func suggestCommands(file *kindFile) map[string]any {
	writable := writableNames(file)
	claims := claimsPerMetric(file, writable)
	commands := map[string]any{}
	for _, command := range file.Commands {
		if !command.SafeStop {
			commands[command.Name] = suggestResponse(command, writable, claims)
		}
	}
	return commands
}

func writableNames(file *kindFile) map[string]bool {
	names := map[string]bool{}
	for _, metric := range file.Metrics {
		if metric.Writable && metric.EURange != nil && metric.DataType != "boolean" {
			names[metric.Name] = true
		}
	}
	return names
}

func settableArguments(command kindCommand, writable map[string]bool) []string {
	var args []string
	properties, _ := command.ArgsSchema["properties"].(map[string]any)
	for _, arg := range slices.Sorted(maps.Keys(properties)) {
		if writable[arg] && numericType(argumentType(command.ArgsSchema, arg)) {
			args = append(args, arg)
		}
	}
	return args
}

func claimsPerMetric(file *kindFile, writable map[string]bool) map[string]int {
	claims := map[string]int{}
	for _, command := range file.Commands {
		if command.SafeStop {
			continue
		}
		for _, arg := range settableArguments(command, writable) {
			claims[arg]++
		}
	}
	return claims
}

func suggestResponse(command kindCommand, writable map[string]bool, claims map[string]int) map[string]any {
	args := settableArguments(command, writable)
	if len(args) == 0 || slices.ContainsFunc(args, func(arg string) bool { return claims[arg] > 1 }) {
		return map[string]any{"response": responseNone}
	}
	sets := map[string]any{}
	for _, arg := range args {
		sets[arg] = arg
	}
	return map[string]any{"response": responseDirect, "sets": sets}
}
