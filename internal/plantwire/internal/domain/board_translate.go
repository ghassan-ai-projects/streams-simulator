package domain

import (
	"errors"
	"fmt"
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
)

var ErrNotTranslated = errors.New("plantwire: the command does not translate through its board mapping")

func CheckArguments(schema map[string]any, args map[string]any) error {
	compiled, err := jsonschema.Compile(schema)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrNotTranslated, err)
	}
	if errs := compiled.Validate(map[string]any(args)); len(errs) > 0 {
		return fmt.Errorf("%w: %w", ErrNotTranslated, errs[0])
	}
	return nil
}

func RawCommand(mapped MappedKind, name string, args map[string]any) (string, map[string]any, error) {
	command, err := translatable(mapped, name, args)
	if err != nil {
		return "", nil, err
	}
	raw, err := rawArguments(command, args, mapped.RawEnums)
	if err != nil {
		return "", nil, err
	}
	if err := tableArguments(command, args, mapped.RawEnums, raw); err != nil {
		return "", nil, err
	}
	return command.Raw, raw, nil
}

func translatable(mapped MappedKind, name string, args map[string]any) (MappedCommand, error) {
	command, known := mapped.Commands[name]
	if !known {
		return MappedCommand{}, fmt.Errorf("%w: no command %q", ErrNotTranslated, name)
	}
	if extra := unconsumed(command, args); extra != "" {
		return MappedCommand{}, fmt.Errorf("%w: the board mapping does not use the argument %q", ErrNotTranslated, extra)
	}
	return command, nil
}

func unconsumed(command MappedCommand, args map[string]any) string {
	used := map[string]bool{command.By: command.By != ""}
	for _, name := range sortedKeys(command.Args) {
		used[command.Args[name].From] = command.Args[name].From != ""
	}
	for _, name := range sortedKeys(args) {
		if !used[name] {
			return name
		}
	}
	return ""
}

func rawArguments(command MappedCommand, args map[string]any, rawEnums map[string][]string) (map[string]any, error) {
	raw := map[string]any{}
	if command.Lease != nil {
		raw[command.Lease.Arg] = command.Lease.MS
	}
	for _, arg := range sortedKeys(command.Args) {
		value, err := rawArgument(command.Args[arg], args)
		if err != nil {
			return nil, err
		}
		raw[arg] = twinValue(value, rawEnums[arg])
	}
	return raw, nil
}

func rawArgument(arg MappedArg, args map[string]any) (any, error) {
	if arg.From == "" {
		return arg.Value, nil
	}
	value, isNumber := scalarOf(args[arg.From])
	if !isNumber {
		return nil, fmt.Errorf("%w: the argument %q is not a number", ErrNotTranslated, arg.From)
	}
	return exact(value * scaleOf(arg.Scale)), nil
}

func tableArguments(command MappedCommand, args map[string]any, rawEnums map[string][]string, raw map[string]any) error {
	if command.By == "" {
		return nil
	}
	word, _ := args[command.By].(string)
	row, known := command.Table[word]
	if !known {
		return fmt.Errorf("%w: %q is not a word of the argument %q", ErrNotTranslated, word, command.By)
	}
	for _, name := range sortedKeys(row) {
		raw[name] = twinValue(row[name], rawEnums[name])
	}
	return nil
}

func twinValue(value any, words []string) any {
	if word, isWord := value.(string); isWord && slices.Contains(words, word) {
		return int64(slices.Index(words, word))
	}
	if number, isNumber := scalarOf(value); isNumber {
		return exact(number)
	}
	return value
}

func exact(number float64) any {
	if number == math.Trunc(number) && math.Abs(number) < 1<<53 {
		return int64(number)
	}
	return number
}

func scalarOf(value any) (float64, bool) {
	switch number := value.(type) {
	case int64:
		return float64(number), true
	case int:
		return float64(number), true
	}
	return numberOf(value)
}

func SemanticReading(mapped MappedKind, rawChannel string, value any) (string, any, bool) {
	for _, name := range sortedKeys(mapped.Metrics) {
		if metric := mapped.Metrics[name]; metric.Raw == rawChannel {
			semantic, known := semanticValue(metric, value, mapped.RawEnums[rawChannel])
			return name, semantic, known
		}
	}
	return "", nil, false
}

func semanticValue(metric MappedMetric, value any, words []string) (any, bool) {
	if len(metric.Table) > 0 {
		semantic, known := metric.Table[rawWord(value, words)]
		return semantic, known
	}
	number, isNumber := scalarOf(value)
	if !isNumber || metric.Scale == "" {
		return value, true
	}
	return number / scaleOf(metric.Scale), true
}

func rawWord(value any, words []string) string {
	if word, isWord := value.(string); isWord {
		return word
	}
	number, isNumber := scalarOf(value)
	if code := int(math.Round(number)); isNumber && code >= 0 && code < len(words) {
		return words[code]
	}
	return ""
}
