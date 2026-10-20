package domain

import (
	"fmt"
	"slices"
)

func (f commandFit) scaledArguments(mapped MappedCommand) error {
	for _, name := range sortedKeys(mapped.Args) {
		arg := mapped.Args[name]
		if arg.From == "" {
			continue
		}
		if err := f.scaled(name, arg); err != nil {
			return err
		}
	}
	return nil
}

func (f commandFit) scaled(name string, arg MappedArg) error {
	kind, raw := objectOf(f.kind[arg.From]), objectOf(f.raw[name])
	switch {
	case kind == nil:
		return fmt.Errorf("%s/args/%s/from: the kind command has no argument %q", f.at, name, arg.From)
	case raw == nil:
		return fmt.Errorf("%s/args/%s: the raw command has no argument %q", f.at, name, name)
	}
	return f.boundsFit(name, kind, raw, scaleOf(arg.Scale))
}

func (f commandFit) boundsFit(name string, kind, raw map[string]any, scale float64) error {
	for _, bound := range []string{"minimum", "maximum"} {
		if value, declared := numberOf(kind[bound]); declared && !withinBounds(raw, value*scale) {
			return fmt.Errorf("%s/args/%s: the kind's %s %v times %v is outside the raw argument", f.at, name, bound, value, scale)
		}
	}
	return nil
}

func (f commandFit) fixedArguments(mapped MappedCommand) error {
	for _, name := range sortedKeys(mapped.Args) {
		if arg := mapped.Args[name]; arg.From == "" {
			if err := f.rawValue(fmt.Sprintf("%s/args/%s/value", f.at, name), name, arg.Value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f commandFit) table(mapped MappedCommand) error {
	if mapped.By == "" {
		return nil
	}
	words := sortedWords(objectOf(f.kind[mapped.By])["enum"])
	if len(words) == 0 || !slices.Equal(words, sortedKeys(mapped.Table)) {
		return fmt.Errorf("%s/table: the words %v are not the enum of the kind argument %q %v", f.at, sortedKeys(mapped.Table), mapped.By, words)
	}
	for _, word := range words {
		if err := f.row(word, mapped.Table[word]); err != nil {
			return err
		}
	}
	return nil
}

func (f commandFit) row(word string, row map[string]any) error {
	for _, name := range sortedKeys(row) {
		if err := f.rawValue(fmt.Sprintf("%s/table/%s/%s", f.at, word, name), name, row[name]); err != nil {
			return err
		}
	}
	return nil
}

func (f commandFit) lease(mapped MappedCommand) error {
	if mapped.Lease == nil {
		return nil
	}
	return f.rawValue(f.at+"/lease", mapped.Lease.Arg, float64(mapped.Lease.MS))
}

func (f commandFit) rawValue(at, name string, value any) error {
	raw := objectOf(f.raw[name])
	if raw == nil {
		return fmt.Errorf("%s: the raw command has no argument %q", at, name)
	}
	number, isNumber := rawNumber(value, f.rawEnums[name])
	if isNumber && !withinBounds(raw, number) {
		return fmt.Errorf("%s: %v is outside the raw argument %q", at, value, name)
	}
	return nil
}

func (f commandFit) covered(mapped MappedCommand, required []string) error {
	set := rawArgumentsSet(mapped)
	for _, name := range required {
		if !set[name] {
			return fmt.Errorf("%s: the raw argument %q is required and the board mapping does not set it", f.at, name)
		}
	}
	return nil
}

func rawArgumentsSet(mapped MappedCommand) map[string]bool {
	set := map[string]bool{}
	for _, name := range sortedKeys(mapped.Args) {
		set[name] = true
	}
	for _, word := range sortedKeys(mapped.Table) {
		for _, name := range sortedKeys(mapped.Table[word]) {
			set[name] = true
		}
	}
	if mapped.Lease != nil {
		set[mapped.Lease.Arg] = true
	}
	return set
}

func withinBounds(schema map[string]any, value float64) bool {
	low, hasLow := numberOf(schema["minimum"])
	high, hasHigh := numberOf(schema["maximum"])
	return (!hasLow || value >= low) && (!hasHigh || value <= high)
}

func rawNumber(value any, words []string) (float64, bool) {
	if word, isWord := value.(string); isWord {
		code := slices.Index(words, word)
		return float64(code), code >= 0
	}
	return numberOf(value)
}
