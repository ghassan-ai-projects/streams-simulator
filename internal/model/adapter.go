package model

// Adapter is the typed form of docs/contracts/output-adapter-v0.1.schema.json:
// a declarative projection from the native event into a consumer's wire
// format. Adapters are data, loaded through one path, exactly like domain
// specs.
type Adapter struct {
	ID              string           `json:"id"`
	Version         string           `json:"version"`
	Title           string           `json:"title,omitempty"`
	Description     string           `json:"description,omitempty"`
	Encoding        string           `json:"encoding"`
	EntityIDRewrite *IDRewrite       `json:"entity_id_rewrite,omitempty"`
	Preamble        []RecordTemplate `json:"preamble,omitempty"`
	Record          RecordTemplate   `json:"record"`
	Postamble       []RecordTemplate `json:"postamble,omitempty"`
	Conformance     *Conformance     `json:"conformance,omitempty"`
	// Raw is the validated source document embedded in run artifacts. It is
	// excluded from normal JSON encoding and digests.
	Raw []byte `json:"-"`
}

// IDRewrite narrows the simulator's entity-id alphabet to a consumer's.
type IDRewrite struct {
	Replace     []Replacement `json:"replace,omitempty"`
	MaxLength   int           `json:"max_length,omitempty"`
	OnViolation string        `json:"on_violation,omitempty"`
}

// Replacement is one character/sequence substitution.
type Replacement struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// RecordTemplate is one output record: an optional guard plus ordered fields.
type RecordTemplate struct {
	When   *WhenClause `json:"when,omitempty"`
	Fields []Field     `json:"fields"`
}

// WhenClause omits the record when the condition is false.
type WhenClause struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value any    `json:"value,omitempty"`
}

// Field is one named output field, projected from a value expression.
type Field struct {
	Name         string    `json:"name"`
	From         ValueExpr `json:"from"`
	OmitWhenNull bool      `json:"omit_when_null,omitempty"`
}

// ValueExpr is one op from the closed transform set. Deliberately not an
// expression language.
type ValueExpr struct {
	Op       string      `json:"op"`
	Source   string      `json:"source,omitempty"`
	Value    any         `json:"value,omitempty"`
	Parts    []ValueExpr `json:"parts,omitempty"`
	Template string      `json:"template,omitempty"`
	Layout   string      `json:"layout,omitempty"`
	Of       *ValueExpr  `json:"of,omitempty"`
	Prefix   string      `json:"prefix,omitempty"`
	Width    int         `json:"width,omitempty"`
	Key      string      `json:"key,omitempty"`
}

// Conformance proves the adapter correct without the consumer present.
type Conformance struct {
	OutputSchema string `json:"output_schema,omitempty"`
	Golden       string `json:"golden,omitempty"`
}
