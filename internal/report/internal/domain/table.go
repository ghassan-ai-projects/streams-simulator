package domain

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type field struct{ name, value string }

func fieldTable(fields []field) []string {
	rows := make([][]string, 0, len(fields))
	for _, f := range fields {
		if f.value != "" {
			rows = append(rows, []string{f.name, f.value})
		}
	}
	if len(rows) == 0 {
		return nil
	}
	return table([]string{"Field", "Value"}, rows)
}

func table(headers []string, rows [][]string) []string {
	lines := []string{row(headers), "|" + strings.Repeat("---|", len(headers))}
	for _, cells := range rows {
		lines = append(lines, row(cells))
	}
	return lines
}

func row(cells []string) string {
	escaped := make([]string, len(cells))
	for i, cell := range cells {
		escaped[i] = escapeCell(cell)
	}
	return "| " + strings.Join(escaped, " | ") + " |"
}

func countTable(header string, counts map[string]int64) []string {
	keys := sortedKeys(counts)
	rows := make([][]string, len(keys))
	for i, key := range keys {
		rows[i] = []string{key, strconv.FormatInt(counts[key], 10)}
	}
	return table([]string{header, "Rows"}, rows)
}

func sortedKeys[V any](values map[string]V) []string {
	return slices.Sorted(maps.Keys(values))
}

func code(value string) string {
	if value == "" {
		return ""
	}
	return spanMark + value + spanMark
}

func codeJSON(value any) string {
	raw, err := json.Marshal(value)
	if err != nil || string(raw) == "null" {
		return ""
	}
	return code(string(raw))
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func number(value *int64) string {
	if value == nil {
		return ""
	}
	return strconv.FormatInt(*value, 10)
}

func instant(ns, start int64) string {
	if ns == 0 {
		return ""
	}
	text := model.FormatTime(ns)
	if start == 0 || ns < start {
		return text
	}
	return fmt.Sprintf("%s (+%ss)", text, strconv.FormatFloat(float64(ns-start)/1e9, 'f', -1, 64))
}

func optionalInstant(ns *int64, start int64) string {
	if ns == nil {
		return ""
	}
	return instant(*ns, start)
}

func reference(ref *model.ArtifactRef) string {
	if ref == nil {
		return ""
	}
	return ref.ID + " " + ref.Version + " " + code(ref.Digest)
}
