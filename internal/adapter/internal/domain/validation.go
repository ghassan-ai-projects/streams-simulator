package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// crossCheck validates the parts the JSON schema cannot: transform arity,
// source names, time/counter operand types, run_meta keys, template
// placeholders, and the mandatory stable event identity.
func crossCheck(a *model.Adapter, src string) error {
	checks := adapterChecks{source: src}
	if err := checks.checkTemplates(a); err != nil {
		return err
	}
	return checks.requireIdentity(a.Record.Fields)
}

type adapterChecks struct{ source string }

func (checks adapterChecks) bad(format string, args ...any) error {
	return fmt.Errorf("adapter: %s: %s", checks.source, fmt.Sprintf(format, args...))
}

func (checks adapterChecks) checkTemplate(t model.RecordTemplate) error {
	if len(t.Fields) == 0 {
		return checks.bad("record template requires at least one field")
	}
	if t.When != nil && !nativeSource(t.When.Field) {
		return checks.bad("when guard references unknown field %q", t.When.Field)
	}
	return checks.checkFields(t.Fields)
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
		return checks.checkSource(e.Source)
	case "const":
		return nil
	case "object":
		return checks.checkObjectFields(e.Fields)
	default:
		return checks.checkTransform(e)
	}
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
	switch e.Op {
	case "source":
		return e.Source == "seq"
	case "counter":
		return e.Of != nil && e.Of.Op == "source" && e.Of.Source == "seq"
	case "object":
		return fieldsCarryIdentity(e.Fields)
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

func (checks adapterChecks) checkTemplates(a *model.Adapter) error {
	if err := checks.checkFraming(a.Preamble, "preamble"); err != nil {
		return err
	}
	if err := checks.checkTemplate(a.Record); err != nil {
		return fmt.Errorf("adapter: %w", err)
	}
	return checks.checkFraming(a.Postamble, "postamble")
}

func (checks adapterChecks) checkFraming(templates []model.RecordTemplate, kind string) error {
	for i, template := range templates {
		if err := checks.checkTemplate(template); err != nil {
			return fmt.Errorf("%s[%d]: %w", kind, i, err)
		}
	}
	return nil
}

func (checks adapterChecks) requireIdentity(fields []model.Field) error {
	if fieldsCarryIdentity(fields) {
		return nil
	}
	return checks.bad("the record projection must carry a stable event identity derived from seq (a seq source or a counter with of=seq)")
}

func fieldsCarryIdentity(fields []model.Field) bool {
	for _, field := range fields {
		if hasIdentity(field.From) {
			return true
		}
	}
	return false
}

func (checks adapterChecks) checkFields(fields []model.Field) error {
	for _, field := range fields {
		if err := checks.checkField(field); err != nil {
			return fmt.Errorf("adapter: %w", err)
		}
	}
	return nil
}

func (checks adapterChecks) checkSource(source string) error {
	if !nativeSource(source) {
		return checks.bad("unknown source %q", source)
	}
	return nil
}

func (checks adapterChecks) checkObjectFields(fields []model.Field) error {
	if len(fields) == 0 {
		return checks.bad("object requires fields")
	}
	return checks.checkFields(fields)
}

func (checks adapterChecks) checkTransform(e model.ValueExpr) error {
	switch e.Op {
	case "concat":
		return checks.checkConcat(e)
	case "template":
		return checks.checkPlaceholders(e)
	case "format_time":
		return checks.checkTimeFormat(e)
	case "counter":
		return checks.checkCounter(e)
	case "run_meta":
		return checks.checkRunMeta(e.Key)
	}
	return checks.bad("unknown transform op %q", e.Op)
}

func (checks adapterChecks) checkCounter(e model.ValueExpr) error {
	if e.Of != nil && e.Of.Source != "seq" {
		return checks.bad("counter of must derive from seq")
	}
	if e.Width < 1 || e.Width > 20 {
		return checks.bad("counter width must be in [1,20]")
	}
	return nil
}

func (checks adapterChecks) checkRunMeta(key string) error {
	if !runMetaKey(key) {
		return checks.bad("unknown run_meta key %q", key)
	}
	return nil
}
