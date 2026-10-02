package adapter

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// crossCheck validates the parts the JSON schema cannot: transform arity,
// source names, time/counter operand types, run_meta keys, template
// placeholders, and the mandatory stable event identity.
func crossCheck(a *model.Adapter, src string) error {
	bad := func(format string, args ...any) error {
		return fmt.Errorf("adapter: %s: %s", src, fmt.Sprintf(format, args...))
	}
	validSource := map[string]bool{
		"seq": true, "world_id": true, "entity_type": true, "entity_id": true,
		"channel": true, "event_time": true, "observed_time": true,
		"value": true, "unit": true, "birth": true,
	}
	timeSources := map[string]bool{"event_time": true, "observed_time": true}
	metaKeys := map[string]bool{
		"run_id": true, "sim_version": true, "domain_id": true,
		"domain_version": true, "world_start_time": true,
		"world_end_time": true, "seed": true,
	}
	var checkExpr func(e model.ValueExpr) error
	var checkField func(f model.Field) error
	checkExpr = func(e model.ValueExpr) error {
		switch e.Op {
		case "source":
			if !validSource[e.Source] {
				return bad("unknown source %q", e.Source)
			}
		case "const":
		case "object":
			if len(e.Fields) == 0 {
				return bad("object requires fields")
			}
			for _, f := range e.Fields {
				if err := checkField(f); err != nil {
					return fmt.Errorf("adapter: %w", err)
				}
			}
		case "concat":
			if len(e.Parts) == 0 {
				return bad("concat requires parts")
			}
			for _, p := range e.Parts {
				if err := checkExpr(p); err != nil {
					return fmt.Errorf("adapter: %w", err)
				}
			}
		case "template":
			for _, m := range templateFieldRE.FindAllStringSubmatch(e.Template, -1) {
				if !validSource[m[1]] {
					return bad("template placeholder {%s} is not a native event field", m[1])
				}
			}
		case "format_time":
			of := e.Of
			timeOK := of != nil && (timeSources[of.Source] ||
				(of.Op == "run_meta" && (of.Key == "world_start_time" || of.Key == "world_end_time")))
			if !timeOK {
				return bad("format_time requires of pointing at a time source")
			}
			switch e.Layout {
			case "rfc3339_nano", "rfc3339", "unix_nano", "unix_millis", "unix_seconds":
			default:
				return bad("unknown time layout %q", e.Layout)
			}
		case "counter":
			if e.Of != nil && e.Of.Source != "seq" {
				return bad("counter of must derive from seq")
			}
			if e.Width < 1 || e.Width > 20 {
				return bad("counter width must be in [1,20]")
			}
		case "run_meta":
			if !metaKeys[e.Key] {
				return bad("unknown run_meta key %q", e.Key)
			}
		default:
			return bad("unknown transform op %q", e.Op)
		}
		return nil
	}
	checkField = func(f model.Field) error {
		if f.Name == "" {
			return bad("field name must not be empty")
		}
		if err := checkExpr(f.From); err != nil {
			return fmt.Errorf("adapter: %w", err)
		}
		return nil
	}
	checkTemplate := func(t model.RecordTemplate) error {
		if len(t.Fields) == 0 {
			return bad("record template requires at least one field")
		}
		if t.When != nil {
			if !validSource[t.When.Field] {
				return bad("when guard references unknown field %q", t.When.Field)
			}
		}
		for _, f := range t.Fields {
			if err := checkField(f); err != nil {
				return fmt.Errorf("adapter: %w", err)
			}
		}
		return nil
	}
	for i := range a.Preamble {
		if err := checkTemplate(a.Preamble[i]); err != nil {
			return fmt.Errorf("preamble[%d]: %w", i, err)
		}
	}
	if err := checkTemplate(a.Record); err != nil {
		return fmt.Errorf("adapter: %w", err)
	}
	for i := range a.Postamble {
		if err := checkTemplate(a.Postamble[i]); err != nil {
			return fmt.Errorf("postamble[%d]: %w", i, err)
		}
	}
	// The per-event projection must carry a stable identity derived from seq.
	var hasIdentity func(e model.ValueExpr) bool
	hasIdentity = func(e model.ValueExpr) bool {
		if e.Op == "source" && e.Source == "seq" {
			return true
		}
		if e.Op == "counter" && e.Of != nil && e.Of.Op == "source" && e.Of.Source == "seq" {
			return true
		}
		if e.Op == "object" {
			for _, f := range e.Fields {
				if hasIdentity(f.From) {
					return true
				}
			}
		}
		return false
	}
	identity := false
	for _, f := range a.Record.Fields {
		if hasIdentity(f.From) {
			identity = true
			break
		}
	}
	if !identity {
		return bad("the record projection must carry a stable event identity derived from seq (a seq source or a counter with of=seq)")
	}
	return nil
}
