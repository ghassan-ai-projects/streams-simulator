package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// recordContext exposes native event fields to the transforms.
type recordContext map[string]any

func newContext(ev *model.SimEvent, meta map[string]any) *recordContext {
	ctx := recordContext{}
	// determinism-safe: copies a map into a map.
	for k, v := range meta {
		ctx[k] = v
	}
	if ev != nil {
		ctx.addEvent(ev)
	}
	return &ctx
}

func (c *recordContext) value(field string) any {
	if c == nil {
		return nil
	}
	return (*c)[field]
}

// evalExpr evaluates one transform.
func evalExpr(e *model.ValueExpr, ctx *recordContext) (any, error) {
	switch e.Op {
	case "source":
		return ctx.value(e.Source), nil
	case "const":
		return e.Value, nil
	case "run_meta":
		return ctx.value(e.Key), nil
	default:
		return evalTransform(e, ctx)
	}
}

func evalObject(e *model.ValueExpr, ctx *recordContext) (any, error) {
	if len(e.Fields) == 0 {
		return nil, fmt.Errorf("adapter: object requires fields")
	}
	return evalObjectFields(e.Fields, ctx)
}

func evalObjectFields(fields []model.Field, ctx *recordContext) (orderedObject, error) {
	out := orderedObject{}
	for _, field := range fields {
		value, err := evalObjectField(field, ctx)
		if err != nil {
			return nil, err
		}
		if value != nil || !field.OmitWhenNull {
			out = append(out, orderedField{name: field.Name, value: value})
		}
	}
	return out, nil
}

func evalConcat(e *model.ValueExpr, ctx *recordContext) (any, error) {
	var b strings.Builder
	for _, p := range e.Parts {
		v, err := evalExpr(&p, ctx)
		if err != nil {
			return nil, fmt.Errorf("adapter: %w", err)
		}
		b.WriteString(stringify(v))
	}
	return b.String(), nil
}

func evalTemplate(e *model.ValueExpr, ctx *recordContext) (any, error) {
	out := e.Template
	for _, m := range templateFieldRE.FindAllStringSubmatch(e.Template, -1) {
		v := ctx.value(m[1])
		out = strings.ReplaceAll(out, m[0], stringify(v))
	}
	return out, nil
}

func evalTimeFormat(e *model.ValueExpr, ctx *recordContext) (any, error) {
	value, err := evalExpr(e.Of, ctx)
	if err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	text, ok := value.(string)
	if !ok || text == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return nil, fmt.Errorf("adapter: format_time: %w", err)
	}
	return formatTimeLayout(parsed, e.Layout)
}

func evalCounter(e *model.ValueExpr, ctx *recordContext) (any, error) {
	n, err := counterValue(e.Of, ctx)
	if err != nil {
		return nil, err
	}
	// The contract requires width in [1,20]; validation has refused anything
	// else, so there is no default to apply here.
	return e.Prefix + fmt.Sprintf("%0*d", e.Width, n), nil
}

func (ctx recordContext) addEvent(ev *model.SimEvent) {
	ctx["seq"] = float64(ev.Seq)
	ctx["world_id"] = ev.WorldID
	ctx["entity_type"] = ev.EntityType
	ctx["entity_id"] = ev.EntityID
	ctx["channel"] = ev.Channel
	ctx["event_time"] = ev.EventTime
	ctx["observed_time"] = ev.ObservedTime
	ctx["value"] = ev.Value
	ctx.addOptionalEventFields(ev)
}

func (ctx recordContext) addOptionalEventFields(ev *model.SimEvent) {
	ctx["unit"], ctx["birth"] = nil, nil
	if ev.Unit != "" {
		ctx["unit"] = ev.Unit
	}
	if ev.Birth {
		ctx["birth"] = true
	}
}

func evalTransform(e *model.ValueExpr, ctx *recordContext) (any, error) {
	switch e.Op {
	case "object":
		return evalObject(e, ctx)
	case "concat":
		return evalConcat(e, ctx)
	case "template":
		return evalTemplate(e, ctx)
	case "format_time":
		return evalTimeFormat(e, ctx)
	case "counter":
		return evalCounter(e, ctx)
	}
	return nil, fmt.Errorf("adapter: unknown op %q", e.Op)
}

func evalObjectField(field model.Field, ctx *recordContext) (any, error) {
	value, err := evalExpr(&field.From, ctx)
	if err != nil {
		return nil, fmt.Errorf("adapter: object field %q: %w", field.Name, err)
	}
	return value, nil
}

func formatTimeLayout(t time.Time, layout string) (any, error) {
	switch layout {
	case "rfc3339_nano":
		return t.UTC().Format(time.RFC3339Nano), nil
	case "rfc3339":
		return t.UTC().Format(time.RFC3339), nil
	case "unix_nano":
		return t.UnixNano(), nil
	case "unix_millis":
		return t.UnixMilli(), nil
	case "unix_seconds":
		return t.Unix(), nil
	}
	return nil, fmt.Errorf("adapter: unknown layout %q", layout)
}

func counterValue(operand *model.ValueExpr, ctx *recordContext) (int64, error) {
	if operand == nil {
		return 0, nil
	}
	value, err := evalExpr(operand, ctx)
	if err != nil {
		return 0, fmt.Errorf("adapter: %w", err)
	}
	return counterInteger(value)
}

func counterInteger(value any) (int64, error) {
	switch x := value.(type) {
	case float64:
		return int64(x), nil
	case int64:
		return x, nil
	case json.Number:
		n, err := x.Int64()
		if err != nil {
			return 0, fmt.Errorf("adapter: %w", err)
		}
		return n, nil
	}
	return 0, nil
}
