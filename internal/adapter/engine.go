package adapter

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

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
