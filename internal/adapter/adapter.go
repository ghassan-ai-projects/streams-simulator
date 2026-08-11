// Package adapter implements the declarative output adapter engine. An
// adapter is data — a JSON file projecting native sim-event-v0.1 records
// into a consumer's wire format through a closed set of transforms. The
// binary contains no consumer-specific code, no consumer's field names and
// no consumer's framing rules; adding a consumer must never require a
// release.
package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
)

// Load reads and validates an adapter file. The validation covers the
// output-adapter schema plus adapter-specific cross-checks (transform
// arity, source names, identity preservation).
func Load(path string) (*model.Adapter, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("adapter: read %s: %w", path, err)
	}
	var doc any
	if err := model.DecodeBytes(raw, &doc); err != nil {
		return nil, fmt.Errorf("adapter: %s: not valid JSON: %w", path, err)
	}
	sch, err := jsonschema.Compile(mustAny(schemas.OutputAdapter()))
	if err != nil {
		return nil, fmt.Errorf("adapter: compile contract schema: %w", err)
	}
	if errs := sch.Validate(doc); len(errs) > 0 {
		return nil, fmt.Errorf("adapter: %s fails output-adapter-v0.1 validation:\n  %s", path, formatErrs(errs))
	}
	var a model.Adapter
	if err := json.Unmarshal(raw, &a); err != nil {
		return nil, fmt.Errorf("adapter: %s: decode: %w", path, err)
	}
	if err := crossCheck(&a, path); err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	return &a, nil
}

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
	checkExpr = func(e model.ValueExpr) error {
		switch e.Op {
		case "source":
			if !validSource[e.Source] {
				return bad("unknown source %q", e.Source)
			}
		case "const":
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
			if f.Name == "" {
				return bad("field name must not be empty")
			}
			if err := checkExpr(f.From); err != nil {
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
	identity := false
	for _, f := range a.Record.Fields {
		if f.From.Op == "source" && f.From.Source == "seq" {
			identity = true
		}
		if f.From.Op == "counter" && f.From.Of != nil && f.From.Of.Source == "seq" {
			identity = true
		}
	}
	if !identity {
		return bad("the record projection must carry a stable event identity derived from seq (a seq source or a counter with of=seq)")
	}
	return nil
}

var templateFieldRE = regexp.MustCompile(`\{([a-z_]+)\}`)

// Engine renders native events through an adapter.
type Engine struct {
	Adapter *model.Adapter
	rewrite []rewriteRule
	meta    map[string]any // run-level bindings
	started bool
	ended   bool
	items   int
}

type rewriteRule struct {
	from, to string
}

// NewEngine builds a rendering engine for an adapter with run-level
// bindings available to run_meta transforms.
func NewEngine(a *model.Adapter, meta map[string]any) (*Engine, error) {
	if meta == nil {
		meta = map[string]any{}
	}
	e := &Engine{Adapter: a, meta: meta}
	if a.EntityIDRewrite != nil {
		for _, r := range a.EntityIDRewrite.Replace {
			e.rewrite = append(e.rewrite, rewriteRule{from: r.From, to: r.To})
		}
	}
	return e, nil
}

// Begin starts a streaming adapter session and returns its framing/preamble
// records. The streaming session is the same lifecycle used by RenderRun;
// callers must not mix the two APIs on one Engine without a new session.
func (e *Engine) Begin() ([]string, error) {
	if e.started && !e.ended {
		return nil, fmt.Errorf("adapter: session already started")
	}
	e.started = true
	e.ended = false
	e.items = 0
	var out []string
	if e.Adapter.Encoding == "json-array" {
		out = append(out, "[")
	}
	for i := range e.Adapter.Preamble {
		line, err := e.renderTemplate(&e.Adapter.Preamble[i], nil)
		if err != nil {
			return nil, fmt.Errorf("adapter: preamble[%d]: %w", i, err)
		}
		if line != "" {
			out = append(out, e.frame(line))
		}
	}
	return out, nil
}

// RenderStreamRecord renders one event in an active streaming session. It
// applies array separators when the adapter declares json-array encoding.
func (e *Engine) RenderStreamRecord(ev *model.SimEvent) (string, error) {
	if !e.started || e.ended {
		return "", fmt.Errorf("adapter: stream record outside active session")
	}
	line, err := e.RenderRecord(ev)
	if err != nil || line == "" {
		return line, err
	}
	return e.frame(line), nil
}

// End closes a streaming adapter session and returns its postamble/framing
// records. worldEndNS is bound before postamble evaluation.
func (e *Engine) End(worldEndNS int64) ([]string, error) {
	if !e.started || e.ended {
		return nil, fmt.Errorf("adapter: session is not active")
	}
	e.meta["world_end_time"] = model.FormatTime(worldEndNS)
	var out []string
	for i := range e.Adapter.Postamble {
		line, err := e.renderTemplate(&e.Adapter.Postamble[i], nil)
		if err != nil {
			return nil, fmt.Errorf("adapter: postamble[%d]: %w", i, err)
		}
		if line != "" {
			out = append(out, e.frame(line))
		}
	}
	if e.Adapter.Encoding == "json-array" {
		out = append(out, "]")
	}
	e.ended = true
	return out, nil
}

func (e *Engine) frame(line string) string {
	if e.Adapter.Encoding != "json-array" {
		return line
	}
	if e.items > 0 {
		line = "," + line
	}
	e.items++
	return line
}

// Meta returns the run-level bindings (for the run layer to fill in).
func (e *Engine) Meta() map[string]any { return e.meta }

// rewriteID applies the declared entity-id rewrite.
func (e *Engine) rewriteID(id string) (string, error) {
	rw := e.Adapter.EntityIDRewrite
	if rw == nil {
		return id, nil
	}
	out := id
	for _, r := range e.rewrite {
		out = strings.ReplaceAll(out, r.from, r.to)
	}
	if rw.MaxLength > 0 && len(out) > rw.MaxLength {
		switch rw.OnViolation {
		case "truncate":
			out = out[:rw.MaxLength]
		case "hash_suffix":
			h := canonical.DigestBytes([]byte(out))[:16]
			out = out[:rw.MaxLength-len(h)] + h
		default: // fail
			return "", fmt.Errorf("adapter: entity id %q exceeds max_length %d after rewrite", id, rw.MaxLength)
		}
	}
	return out, nil
}

// RenderRun renders preamble, every event, and postamble into the adapter's
// encoding. worldEndNS is the run's final clock (for run_meta world_end_time).
func (e *Engine) RenderRun(events []model.SimEvent, worldEndNS int64) ([]byte, error) {
	var buf bytes.Buffer
	lines, err := e.Begin()
	if err != nil {
		return nil, err
	}
	writeLines := func(lines []string) {
		for _, line := range lines {
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}
	writeLines(lines)
	for i := range events {
		line, err := e.RenderStreamRecord(&events[i])
		if err != nil {
			return nil, fmt.Errorf("adapter: record %d: %w", i, err)
		}
		if line != "" {
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}
	lines, err = e.End(worldEndNS)
	if err != nil {
		return nil, err
	}
	writeLines(lines)
	return buf.Bytes(), nil
}

// RenderRecord renders a single event through the record template,
// returning "" when the when-guard excludes it. Used by the run layer for
// streaming delivery.
func (e *Engine) RenderRecord(ev *model.SimEvent) (string, error) {
	return e.renderTemplate(&e.Adapter.Record, ev)
}

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
	case "concat":
		var b strings.Builder
		for _, p := range e.Parts {
			v, err := evalExpr(&p, ctx)
			if err != nil {
				return nil, fmt.Errorf("adapter: %w", err)
			}
			b.WriteString(stringify(v))
		}
		return b.String(), nil
	case "template":
		out := e.Template
		for _, m := range templateFieldRE.FindAllStringSubmatch(e.Template, -1) {
			v := ctx.value(m[1])
			out = strings.ReplaceAll(out, m[0], stringify(v))
		}
		return out, nil
	case "format_time":
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
	case "counter":
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
	case "run_meta":
		return ctx.value(e.Key), nil
	}
	return nil, fmt.Errorf("adapter: unknown op %q", e.Op)
}

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
	default:
		raw, _ := json.Marshal(v)
		b.Write(raw)
	}
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
