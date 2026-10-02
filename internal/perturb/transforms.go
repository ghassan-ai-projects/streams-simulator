package perturb

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"strconv"
	"strings"
)

func (l *Layer) applyOne(a *Active, recs []Delivered, atNS int64) []Delivered {
	switch a.Name {
	case DuplicateBurst:
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
	case IDReuse:
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
	case Reorder:
		out := a.window.push(recs, atNS)
		for i := range out {
			if out[i].Delivered && out[i].Reason == model.DeliveryOK {
				out[i].Reason = model.DeliveryReordered
			}
		}
		return out
	case DelayTail:
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
	case GrossBackfill:
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
	case Drop:
		return mapRecs(recs, func(r Delivered) []Delivered {
			if r.Delivered && a.rng.Float64() < paramFloat(a.Params, "rate", 0.01) {
				return []Delivered{{DeliveryID: r.DeliveryID, Event: r.Event, Reason: model.DeliveryDroppedByPerturb, Delivered: false}}
			}
			return []Delivered{r}
		})
	case ClockSkew:
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
	case NonMonotonic:
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
	case OutOfEnum:
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
	case OutOfRange:
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
	case UnitMismatch:
		return mapRecs(recs, func(r Delivered) []Delivered {
			if !r.Delivered || r.Event.Unit == "" {
				return []Delivered{r}
			}
			r.Event.Unit = "err"
			r.Reason = model.DeliveryMangled
			return []Delivered{r}
		})
	case Oversize:
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
	case Malformed:
		return mapRecs(recs, func(r Delivered) []Delivered {
			if !r.Delivered {
				return []Delivered{r}
			}
			r.Malformed = true
			r.Reason = model.DeliveryMangled
			return []Delivered{r}
		})
	case NaNInf:
		return mapRecs(recs, func(r Delivered) []Delivered {
			if !r.Delivered {
				return []Delivered{r}
			}
			choices := []string{"NaN", "Infinity", "-Infinity"}
			r.Event.Value = choices[a.rng.Intn(len(choices))]
			r.Reason = model.DeliveryMangled
			return []Delivered{r}
		})
	case Storm:
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
	case ProducerFlap:
		return mapRecs(recs, func(r Delivered) []Delivered {
			if !r.Delivered {
				return []Delivered{r}
			}
			// Hold every event during the flap window; Flush republishes.
			a.buffer = append(a.buffer, r)
			return nil
		})
	case TimeEncoding:
		return mapRecs(recs, func(r Delivered) []Delivered {
			if !r.Delivered {
				return []Delivered{r}
			}
			r.Event.ObservedTime = alternateEncoding(r.Event.ObservedTime)
			r.Reason = model.DeliveryRewritten
			return []Delivered{r}
		})
	case PrecisionEdge:
		return mapRecs(recs, func(r Delivered) []Delivered {
			if !r.Delivered {
				return []Delivered{r}
			}
			r.Event.ObservedTime = truncatePrecision(r.Event.ObservedTime)
			r.Reason = model.DeliveryRewritten
			return []Delivered{r}
		})
	case InjectionProbe:
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
	return recs
}
