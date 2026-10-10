package domain

import (
	"fmt"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// renderTemplate renders one record template; returns "" when the when-guard
// excludes it.
func (e *Engine) renderTemplate(t *model.RecordTemplate, ev *model.SimEvent) (string, error) {
	ctx, err := e.renderContext(ev)
	if err != nil {
		return "", err
	}
	if t.When != nil && !templateIncludes(t.When, ctx) {
		return "", nil
	}
	return e.renderEncoding(t.Fields, ctx)
}

func (e *Engine) renderJSONObject(fields []model.Field, ctx *recordContext) (string, error) {
	var b strings.Builder
	b.WriteByte('{')
	first := true
	for _, field := range fields {
		if err := renderJSONField(&b, &first, field, ctx); err != nil {
			return "", err
		}
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

func (e *Engine) renderContext(ev *model.SimEvent) (*recordContext, error) {
	ctx := newContext(ev, e.meta)
	if ev != nil {
		id, err := e.rewriteID(ev.EntityID)
		if err != nil {
			return nil, err
		}
		(*ctx)["entity_id"] = id
	}
	return ctx, nil
}

func templateIncludes(guard *model.WhenClause, ctx *recordContext) bool {
	value := ctx.value(guard.Field)
	switch guard.Op {
	case "present":
		return value != nil
	case "absent":
		return value == nil
	case "eq":
		return jsonEqualish(value, guard.Value)
	case "ne":
		return !jsonEqualish(value, guard.Value)
	}
	return true
}

func (e *Engine) renderEncoding(fields []model.Field, ctx *recordContext) (string, error) {
	switch e.Adapter.Encoding {
	case "jsonl", "json-array":
		return e.renderJSONObject(fields, ctx)
	case "csv":
		return e.renderCSV(fields, ctx)
	case "line":
		return e.renderLine(fields, ctx)
	}
	return "", fmt.Errorf("adapter: unsupported encoding %q", e.Adapter.Encoding)
}

func renderJSONField(b *strings.Builder, first *bool, field model.Field, ctx *recordContext) error {
	value, err := evalExpr(&field.From, ctx)
	if err != nil {
		return err
	}
	if value == nil && field.OmitWhenNull {
		return nil
	}
	writeRenderedField(b, first, field.Name, value)
	return nil
}

func writeRenderedField(b *strings.Builder, first *bool, name string, value any) {
	if !*first {
		b.WriteByte(',')
	}
	*first = false
	writeJSONString(b, name)
	b.WriteByte(':')
	writeJSONValue(b, value)
}
