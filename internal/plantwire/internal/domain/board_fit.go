package domain

import (
	"fmt"
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func objectOf(value any) map[string]any {
	object, _ := value.(map[string]any)
	return object
}

func wordsOf(value any) []string {
	var words []string
	for _, word := range listOf(value) {
		if text, ok := word.(string); ok {
			words = append(words, text)
		}
	}
	return words
}

func channelRange(value any) *model.Range {
	if bounds := rangeOf(value); bounds != nil {
		return &model.Range{Min: bounds.Min, Max: bounds.Max}
	}
	return nil
}

func readingFits(name string, metric map[string]any, mapped MappedMetric, raw *model.Channel, rawWords []string) error {
	if len(mapped.Table) > 0 {
		return tableReadingFits(name, metric, mapped, raw, rawWords)
	}
	declared, scale := channelRange(metric["eu_range"]), scaleOf(mapped.Scale)
	if declared == nil || raw.Range == nil || (declared.Min*scale == raw.Range.Min && declared.Max*scale == raw.Range.Max) {
		return nil
	}
	return fmt.Errorf("/metrics/%s/eu_range: %v to %v times %v is not the raw range %v to %v of %s", name, declared.Min, declared.Max, scale,
		raw.Range.Min, raw.Range.Max, mapped.Raw)
}

func tableReadingFits(name string, metric map[string]any, mapped MappedMetric, raw *model.Channel, rawWords []string) error {
	if words := slices.Sorted(maps.Values(mapped.Table)); !slices.Equal(slices.Compact(words), sortedWords(metric["enum_values"])) {
		return fmt.Errorf("/metrics/%s/table: the words %v are not the kind's enum_values %v", name, words, metric["enum_values"])
	}
	if raw.Range != nil && (raw.Range.Min != 0 || raw.Range.Max != float64(len(rawWords)-1)) {
		return fmt.Errorf("/metrics/%s/raw: the codes of %s run %v to %v, its raw words %v", name, mapped.Raw, raw.Range.Min, raw.Range.Max, rawWords)
	}
	return nil
}

func sortedWords(value any) []string {
	return slices.Compact(slices.Sorted(slices.Values(wordsOf(value))))
}

func commandFits(at string, schema map[string]any, mapped MappedCommand, rawSchema map[string]any, rawEnums map[string][]string) error {
	fits := commandFit{at: at, kind: objectOf(schema["properties"]), raw: objectOf(rawSchema["properties"]), rawEnums: rawEnums}
	for _, check := range []func(MappedCommand) error{fits.scaledArguments, fits.fixedArguments, fits.table, fits.lease} {
		if err := check(mapped); err != nil {
			return err
		}
	}
	return fits.covered(mapped, wordsOf(rawSchema["required"]))
}

type commandFit struct {
	at       string
	kind     map[string]any
	raw      map[string]any
	rawEnums map[string][]string
}
