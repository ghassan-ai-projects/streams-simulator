package adapter

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

func stringify(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case bool:
		return strconv.FormatBool(x)
	default:
		return stringifyNumberOrObject(v)
	}
}

func writeJSONString(b *strings.Builder, s string) {
	raw, _ := json.Marshal(s)
	b.Write(raw)
}

func writeJSONValue(b *strings.Builder, v any) {
	switch x := v.(type) {
	case nil:
		b.WriteString("null")
	case string:
		writeJSONString(b, x)
	case bool:
		b.WriteString(strconv.FormatBool(x))
	default:
		writeStructuredJSON(b, v)
	}
}

type orderedObject []orderedField

type orderedField struct {
	name  string
	value any
}

func jsonEqualish(a, b any) bool {
	sa, err1 := canonical.MarshalString(a)
	sb, err2 := canonical.MarshalString(b)
	return err1 == nil && err2 == nil && sa == sb
}

func csvEscape(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

func stringifyNumberOrObject(v any) string {
	switch x := v.(type) {
	case float64:
		s, _ := canonical.MarshalString(x)
		return s
	case int64:
		return strconv.FormatInt(x, 10)
	case json.Number:
		return x.String()
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func writeStructuredJSON(b *strings.Builder, v any) {
	switch x := v.(type) {
	case float64:
		writeJSONFloat(b, x)
	case int64:
		b.WriteString(strconv.FormatInt(x, 10))
	case json.Number:
		b.WriteString(x.String())
	case orderedObject:
		writeOrderedObject(b, x)
	default:
		raw, _ := json.Marshal(v)
		b.Write(raw)
	}
}

func writeJSONFloat(b *strings.Builder, value float64) {
	text, err := canonical.MarshalString(value)
	if err != nil {
		text = "null"
	}
	b.WriteString(text)
}

func writeOrderedObject(b *strings.Builder, fields orderedObject) {
	b.WriteByte('{')
	for i, field := range fields {
		if i > 0 {
			b.WriteByte(',')
		}
		writeJSONString(b, field.name)
		b.WriteByte(':')
		writeJSONValue(b, field.value)
	}
	b.WriteByte('}')
}
