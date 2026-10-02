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
	"regexp"
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
