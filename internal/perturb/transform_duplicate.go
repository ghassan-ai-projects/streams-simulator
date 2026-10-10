package perturb

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (l *Layer) duplicateRecords(a *applied, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered || a.rng.Float64() >= paramFloat(a.Params, "rate", 0.02) {
			return []Delivered{r}
		}
		dup := l.duplicateDelivery(r)
		return []Delivered{r, dup}
	})
}

func (l *Layer) reuseIdentity(a *applied, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(record Delivered) []Delivered { return l.reusedRecord(a, record) })
}

func (l *Layer) dropRecords(a *applied, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if r.Delivered && a.rng.Float64() < paramFloat(a.Params, "rate", 0.01) {
			return []Delivered{{DeliveryID: r.DeliveryID, Event: r.Event, Reason: model.DeliveryDroppedByPerturb, Delivered: false}}
		}
		return []Delivered{r}
	})
}

func (l *Layer) multiplyRecords(a *applied, recs []Delivered, atNS int64) []Delivered {
	multiplier := paramInt(a.Params, "multiplier", 3)
	return mapRecs(recs, func(record Delivered) []Delivered { return l.multipliedRecord(record, multiplier) })
}

func (l *Layer) reusedRecord(a *applied, record Delivered) []Delivered {
	if !record.Delivered || a.rng.Float64() >= paramFloat(a.Params, "rate", 0.01) {
		return []Delivered{record}
	}
	reuse := l.duplicateDelivery(record)
	reuse.Event.Value = shiftedIdentityValue(record.Event.Value)
	return []Delivered{record, reuse}
}

func (l *Layer) duplicateDelivery(record Delivered) Delivered {
	duplicate := record
	l.nextID++
	duplicate.DeliveryID = l.nextID
	duplicate.Reason = model.DeliveryDuplicated
	return duplicate
}

func shiftedIdentityValue(value any) any {
	// Same event identity with a deliberately different payload.
	switch value := value.(type) {
	case float64:
		return value + 100
	case string:
		return value + " (reused)"
	}
	return value
}

func (l *Layer) multipliedRecord(record Delivered, multiplier int) []Delivered {
	if !record.Delivered {
		return []Delivered{record}
	}
	out := []Delivered{record}
	for i := 1; i < multiplier; i++ {
		out = append(out, l.duplicateDelivery(record))
	}
	return out
}
