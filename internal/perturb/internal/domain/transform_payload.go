package domain

import (
	"strconv"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (l *Layer) mangleEnum(a *applied, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered || a.rng.Float64() >= paramFloat(a.Params, "rate", 0.01) {
			return []Delivered{r}
		}
		ch := l.domain.Channel(r.Event.Channel)
		if ch == nil || ch.ValueType != "string" {
			return []Delivered{r}
		}
		r.Event.Value = "undeclared_" + strconv.Itoa(a.rng.Intn(1000))
		r.Reason = model.DeliveryMangled
		return []Delivered{r}
	})
}

func (l *Layer) mangleRange(a *applied, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered || a.rng.Float64() >= paramFloat(a.Params, "rate", 0.01) {
			return []Delivered{r}
		}
		ch := l.domain.Channel(r.Event.Channel)
		if ch == nil || ch.Range == nil {
			return []Delivered{r}
		}
		mag := paramFloat(a.Params, "magnitude", 10)
		r.Event.Value = ch.Range.Max + mag
		r.Reason = model.DeliveryMangled
		return []Delivered{r}
	})
}

func (l *Layer) mangleUnit(a *applied, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered || r.Event.Unit == "" {
			return []Delivered{r}
		}
		r.Event.Unit = "err"
		r.Reason = model.DeliveryMangled
		return []Delivered{r}
	})
}

func (l *Layer) enlargePayload(a *applied, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(record Delivered) []Delivered { return []Delivered{enlargedRecord(a, record)} })
}

func (l *Layer) markMalformed(a *applied, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered {
			return []Delivered{r}
		}
		r.Malformed = true
		r.Reason = model.DeliveryMangled
		return []Delivered{r}
	})
}

func (l *Layer) mangleNumericValue(a *applied, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered {
			return []Delivered{r}
		}
		choices := []string{"NaN", "Infinity", "-Infinity"}
		r.Event.Value = choices[a.rng.Intn(len(choices))]
		r.Reason = model.DeliveryMangled
		return []Delivered{r}
	})
}

func (l *Layer) injectTextProbe(a *applied, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(record Delivered) []Delivered { return []Delivered{l.probedRecord(a, record)} })
}

func enlargedRecord(a *applied, record Delivered) Delivered {
	if !record.Delivered {
		return record
	}
	value, changed := enlargedValue(record.Event.Value, paramInt(a.Params, "bytes", 4096))
	if changed {
		record.Event.Value = value
		record.Reason = model.DeliveryMangled
	}
	return record
}

func enlargedValue(value any, limit int) (any, bool) {
	if text, ok := value.(string); ok {
		if len(text) >= limit {
			return value, false
		}
		return text + strings.Repeat("x", limit-len(text)), true
	}
	return strings.Repeat("x", limit), true
}

func (l *Layer) probedRecord(a *applied, record Delivered) Delivered {
	if !record.Delivered {
		return record
	}
	channel := l.domain.Channel(record.Event.Channel)
	if channel == nil || !channel.AttackerControlled {
		return record
	}
	applyProbePayload(a, &record)
	record.Reason = model.DeliveryRewritten
	return record
}

func applyProbePayload(a *applied, record *Delivered) {
	if payloads, ok := a.Params["payloads"].([]any); ok && len(payloads) > 0 {
		index := a.rng.Intn(len(payloads))
		if text, ok := payloads[index].(string); ok {
			record.Event.Value = text
		}
	}
}
