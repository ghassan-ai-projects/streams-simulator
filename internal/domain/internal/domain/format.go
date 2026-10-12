package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const formatWidth = 110

var documentKeys = []string{"id", "version", "title", "description", "axes", "stresses", "kinds", "instances", "link_types", "links", "boards", "profiles", "ground_truth"}

var memberKeys = []string{"id", "name", "kind", "system", "names", "id_template", "count", "churn", "target", "tier", "form", "state", "unit", "initial", "default", "min", "max", "dynamics", "channels", "faults", "effectors"}

func Format(c *Compiled) ([]byte, error) {
	var doc any
	if err := model.DecodeBytes(c.Raw, &doc); err != nil {
		return nil, fmt.Errorf("domain: format %s: %w", c.Spec.ID, err)
	}
	return []byte(formatValue(doc, 0) + "\n"), nil
}

func formatValue(value any, indent int) string {
	compact := compactValue(value)
	if indent > 0 && 2*indent+len(compact) <= formatWidth {
		return compact
	}
	switch typed := value.(type) {
	case map[string]any:
		return formatObject(typed, indent)
	case []any:
		return formatArray(typed, indent)
	}
	return compact
}

func formatObject(object map[string]any, indent int) string {
	order := memberKeys
	if indent == 0 {
		order = documentKeys
	}
	if len(object) == 0 {
		return "{}"
	}
	pad := strings.Repeat("  ", indent+1)
	members := make([]string, 0, len(object))
	for _, key := range orderedKeys(object, order) {
		members = append(members, pad+quoted(key)+": "+formatValue(object[key], indent+1))
	}
	return "{\n" + strings.Join(members, ",\n") + "\n" + strings.Repeat("  ", indent) + "}"
}

func formatArray(array []any, indent int) string {
	if len(array) == 0 {
		return "[]"
	}
	pad := strings.Repeat("  ", indent+1)
	items := make([]string, 0, len(array))
	for _, item := range array {
		items = append(items, pad+formatValue(item, indent+1))
	}
	return "[\n" + strings.Join(items, ",\n") + "\n" + strings.Repeat("  ", indent) + "]"
}

func compactValue(value any) string {
	switch typed := value.(type) {
	case map[string]any:
		return compactObject(typed)
	case []any:
		return compactArray(typed)
	case nil:
		return "null"
	case string:
		return quoted(typed)
	case json.Number:
		return typed.String()
	}
	return fmt.Sprint(value)
}

func compactObject(object map[string]any) string {
	members := make([]string, 0, len(object))
	for _, key := range orderedKeys(object, memberKeys) {
		members = append(members, quoted(key)+": "+compactValue(object[key]))
	}
	return "{" + strings.Join(members, ", ") + "}"
}

func compactArray(array []any) string {
	items := make([]string, 0, len(array))
	for _, item := range array {
		items = append(items, compactValue(item))
	}
	return "[" + strings.Join(items, ", ") + "]"
}

func orderedKeys(object map[string]any, order []string) []string {
	keys := make([]string, 0, len(object))
	// determinism-safe: the keys are sorted below.
	for key := range object {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(a, b string) int {
		if ra, rb := keyRank(a, order), keyRank(b, order); ra != rb {
			return ra - rb
		}
		return strings.Compare(a, b)
	})
	return keys
}

func keyRank(key string, order []string) int {
	if i := slices.Index(order, key); i >= 0 {
		return i
	}
	if key == "description" {
		return len(order) + 1
	}
	return len(order)
}

func quoted(text string) string {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(text)
	return strings.TrimSuffix(out.String(), "\n")
}
