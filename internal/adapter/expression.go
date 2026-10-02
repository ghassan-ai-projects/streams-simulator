package adapter

import (
	"encoding/json"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"strings"
	"time"
)

// recordContext exposes native event fields to the transforms.
type recordContext map[string]any

func newContext(ev *model.SimEvent, meta map[string]any) *recordContext {
	ctx := recordContext{}
	for k, v := range meta {
		ctx[k] = v
	}
	if ev != nil {
		ctx["seq"] = float64(ev.Seq)
		ctx["world_id"] = ev.WorldID
		ctx["entity_type"] = ev.EntityType
		ctx["entity_id"] = ev.EntityID
		ctx["channel"] = ev.Channel
		ctx["event_time"] = ev.EventTime
		ctx["observed_time"] = ev.ObservedTime
		ctx["value"] = ev.Value
		if ev.Unit != "" {
			ctx["unit"] = ev.Unit
		} else {
			ctx["unit"] = nil
		}
		if ev.Birth {
			ctx["birth"] = true
		} else {
			ctx["birth"] = nil
		}
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
	case "run_meta":
		return ctx.value(e.Key), nil
	}
	return nil, fmt.Errorf("adapter: unknown op %q", e.Op)
}

func evalObject(e *model.ValueExpr, ctx *recordContext) (any, error) {
	if len(e.Fields) == 0 {
		return nil, fmt.Errorf("adapter: object requires fields")
	}
	out := orderedObject{}
	for _, f := range e.Fields {
		v, err := evalExpr(&f.From, ctx)
		if err != nil {
			return nil, fmt.Errorf("adapter: object field %q: %w", f.Name, err)
		}
		if v == nil && f.OmitWhenNull {
			continue
		}
		out = append(out, orderedField{name: f.Name, value: v})
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
	v, err := evalExpr(e.Of, ctx)
	if err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return nil, fmt.Errorf("adapter: format_time: %w", err)
	}
	switch e.Layout {
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
	return nil, fmt.Errorf("adapter: unknown layout %q", e.Layout)
}

func evalCounter(e *model.ValueExpr, ctx *recordContext) (any, error) {
	n := int64(0)
	if e.Of != nil {
		v, err := evalExpr(e.Of, ctx)
		if err != nil {
			return nil, fmt.Errorf("adapter: %w", err)
		}
		switch x := v.(type) {
		case float64:
			n = int64(x)
		case int64:
			n = x
		case json.Number:
			f, err := x.Int64()
			if err != nil {
				return nil, fmt.Errorf("adapter: %w", err)
			}
			n = f
		}
	}
	width := e.Width
	if width <= 0 {
		width = 6
	}
	return e.Prefix + fmt.Sprintf("%0*d", width, n), nil
}
