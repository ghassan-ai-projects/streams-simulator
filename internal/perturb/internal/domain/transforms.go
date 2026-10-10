package domain

// transform applies one active perturbation to the records of one event.
type transform func(l *Layer, a *applied, recs []Delivered, atNS int64) []Delivered

// transforms maps every catalog name to the function that applies it. Apply
// admits only catalog names, so a name always has an entry.
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
