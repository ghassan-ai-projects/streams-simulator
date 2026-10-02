package perturb

import (
	"strconv"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (l *Layer) mangleEnum(a *Active, recs []Delivered, atNS int64) []Delivered {
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

func (l *Layer) mangleRange(a *Active, recs []Delivered, atNS int64) []Delivered {
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

func (l *Layer) mangleUnit(a *Active, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered || r.Event.Unit == "" {
			return []Delivered{r}
		}
		r.Event.Unit = "err"
		r.Reason = model.DeliveryMangled
		return []Delivered{r}
	})
}

func (l *Layer) enlargePayload(a *Active, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered {
			return []Delivered{r}
		}
		limit := paramInt(a.Params, "bytes", 4096)
		switch v := r.Event.Value.(type) {
		case string:
			if len(v) < limit {
				r.Event.Value = v + strings.Repeat("x", limit-len(v))
				r.Reason = model.DeliveryMangled
			}
		default:
			r.Event.Value = strings.Repeat("x", limit)
			r.Reason = model.DeliveryMangled
		}
		return []Delivered{r}
	})
}

func (l *Layer) markMalformed(a *Active, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered {
			return []Delivered{r}
		}
		r.Malformed = true
		r.Reason = model.DeliveryMangled
		return []Delivered{r}
	})
}

func (l *Layer) mangleNumericValue(a *Active, recs []Delivered, atNS int64) []Delivered {
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

func (l *Layer) injectTextProbe(a *Active, recs []Delivered, atNS int64) []Delivered {
	return mapRecs(recs, func(r Delivered) []Delivered {
		if !r.Delivered {
			return []Delivered{r}
		}
		ch := l.domain.Channel(r.Event.Channel)
		if ch == nil || !ch.AttackerControlled {
			return []Delivered{r}
		}
		if payloads, ok := a.Params["payloads"].([]any); ok && len(payloads) > 0 {
			idx := a.rng.Intn(len(payloads))
			if s, ok := payloads[idx].(string); ok {
				r.Event.Value = s
			}
		}
		r.Reason = model.DeliveryRewritten
		return []Delivered{r}
	})
}
