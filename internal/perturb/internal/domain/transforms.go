package domain

type transform func(l *Layer, a *applied, recs []Delivered, atNS int64) []Delivered

var transforms = map[string]transform{
	DuplicateBurst: (*Layer).duplicateRecords,
	IDReuse:        (*Layer).reuseIdentity,
	Storm:          (*Layer).multiplyRecords,
	Reorder:        (*Layer).reorderRecords,
	DelayTail:      (*Layer).delayRecords,
	GrossBackfill:  (*Layer).withholdBackfill,
	Drop:           (*Layer).dropRecords,
	ProducerFlap:   (*Layer).withholdProducerRecords,
	ClockSkew:      (*Layer).skewClock,
	NonMonotonic:   (*Layer).rewriteNonMonotonicTime,
	TimeEncoding:   (*Layer).rewriteTimeEncoding,
	PrecisionEdge:  (*Layer).truncateTimePrecision,
	OutOfEnum:      (*Layer).mangleEnum,
	OutOfRange:     (*Layer).mangleRange,
	UnitMismatch:   (*Layer).mangleUnit,
	NaNInf:         (*Layer).mangleNumericValue,
	Oversize:       (*Layer).enlargePayload,
	Malformed:      (*Layer).markMalformed,
	InjectionProbe: (*Layer).injectTextProbe,
}

func (l *Layer) applyOne(a *applied, recs []Delivered, atNS int64) []Delivered {
	if apply, ok := transforms[a.Name]; ok {
		return apply(l, a, recs, atNS)
	}
	return recs
}
