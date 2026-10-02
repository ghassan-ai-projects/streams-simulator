package adapter

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"strings"
)

// renderTemplate renders one record template; returns "" when the when-guard
// excludes it.
func (e *Engine) renderTemplate(t *model.RecordTemplate, ev *model.SimEvent) (string, error) {
	// Resolve the context: entity id rewrite applies to the event.
	ctx := newContext(ev, e.meta)
	if ev != nil {
		id, err := e.rewriteID(ev.EntityID)
		if err != nil {
			return "", err
		}
		(*ctx)["entity_id"] = id
	}
	if t.When != nil {
		val := ctx.value(t.When.Field)
		switch t.When.Op {
		case "present":
			if val == nil {
				return "", nil
			}
		case "absent":
			if val != nil {
				return "", nil
			}
		case "eq":
			if !jsonEqualish(val, t.When.Value) {
				return "", nil
			}
		case "ne":
			if jsonEqualish(val, t.When.Value) {
				return "", nil
			}
		}
	}
	switch e.Adapter.Encoding {
	case "jsonl", "json-array":
		return e.renderJSONObject(t.Fields, ctx)
	case "csv":
		return e.renderCSV(t.Fields, ctx)
	case "line":
		return e.renderLine(t.Fields, ctx)
	}
	return "", fmt.Errorf("adapter: unsupported encoding %q", e.Adapter.Encoding)
}

func (e *Engine) renderJSONObject(fields []model.Field, ctx *recordContext) (string, error) {
	var b strings.Builder
	b.WriteByte('{')
	first := true
	for _, f := range fields {
		v, err := evalExpr(&f.From, ctx)
		if err != nil {
			return "", err
		}
		if v == nil && f.OmitWhenNull {
			continue
		}
		if !first {
			b.WriteByte(',')
		}
		first = false
		writeJSONString(&b, f.Name)
		b.WriteByte(':')
		writeJSONValue(&b, v)
	}
	b.WriteByte('}')
	return b.String(), nil
}

func (e *Engine) renderCSV(fields []model.Field, ctx *recordContext) (string, error) {
	var parts []string
	for _, f := range fields {
		v, err := evalExpr(&f.From, ctx)
		if err != nil {
			return "", err
		}
		parts = append(parts, csvEscape(stringify(v)))
	}
	return strings.Join(parts, ","), nil
}

func (e *Engine) renderLine(fields []model.Field, ctx *recordContext) (string, error) {
	var parts []string
	for _, f := range fields {
		v, err := evalExpr(&f.From, ctx)
		if err != nil {
			return "", err
		}
		parts = append(parts, stringify(v))
	}
	return strings.Join(parts, " "), nil
}
