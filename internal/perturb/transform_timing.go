package perturb

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (l *Layer) reorderRecords(a *Active, recs []Delivered, atNS int64) []Delivered {
	out := a.window.push(recs, atNS)
	for i := range out {
		if out[i].Delivered && out[i].Reason == model.DeliveryOK {
			out[i].Reason = model.DeliveryReordered
		}
	}
	return out
}

func (l *Layer) delayRecords(a *Active, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered {
			return []Delivered{r}
		}
		delay := a.rng.Exp(paramFloat(a.Params, "mean_s", 60)) +
			a.rng.Exp(paramFloat(a.Params, "sigma_s", 300))
		if delay <= 0 {
			return []Delivered{r}
		}
		r.Event.ObservedTime = addSeconds(r.Event.ObservedTime, delay)
		r.Reason = model.DeliveryDelayed
		return []Delivered{r}
	})
}

func (l *Layer) withholdBackfill(a *Active, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered {
			return []Delivered{r}
		}
		// Withhold during the gap (fromNS..untilNS); flush handled in
		// Flush().
		if atNS < a.UntilNS && a.UntilNS > 0 {
			a.buffer = append(a.buffer, r)
			return nil
		}
		return []Delivered{r}
	})
}

func (l *Layer) skewClock(a *Active, recs []Delivered, atNS int64) []Delivered {
	offset := paramFloat(a.Params, "offset_s", 300)
	if paramStr(a.Params, "sign", "positive") == "negative" {
		offset = -offset
	}
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered {
			return []Delivered{r}
		}
		r.Event.ObservedTime = addSeconds(r.Event.ObservedTime, offset)
		r.Reason = model.DeliveryRewritten
		return []Delivered{r}
	})
}

func (l *Layer) rewriteNonMonotonicTime(a *Active, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered {
			return []Delivered{r}
		}
		// Force observed_time before event_time: a receipt-order
		// violation an honest consumer must reject.
		et, err1 := model.ParseTime(r.Event.EventTime)
		ot, err2 := model.ParseTime(r.Event.ObservedTime)
		if err1 == nil && err2 == nil && ot > et {
			r.Event.ObservedTime = model.FormatTime(et - 1)
			r.Reason = model.DeliveryRewritten
		}
		return []Delivered{r}
	})
}

func (l *Layer) withholdProducerRecords(a *Active, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered {
			return []Delivered{r}
		}
		// Hold every event during the flap window; Flush republishes.
		a.buffer = append(a.buffer, r)
		return nil
	})
}

func (l *Layer) rewriteTimeEncoding(a *Active, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered {
			return []Delivered{r}
		}
		r.Event.ObservedTime = alternateEncoding(r.Event.ObservedTime)
		r.Reason = model.DeliveryRewritten
		return []Delivered{r}
	})
}

func (l *Layer) truncateTimePrecision(a *Active, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered {
			return []Delivered{r}
		}
		r.Event.ObservedTime = truncatePrecision(r.Event.ObservedTime)
		r.Reason = model.DeliveryRewritten
		return []Delivered{r}
	})
}
