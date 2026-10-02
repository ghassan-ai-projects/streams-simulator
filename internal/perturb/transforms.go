package perturb

import ()

func (l *Layer) applyOne(a *Active, recs []Delivered, atNS int64) []Delivered {
	switch a.Name {
	case DuplicateBurst:
		return l.duplicateRecords(a, recs, atNS)
	case IDReuse:
		return l.reuseIdentity(a, recs, atNS)
	case Reorder:
		return l.reorderRecords(a, recs, atNS)
	case DelayTail:
		return l.delayRecords(a, recs, atNS)
	case GrossBackfill:
		return l.withholdBackfill(a, recs, atNS)
	case Drop:
		return l.dropRecords(a, recs, atNS)
	case ClockSkew:
		return l.skewClock(a, recs, atNS)
	case NonMonotonic:
		return l.rewriteNonMonotonicTime(a, recs, atNS)
	case OutOfEnum:
		return l.mangleEnum(a, recs, atNS)
	case OutOfRange:
		return l.mangleRange(a, recs, atNS)
	case UnitMismatch:
		return l.mangleUnit(a, recs, atNS)
	case Oversize:
		return l.enlargePayload(a, recs, atNS)
	case Malformed:
		return l.markMalformed(a, recs, atNS)
	case NaNInf:
		return l.mangleNumericValue(a, recs, atNS)
	case Storm:
		return l.multiplyRecords(a, recs, atNS)
	case ProducerFlap:
		return l.withholdProducerRecords(a, recs, atNS)
	case TimeEncoding:
		return l.rewriteTimeEncoding(a, recs, atNS)
	case PrecisionEdge:
		return l.truncateTimePrecision(a, recs, atNS)
	case InjectionProbe:
		return l.injectTextProbe(a, recs, atNS)
	}
	return recs
}
