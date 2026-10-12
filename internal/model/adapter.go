package model

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

	Raw []byte `json:"-"`
}

type IDRewrite struct {
	Replace     []Replacement `json:"replace,omitempty"`
	MaxLength   int           `json:"max_length,omitempty"`
	OnViolation string        `json:"on_violation,omitempty"`
}

type Replacement struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type RecordTemplate struct {
	When   *WhenClause `json:"when,omitempty"`
	Fields []Field     `json:"fields"`
}

type WhenClause struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value any    `json:"value,omitempty"`
}

type Field struct {
	Name         string    `json:"name"`
	From         ValueExpr `json:"from"`
	OmitWhenNull bool      `json:"omit_when_null,omitempty"`
}

type ValueExpr struct {
	Op       string      `json:"op"`
	Source   string      `json:"source,omitempty"`
	Value    any         `json:"value,omitempty"`
	Parts    []ValueExpr `json:"parts,omitempty"`
	Fields   []Field     `json:"fields,omitempty"`
	Template string      `json:"template,omitempty"`
	Layout   string      `json:"layout,omitempty"`
	Of       *ValueExpr  `json:"of,omitempty"`
	Prefix   string      `json:"prefix,omitempty"`
	Width    int         `json:"width,omitempty"`
	Key      string      `json:"key,omitempty"`
}

type Conformance struct {
	OutputSchema string `json:"output_schema,omitempty"`
	Golden       string `json:"golden,omitempty"`
}
