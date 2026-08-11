// Package jsonschema implements the subset of JSON Schema draft 2020-12
// that the simulator's own contract schemas (docs/contracts/*.schema.json)
// actually use, with local $ref resolution only. It exists so that domain
// specs, adapters, verdicts and run artifacts can be validated against their
// committed schemas without pulling a schema-engine dependency into the
// zero-external-dependency core.
//
// Supported keywords: type, enum, const, properties, required,
// additionalProperties, items, minItems, maxItems, uniqueItems, contains,
// minLength, maxLength, pattern, minimum, maximum, exclusiveMinimum,
// exclusiveMaximum, multipleOf, oneOf, anyOf, allOf, not, if/then/else,
// $defs, $ref (local), format (date-time, regex, uri, hostname). Unknown
// formats and annotation-only keywords are ignored, per the specification.
package jsonschema

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

// Error is one validation failure with a JSON-pointer path into the document.
type Error struct {
	Path string
	Msg  string
}

func (e Error) Error() string {
	if e.Path == "" {
		return e.Msg
	}
	return e.Path + ": " + e.Msg
}

// Schema is a compiled JSON Schema.
type Schema struct {
	types             []string
	enum              []any
	constVal          any
	hasConst          bool
	properties        map[string]*Schema
	required          []string
	additional        *Schema // nil means allowed (no constraint); NonNil means constraint
	addAllowed        bool
	items             *Schema
	minItems          int
	hasMinItems       bool
	maxItems          int
	hasMaxItems       bool
	uniqueItems       bool
	contains          *Schema
	minLen            int
	hasMinLen         bool
	maxLen            int
	hasMaxLen         bool
	pattern           *regexp.Regexp
	min               float64
	hasMin            bool
	max               float64
	hasMax            bool
	exMin             float64
	hasExMin          bool
	exMax             float64
	hasExMax          bool
	multipleOf        float64
	hasMult           bool
	oneOf             []*Schema
	anyOf             []*Schema
	allOf             []*Schema
	not               *Schema
	ifS, thenS, elseS *Schema
	format            string
}

// Compile builds a Schema from a decoded JSON Schema document. Recursive
// $defs references are supported: every top-level $def is pre-registered as
// a placeholder and filled after the root schema compiles, so self-referring
// definitions (like the adapter's valueExpr) compile fine and only
// terminate at validation time, walking finite values.
func Compile(doc any) (*Schema, error) {
	m, ok := doc.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("jsonschema: schema document must be an object")
	}
	defs := map[string]any{}
	if d, ok := m["$defs"].(map[string]any); ok {
		defs = d
	}
	reg := map[string]*Schema{}
	for name := range defs {
		reg[name] = &Schema{addAllowed: true}
	}
	root, err := compile(m, defs, reg, 0)
	if err != nil {
		return nil, fmt.Errorf("jsonschema: %w", err)
	}
	for name, raw := range defs {
		rm, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("jsonschema: $defs/%s must be an object", name)
		}
		if err := fill(reg[name], rm, defs, reg, 0); err != nil {
			return nil, fmt.Errorf("jsonschema: $defs/%s: %w", name, err)
		}
	}
	return root, nil
}

func compile(m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) (*Schema, error) {
	if depth > 64 {
		return nil, fmt.Errorf("jsonschema: schema nesting too deep")
	}
	if r, ok := m["$ref"].(string); ok {
		name, err := refName(r)
		if err != nil {
			return nil, fmt.Errorf("jsonschema: %w", err)
		}
		if s, ok := reg[name]; ok {
			return s, nil
		}
		raw, ok := defs[name]
		if !ok {
			return nil, fmt.Errorf("jsonschema: unknown $defs entry %q", name)
		}
		rm, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("jsonschema: $defs/%s is not an object", name)
		}
		s := &Schema{addAllowed: true}
		reg[name] = s
		if err := fill(s, rm, defs, reg, depth+1); err != nil {
			return nil, fmt.Errorf("jsonschema: %w", err)
		}
		return s, nil
	}
	s := &Schema{addAllowed: true}
	if err := fill(s, m, defs, reg, depth+1); err != nil {
		return nil, fmt.Errorf("jsonschema: %w", err)
	}
	return s, nil
}

// fill populates a Schema from its JSON form (the compile body, without the
// $ref shortcut).
func fill(s *Schema, m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) error {
	if t, ok := m["type"]; ok {
		switch tv := t.(type) {
		case string:
			s.types = []string{tv}
		case []any:
			for _, e := range tv {
				str, ok := e.(string)
				if !ok {
					return fmt.Errorf("jsonschema: type array must hold strings")
				}
				s.types = append(s.types, str)
			}
		default:
			return fmt.Errorf("jsonschema: type must be a string or array")
		}
	}
	if e, ok := m["enum"]; ok {
		arr, ok := e.([]any)
		if !ok {
			return fmt.Errorf("jsonschema: enum must be an array")
		}
		s.enum = arr
	}
	if c, ok := m["const"]; ok {
		s.constVal, s.hasConst = c, true
	}
	if p, ok := m["properties"].(map[string]any); ok {
		s.properties = make(map[string]*Schema, len(p))
		for name, sub := range p {
			subm, ok := sub.(map[string]any)
			if !ok {
				return fmt.Errorf("jsonschema: property %q must be an object", name)
			}
			cs, err := compile(subm, defs, reg, depth+1)
			if err != nil {
				return fmt.Errorf("jsonschema: property %q: %w", name, err)
			}
			s.properties[name] = cs
		}
	}
	if r, ok := m["required"]; ok {
		arr, ok := r.([]any)
		if !ok {
			return fmt.Errorf("jsonschema: required must be an array")
		}
		for _, e := range arr {
			str, ok := e.(string)
			if !ok {
				return fmt.Errorf("jsonschema: required entries must be strings")
			}
			s.required = append(s.required, str)
		}
		sort.Strings(s.required)
	}
	switch add := m["additionalProperties"].(type) {
	case bool:
		s.addAllowed = add
	case map[string]any:
		cs, err := compile(add, defs, reg, depth+1)
		if err != nil {
			return fmt.Errorf("jsonschema: %w", err)
		}
		s.additional = cs
	case nil:
	default:
		return fmt.Errorf("jsonschema: additionalProperties must be a boolean or object")
	}
	if it, ok := m["items"].(map[string]any); ok {
		cs, err := compile(it, defs, reg, depth+1)
		if err != nil {
			return fmt.Errorf("jsonschema: %w", err)
		}
		s.items = cs
	}
	if v, ok := m["minItems"]; ok {
		s.minItems = toInt(v)
		s.hasMinItems = true
	}
	if v, ok := m["maxItems"]; ok {
		s.maxItems = toInt(v)
		s.hasMaxItems = true
	}
	if v, ok := m["uniqueItems"]; ok {
		s.uniqueItems = v.(bool)
	}
	if c, ok := m["contains"].(map[string]any); ok {
		cs, err := compile(c, defs, reg, depth+1)
		if err != nil {
			return fmt.Errorf("jsonschema: %w", err)
		}
		s.contains = cs
	}
	if v, ok := m["minLength"]; ok {
		s.minLen = toInt(v)
		s.hasMinLen = true
	}
	if v, ok := m["maxLength"]; ok {
		s.maxLen = toInt(v)
		s.hasMaxLen = true
	}
	if p, ok := m["pattern"].(string); ok {
		re, err := regexp.Compile(p)
		if err != nil {
			return fmt.Errorf("jsonschema: bad pattern %q: %w", p, err)
		}
		s.pattern = re
	}
	if v, ok := m["minimum"]; ok {
		s.min, s.hasMin = toFloat(v)
	}
	if v, ok := m["maximum"]; ok {
		s.max, s.hasMax = toFloat(v)
	}
	if v, ok := m["exclusiveMinimum"]; ok {
		s.exMin, s.hasExMin = toFloat(v)
	}
	if v, ok := m["exclusiveMaximum"]; ok {
		s.exMax, s.hasExMax = toFloat(v)
	}
	if v, ok := m["multipleOf"]; ok {
		s.multipleOf, s.hasMult = toFloat(v)
	}
	for _, kw := range []string{"oneOf", "anyOf", "allOf"} {
		if arr, ok := m[kw].([]any); ok {
			var out []*Schema
			for _, sub := range arr {
				subm, ok := sub.(map[string]any)
				if !ok {
					return fmt.Errorf("jsonschema: %s entries must be objects", kw)
				}
				cs, err := compile(subm, defs, reg, depth+1)
				if err != nil {
					return fmt.Errorf("jsonschema: %w", err)
				}
				out = append(out, cs)
			}
			switch kw {
			case "oneOf":
				s.oneOf = out
			case "anyOf":
				s.anyOf = out
			case "allOf":
				s.allOf = out
			}
		}
	}
	if n, ok := m["not"].(map[string]any); ok {
		cs, err := compile(n, defs, reg, depth+1)
		if err != nil {
			return fmt.Errorf("jsonschema: %w", err)
		}
		s.not = cs
	}
	if cond, ok := m["if"].(map[string]any); ok {
		cs, err := compile(cond, defs, reg, depth+1)
		if err != nil {
			return fmt.Errorf("jsonschema: %w", err)
		}
		s.ifS = cs
		if then, ok := m["then"].(map[string]any); ok {
			tcs, err := compile(then, defs, reg, depth+1)
			if err != nil {
				return fmt.Errorf("jsonschema: %w", err)
			}
			s.thenS = tcs
		}
		if els, ok := m["else"].(map[string]any); ok {
			ecs, err := compile(els, defs, reg, depth+1)
			if err != nil {
				return fmt.Errorf("jsonschema: %w", err)
			}
			s.elseS = ecs
		}
	}
	if f, ok := m["format"].(string); ok {
		s.format = f
	}
	return nil
}

func refName(ref string) (string, error) {
	if !strings.HasPrefix(ref, "#/$defs/") {
		return "", fmt.Errorf("jsonschema: unsupported $ref %q (local #/$defs refs only)", ref)
	}
	return strings.TrimPrefix(ref, "#/$defs/"), nil
}

func toInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case uint64:
		// #nosec G115 -- schema integer keywords are small literals.
		return int(n)
	case jsonNumber:
		i, err := strconv.Atoi(n.String())
		if err != nil {
			return 0
		}
		return i
	default:
		return 0
	}
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case jsonNumber:
		f, err := n.Float64()
		return f, err == nil
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint64:
		return float64(n), true
	}
	return 0, false
}

// Validate checks v against the schema and returns every failure found.
func (s *Schema) Validate(v any) []Error {
	var errs []Error
	s.validate(v, "", &errs)
	return errs
}

func (s *Schema) validate(v any, path string, errs *[]Error) {
	fail := func(msg string) {
		*errs = append(*errs, Error{Path: path, Msg: msg})
	}
	if len(s.types) > 0 && !typeMatches(s.types, v) {
		fail(fmt.Sprintf("expected type %s, got %s", strings.Join(s.types, "|"), typeName(v)))
		return // type failure short-circuits value keywords
	}
	if len(s.enum) > 0 && !enumContains(s.enum, v) {
		fail("value not in enum")
	}
	if s.hasConst && !jsonEqual(s.constVal, v) {
		fail("value does not match const")
	}
	if s.pattern != nil {
		if str, ok := v.(string); ok && !s.pattern.MatchString(str) {
			fail(fmt.Sprintf("string does not match pattern %s", s.pattern))
		}
	}
	if str, ok := v.(string); ok {
		if s.hasMinLen && len([]rune(str)) < s.minLen {
			fail(fmt.Sprintf("string shorter than minLength %d", s.minLen))
		}
		if s.hasMaxLen && len([]rune(str)) > s.maxLen {
			fail(fmt.Sprintf("string longer than maxLength %d", s.maxLen))
		}
		if s.format != "" {
			if msg := checkFormat(s.format, str); msg != "" {
				fail(msg)
			}
		}
	}
	if f, ok := asFloat(v); ok {
		if s.hasMin && f < s.min {
			fail(fmt.Sprintf("number %v below minimum %v", f, s.min))
		}
		if s.hasMax && f > s.max {
			fail(fmt.Sprintf("number %v above maximum %v", f, s.max))
		}
		if s.hasExMin && f <= s.exMin {
			fail(fmt.Sprintf("number %v not above exclusiveMinimum %v", f, s.exMin))
		}
		if s.hasExMax && f >= s.exMax {
			fail(fmt.Sprintf("number %v not below exclusiveMaximum %v", f, s.exMax))
		}
		if s.hasMult && s.multipleOf != 0 {
			q := f / s.multipleOf
			if q != float64(int64(q)) {
				fail(fmt.Sprintf("number %v is not a multiple of %v", f, s.multipleOf))
			}
		}
	}
	switch x := v.(type) {
	case map[string]any:
		s.validateObject(x, path, errs)
	case []any:
		s.validateArray(x, path, errs)
	}
	for _, sub := range s.allOf {
		sub.validate(v, path, errs)
	}
	if len(s.oneOf) > 0 {
		matches := 0
		for _, sub := range s.oneOf {
			if len(sub.Validate(v)) == 0 {
				matches++
			}
		}
		if matches != 1 {
			fail(fmt.Sprintf("value must match exactly one oneOf branch (matched %d)", matches))
		}
	}
	if len(s.anyOf) > 0 {
		matched := false
		for _, sub := range s.anyOf {
			if len(sub.Validate(v)) == 0 {
				matched = true
				break
			}
		}
		if !matched {
			fail("value matches no anyOf branch")
		}
	}
	if s.not != nil {
		if len(s.not.Validate(v)) == 0 {
			fail("value must not match the not schema")
		}
	}
	if s.ifS != nil {
		if len(s.ifS.Validate(v)) == 0 {
			if s.thenS != nil {
				s.thenS.validate(v, path, errs)
			}
		} else if s.elseS != nil {
			s.elseS.validate(v, path, errs)
		}
	}
}

func (s *Schema) validateObject(obj map[string]any, path string, errs *[]Error) {
	for _, name := range s.required {
		if _, ok := obj[name]; !ok {
			*errs = append(*errs, Error{Path: path, Msg: fmt.Sprintf("missing required property %q", name)})
		}
	}
	// Properties validate in sorted order so the error report is stable
	// across runs (map iteration order is not).
	var names []string
	// determinism-safe: collected here, sorted below before any output.
	for name := range s.properties {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		sub := s.properties[name]
		val, ok := obj[name]
		if !ok {
			continue
		}
		sub.validate(val, joinPath(path, name), errs)
	}
	// Additional-property checks iterate in sorted order so validation
	// output is deterministic across runs (map order is not).
	var extra []string
	// determinism-safe: collected here, sorted below before any output.
	for name := range obj {
		if _, known := s.properties[name]; !known {
			extra = append(extra, name)
		}
	}
	sort.Strings(extra)
	for _, name := range extra {
		val := obj[name]
		switch {
		case s.additional != nil:
			s.additional.validate(val, joinPath(path, name), errs)
		case !s.addAllowed:
			*errs = append(*errs, Error{Path: path, Msg: fmt.Sprintf("additional property %q not allowed", name)})
		}
	}
}

func (s *Schema) validateArray(arr []any, path string, errs *[]Error) {
	if s.hasMinItems && len(arr) < s.minItems {
		*errs = append(*errs, Error{Path: path, Msg: fmt.Sprintf("array shorter than minItems %d", s.minItems)})
	}
	if s.hasMaxItems && len(arr) > s.maxItems {
		*errs = append(*errs, Error{Path: path, Msg: fmt.Sprintf("array longer than maxItems %d", s.maxItems)})
	}
	if s.uniqueItems {
		seen := map[string]bool{}
		for _, e := range arr {
			key, err := canonical.MarshalString(e)
			if err == nil && seen[key] {
				*errs = append(*errs, Error{Path: path, Msg: "array items must be unique"})
				break
			}
			seen[key] = true
		}
	}
	if s.contains != nil {
		matched := false
		for i, e := range arr {
			if len(s.contains.Validate(e)) == 0 {
				matched = true
				break
			}
			_ = i
		}
		if !matched {
			*errs = append(*errs, Error{Path: path, Msg: "no array item matches contains"})
		}
	}
	if s.items != nil {
		for i, e := range arr {
			s.items.validate(e, joinPath(path, fmt.Sprintf("%d", i)), errs)
		}
	}
}

func joinPath(base, name string) string {
	if base == "" {
		return "/" + escapePointer(name)
	}
	return base + "/" + escapePointer(name)
}

func escapePointer(s string) string {
	s = strings.ReplaceAll(s, "~", "~0")
	return strings.ReplaceAll(s, "/", "~1")
}

func typeMatches(types []string, v any) bool {
	for _, t := range types {
		if matchesType(t, v) {
			return true
		}
	}
	return false
}

func matchesType(t string, v any) bool {
	switch t {
	case "null":
		return v == nil
	case "boolean":
		_, ok := v.(bool)
		return ok
	case "object":
		_, ok := v.(map[string]any)
		return ok
	case "array":
		_, ok := v.([]any)
		return ok
	case "string":
		_, ok := v.(string)
		return ok
	case "number":
		_, ok := asFloat(v)
		return ok
	case "integer":
		f, ok := asFloat(v)
		if !ok {
			return false
		}
		return f == float64(int64(f))
	}
	return false
}

type jsonNumber interface {
	Float64() (float64, error)
	String() string
}

func asFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint64:
		return float64(n), true
	case jsonNumber:
		f, err := n.Float64()
		return f, err == nil
	}
	return 0, false
}

func typeName(v any) string {
	switch v.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case map[string]any:
		return "object"
	case []any:
		return "array"
	case string:
		return "string"
	default:
		if _, ok := asFloat(v); ok {
			return "number"
		}
		return "unknown"
	}
}

func enumContains(enum []any, v any) bool {
	for _, e := range enum {
		if jsonEqual(e, v) {
			return true
		}
	}
	return false
}

func jsonEqual(a, b any) bool {
	sa, err := canonical.MarshalString(a)
	if err != nil {
		return false
	}
	sb, err := canonical.MarshalString(b)
	if err != nil {
		return false
	}
	return sa == sb
}

func checkFormat(format, s string) string {
	switch format {
	case "date-time":
		if _, err := time.Parse(time.RFC3339Nano, s); err != nil {
			return "string is not a valid RFC 3339 date-time"
		}
	case "regex":
		if _, err := regexp.Compile(s); err != nil {
			return "string is not a valid regular expression"
		}
	case "uri":
		if !strings.Contains(s, ":") {
			return "string is not a valid URI"
		}
	case "hostname":
		if s == "" || strings.ContainsAny(s, " /") {
			return "string is not a valid hostname"
		}
	}
	return ""
}
