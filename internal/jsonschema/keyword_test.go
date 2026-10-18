package jsonschema

import "testing"

func TestEveryViolationNamesTheKeywordThatFailedAndTheMemberItIsAbout(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, schema, value string
		path, keyword       string
		member              string
	}{
		{"type", `{"type": "string"}`, `1`, "", "type", ""},
		{"enum", `{"enum": ["a"]}`, `"b"`, "", "enum", ""},
		{"const", `{"const": 1}`, `2`, "", "const", ""},
		{"required", `{"type": "object", "required": ["resolution"]}`, `{}`, "/resolution", "required", "resolution"},
		{"additionalProperties", `{"type": "object", "additionalProperties": false}`, `{"extra": 1}`, "/extra", "additionalProperties", "extra"},
		{"pattern", `{"type": "string", "pattern": "^a"}`, `"b"`, "", "pattern", ""},
		{"minLength", `{"type": "string", "minLength": 2}`, `"b"`, "", "minLength", ""},
		{"maxLength", `{"type": "string", "maxLength": 1}`, `"bb"`, "", "maxLength", ""},
		{"minimum", `{"minimum": 1}`, `0`, "", "minimum", ""},
		{"maximum", `{"maximum": 1}`, `2`, "", "maximum", ""},
		{"exclusiveMinimum", `{"exclusiveMinimum": 1}`, `1`, "", "exclusiveMinimum", ""},
		{"exclusiveMaximum", `{"exclusiveMaximum": 1}`, `1`, "", "exclusiveMaximum", ""},
		{"multipleOf", `{"multipleOf": 2}`, `3`, "", "multipleOf", ""},
		{"minItems", `{"minItems": 1}`, `[]`, "", "minItems", ""},
		{"maxItems", `{"maxItems": 1}`, `[1, 2]`, "", "maxItems", ""},
		{"uniqueItems", `{"uniqueItems": true}`, `[1, 1]`, "", "uniqueItems", ""},
		{"contains", `{"contains": {"const": 3}}`, `[1]`, "", "contains", ""},
		{"oneOf", `{"oneOf": [{"type": "string"}, {"type": "number"}]}`, `true`, "", "oneOf", ""},
		{"anyOf", `{"anyOf": [{"type": "string"}]}`, `1`, "", "anyOf", ""},
		{"not", `{"not": {"type": "string"}}`, `"a"`, "", "not", ""},
		{"format", `{"type": "string", "format": "date-time"}`, `"yesterday"`, "", "format", ""},
		{"a keyword inside a property", `{"type": "object", "properties": {"n": {"minimum": 5}}}`, `{"n": 1}`, "/n", "minimum", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			errs := validateDoc(t, compileDoc(t, tc.schema), tc.value)
			if len(errs) != 1 || errs[0].Path != tc.path || errs[0].Keyword != tc.keyword || errs[0].Member != tc.member {
				t.Fatalf("errors = %+v, want one at %q for keyword %s and member %q", errs, tc.path, tc.keyword, tc.member)
			}
		})
	}
}
