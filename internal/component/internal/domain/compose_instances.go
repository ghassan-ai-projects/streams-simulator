package domain

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonvalue"
)

const indexHolder = "{n}"

var placeholder = regexp.MustCompile(`\{[^{}]*\}`)

func hasInstance(entry Entry, instance string) bool {
	return slices.Contains(entry.Names, instance) || entry.IDTemplate != "" && rendersID(entry, instance)
}

func rendersID(entry Entry, instance string) bool {
	match := templatePattern(entry.IDTemplate).FindStringSubmatch(instance)
	if match == nil {
		return false
	}
	index, numeric := renderedIndex(match)
	return numeric && index >= 1 && index <= countLimit(entry.Count) && renderID(entry.IDTemplate, index) == instance
}

func renderedIndex(match []string) (int, bool) {
	if len(match) < 2 {
		return 1, true
	}
	index, err := strconv.Atoi(match[1])
	return index, err == nil
}

func renderID(template string, index int) string {
	return placeholder.ReplaceAllStringFunc(template, func(holder string) string {
		if holder == indexHolder {
			return strconv.Itoa(index)
		}
		return "a"
	})
}

func countLimit(count map[string]any) int {
	if limit, bounded := jsonvalue.Number(count["max"]); bounded && limit > 0 {
		return int(limit)
	}
	initial, _ := jsonvalue.Number(count["default"])
	return int(initial)
}

func templatePattern(template string) *regexp.Regexp {
	literals := placeholder.Split(template, -1)
	holders := placeholder.FindAllString(template, -1)
	var pattern strings.Builder
	pattern.WriteString("^")
	for i, literal := range literals {
		pattern.WriteString(regexp.QuoteMeta(literal))
		if i < len(holders) {
			pattern.WriteString(holderPattern(holders[i]))
		}
	}
	pattern.WriteString("$")
	return regexp.MustCompile(pattern.String())
}

func holderPattern(holder string) string {
	if holder == indexHolder {
		return "([0-9]+)"
	}
	return "a"
}
