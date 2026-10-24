package domain

import "slices"

const CategoryDelivery = "delivery"

type ParamInfo struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Minimum *float64 `json:"minimum,omitempty"`
	Maximum *float64 `json:"maximum,omitempty"`
	Values  []string `json:"values,omitempty"`
}

type KindInfo struct {
	Kind        string      `json:"kind"`
	Category    string      `json:"category"`
	Description string      `json:"description"`
	RequiresEnd bool        `json:"requires_end"`
	Params      []ParamInfo `json:"params"`
}

var descriptions = map[string]string{
	DuplicateBurst: "Delivers some records twice.",
	IDReuse:        "Gives some records the identity of an earlier record.",
	Reorder:        "Delivers records out of order, displaced by a few positions.",
	DelayTail:      "Delivers records late, with a long-tailed delay.",
	GrossBackfill:  "Holds records back until its window ends, then delivers them in one late batch.",
	Drop:           "Loses records in transport; the consumer never receives them.",
	ClockSkew:      "Shifts the timestamps of delivered records by a fixed offset.",
	NonMonotonic:   "Makes timestamps go backwards for some records.",
	OutOfEnum:      "Replaces an enumerated value with one outside its declared set.",
	OutOfRange:     "Moves a numeric value outside its declared range.",
	UnitMismatch:   "Reports a value in a different unit than the one declared.",
	Oversize:       "Enlarges the payload of delivered records.",
	Malformed:      "Delivers records that do not parse.",
	NaNInf:         "Replaces numeric values with not-a-number or infinity.",
	Storm:          "Multiplies the delivery rate by repeating records.",
	ProducerFlap:   "Silences the producer during its window, then delivers what it held and announces a restart.",
	TimeEncoding:   "Rewrites the time of records in a different encoding.",
	PrecisionEdge:  "Truncates the precision of record times.",
	InjectionProbe: "Delivers records that carry text a consumer must treat as data, not instructions.",
}

func Describe() []KindInfo {
	kinds := make([]KindInfo, 0, len(catalogNames))
	for _, name := range catalogNames {
		kinds = append(kinds, KindInfo{Kind: name, Category: CategoryDelivery, Description: descriptions[name], RequiresEnd: windowEndRequired(name), Params: paramInfos(name)})
	}
	return kinds
}

func windowEndRequired(kind string) bool {
	return kind == GrossBackfill || kind == ProducerFlap
}

func paramInfos(kind string) []ParamInfo {
	infos := []ParamInfo{}
	for _, spec := range paramSpecs[kind] {
		infos = append(infos, spec.info())
	}
	return infos
}

func (spec paramSpec) info() ParamInfo {
	info := ParamInfo{Name: spec.name, Type: spec.kind, Values: slices.Clone(spec.values)}
	if spec.kind == ParamNumber || spec.kind == ParamInteger {
		lo, hi := spec.lo, spec.hi
		info.Minimum, info.Maximum = &lo, &hi
	}
	return info
}
