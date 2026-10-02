package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
)

func stringify(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case bool:
		return strconv.FormatBool(x)
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
		if x {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case float64:
		s, err := canonical.MarshalString(x)
		if err != nil {
			b.WriteString("null")
			return
		}
		b.WriteString(s)
	case int64:
		b.WriteString(strconv.FormatInt(x, 10))
	case json.Number:
		b.WriteString(x.String())
	case orderedObject:
		b.WriteByte('{')
		for i, field := range x {
			if i > 0 {
				b.WriteByte(',')
			}
			writeJSONString(b, field.name)
			b.WriteByte(':')
			writeJSONValue(b, field.value)
		}
		b.WriteByte('}')
	default:
		raw, _ := json.Marshal(v)
		b.Write(raw)
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

func formatErrs(errs []jsonschema.Error) string {
	var b strings.Builder
	for i, e := range errs {
		if i == 10 {
			fmt.Fprintf(&b, "  ... and %d more\n", len(errs)-10)
			break
		}
		fmt.Fprintf(&b, "  %s\n", e.Error())
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func mustAny(b []byte) any {
	var v any
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		panic(err)
	}
	return v
}
