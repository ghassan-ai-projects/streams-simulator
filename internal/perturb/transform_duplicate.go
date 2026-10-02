package perturb

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (l *Layer) duplicateRecords(a *Active, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered || a.rng.Float64() >= paramFloat(a.Params, "rate", 0.02) {
			return []Delivered{r}
		}
		dup := r
		l.nextID++
		dup.DeliveryID = l.nextID
		dup.Reason = model.DeliveryDuplicated
		return []Delivered{r, dup}
	})
}

func (l *Layer) reuseIdentity(a *Active, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered || a.rng.Float64() >= paramFloat(a.Params, "rate", 0.01) {
			return []Delivered{r}
		}
		reuse := r
		l.nextID++
		reuse.DeliveryID = l.nextID
		reuse.Reason = model.DeliveryDuplicated
		// Same identity, different payload: shift the value.
		switch v := r.Event.Value.(type) {
		case float64:
			reuse.Event.Value = v + 100
		case string:
			reuse.Event.Value = v + " (reused)"
		}
		return []Delivered{r, reuse}
	})
}

func (l *Layer) dropRecords(a *Active, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if r.Delivered && a.rng.Float64() < paramFloat(a.Params, "rate", 0.01) {
			return []Delivered{{DeliveryID: r.DeliveryID, Event: r.Event, Reason: model.DeliveryDroppedByPerturb, Delivered: false}}
		}
		return []Delivered{r}
	})
}

func (l *Layer) multiplyRecords(a *Active, recs []Delivered, atNS int64) []Delivered {
	mult := paramInt(a.Params, "multiplier", 3)
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered {
			return []Delivered{r}
		}
		out := []Delivered{r}
		for i := 1; i < mult; i++ {
			dup := r
			l.nextID++
			dup.DeliveryID = l.nextID
			dup.Reason = model.DeliveryDuplicated
			out = append(out, dup)
		}
		return out
	})
}
