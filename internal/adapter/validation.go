package adapter

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// crossCheck validates the parts the JSON schema cannot: transform arity,
// source names, time/counter operand types, run_meta keys, template
// placeholders, and the mandatory stable event identity.
func crossCheck(a *model.Adapter, src string) error {
	checks := adapterChecks{source: src}

	for i := range a.Preamble {
		if err := checks.checkTemplate(a.Preamble[i]); err != nil {
			return fmt.Errorf("preamble[%d]: %w", i, err)
		}
	}
	if err := checks.checkTemplate(a.Record); err != nil {
		return fmt.Errorf("adapter: %w", err)
	}
	for i := range a.Postamble {
		if err := checks.checkTemplate(a.Postamble[i]); err != nil {
			return fmt.Errorf("postamble[%d]: %w", i, err)
		}
	}

	identity := false
	for _, f := range a.Record.Fields {
		if hasIdentity(f.From) {
			identity = true
			break
		}
	}
	if !identity {
		return checks.bad("the record projection must carry a stable event identity derived from seq (a seq source or a counter with of=seq)")
	}
	return nil
}

type adapterChecks struct{ source string }

func (checks adapterChecks) bad(format string, args ...any) error {
	return fmt.Errorf("adapter: %s: %s", checks.source, fmt.Sprintf(format, args...))
}

func (checks adapterChecks) checkTemplate(t model.RecordTemplate) error {
	if len(t.Fields) == 0 {
		return checks.bad("record template requires at least one field")
	}
	if t.When != nil {
		if !nativeSource(t.When.Field) {
			return checks.bad("when guard references unknown field %q", t.When.Field)
		}
	}
	for _, f := range t.Fields {
		if err := checks.checkField(f); err != nil {
			return fmt.Errorf("adapter: %w", err)
		}
	}
	return nil
}

func (checks adapterChecks) checkField(f model.Field) error {
	if f.Name == "" {
		return checks.bad("field name must not be empty")
	}
	if err := checks.checkExpr(f.From); err != nil {
		return fmt.Errorf("adapter: %w", err)
	}
	return nil
}

func (checks adapterChecks) checkExpr(e model.ValueExpr) error {
	switch e.Op {
	case "source":
		if !nativeSource(e.Source) {
			return checks.bad("unknown source %q", e.Source)
		}
	case "const":
	case "object":
		if len(e.Fields) == 0 {
			return checks.bad("object requires fields")
		}
		for _, f := range e.Fields {
			if err := checks.checkField(f); err != nil {
				return fmt.Errorf("adapter: %w", err)
			}
		}
	case "concat":
		return checks.checkConcat(e)
	case "template":
		return checks.checkPlaceholders(e)
	case "format_time":
		return checks.checkTimeFormat(e)
	case "counter":
		if e.Of != nil && e.Of.Source != "seq" {
			return checks.bad("counter of must derive from seq")
		}
		if e.Width < 1 || e.Width > 20 {
			return checks.bad("counter width must be in [1,20]")
		}
	case "run_meta":
		if !runMetaKey(e.Key) {
			return checks.bad("unknown run_meta key %q", e.Key)
		}
	default:
		return checks.bad("unknown transform op %q", e.Op)
	}
	return nil
}

func (checks adapterChecks) checkConcat(e model.ValueExpr) error {
	if len(e.Parts) == 0 {
		return checks.bad("concat requires parts")
	}
	for _, p := range e.Parts {
		if err := checks.checkExpr(p); err != nil {
			return fmt.Errorf("adapter: %w", err)
		}
	}
	return nil
}

func (checks adapterChecks) checkPlaceholders(e model.ValueExpr) error {
	for _, m := range templateFieldRE.FindAllStringSubmatch(e.Template, -1) {
		if !nativeSource(m[1]) {
			return checks.bad("template placeholder {%s} is not a native event field", m[1])
		}
	}
	return nil
}

func (checks adapterChecks) checkTimeFormat(e model.ValueExpr) error {
	of := e.Of
	timeOK := of != nil && ((of.Source == "event_time" || of.Source == "observed_time") ||
		(of.Op == "run_meta" && (of.Key == "world_start_time" || of.Key == "world_end_time")))
	if !timeOK {
		return checks.bad("format_time requires of pointing at a time source")
	}
	switch e.Layout {
	case "rfc3339_nano", "rfc3339", "unix_nano", "unix_millis", "unix_seconds":
	default:
		return checks.bad("unknown time layout %q", e.Layout)
	}
	return nil
}

func hasIdentity(e model.ValueExpr) bool {
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

func nativeSource(name string) bool {
	switch name {
	case "seq", "world_id", "entity_type", "entity_id", "channel", "event_time", "observed_time", "value", "unit", "birth":
		return true
	}
	return false
}

func runMetaKey(name string) bool {
	switch name {
	case "run_id", "sim_version", "domain_id", "domain_version", "world_start_time", "world_end_time", "seed":
		return true
	}
	return false
}
