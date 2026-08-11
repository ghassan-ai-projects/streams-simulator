// Package perturb implements the delivery perturbation layer, which sits
// between the world and the adapter: the world produces what physically
// happened, the perturbation layer produces what the observer got. It
// changes nothing physical and tests a consumer's ingest, time handling and
// deduplication. The delivery ledger records what each perturbation did, so
// a scenario the consumer never saw is scored as a transport miss, not a
// reasoning miss.
package perturb

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

// Perturbation names (the catalog in TECHNICAL_DESIGN §5.2).
const (
	DuplicateBurst = "duplicate_burst"
	IDReuse        = "id_reuse"
	Reorder        = "reorder"
	DelayTail      = "delay_tail"
	GrossBackfill  = "gross_backfill"
	Drop           = "drop"
	ClockSkew      = "clock_skew"
	NonMonotonic   = "non_monotonic"
	OutOfEnum      = "out_of_enum"
	OutOfRange     = "out_of_range"
	UnitMismatch   = "unit_mismatch"
	Oversize       = "oversize"
	Malformed      = "malformed"
	NaNInf         = "nan_inf"
	Storm          = "storm"
	ProducerFlap   = "producer_flap"
	TimeEncoding   = "time_encoding"
	PrecisionEdge  = "precision_edge"
	InjectionProbe = "injection_probe"
)

// Names is the full catalog in a stable order.
var Names = []string{
	DuplicateBurst, IDReuse, Reorder, DelayTail, GrossBackfill, Drop,
	ClockSkew, NonMonotonic, OutOfEnum, OutOfRange, UnitMismatch, Oversize,
	Malformed, NaNInf, Storm, ProducerFlap, TimeEncoding, PrecisionEdge,
	InjectionProbe,
}

// Delivered is one delivered record: the (possibly modified) event plus its
// ledger metadata. Drop yields Delivered=false; duplicate yields extra
// records with the same seq.
type Delivered struct {
	DeliveryID uint64
	Event      model.SimEvent
	Reason     string
	Delivered  bool
	Malformed  bool
}

// Active is one applied perturbation.
type Active struct {
	ID      string
	Name    string
	Params  map[string]any
	FromNS  int64
	UntilNS int64
	window  *reorderWindow
	rng     *randutil.SplitMix64
	// producer-flap buffer
	buffer []Delivered
}

// Layer applies active perturbations to a native event stream.
type Layer struct {
	worldID string
	seed    uint64
	domain  *domain.Compiled
	active  map[string]*Active
	order   []string
	seq     int
	nextID  uint64
	pending []Delivered
}

// New builds a perturbation layer for a world. The domain spec is needed
// only for contract knowledge (declared enums, ranges, units, attacker-
// controlled channels); consumer knowledge never enters here.
func New(worldID string, seed uint64, spec *domain.Compiled) *Layer {
	return &Layer{
		worldID: worldID,
		seed:    seed,
		domain:  spec,
		active:  map[string]*Active{},
	}
}

// Apply activates a perturbation. fromNS/untilNS bound its activity window
// (0 = from now / no end).
func (l *Layer) Apply(name string, params map[string]any, fromNS, untilNS int64) (string, error) {
	if !isName(name) {
		return "", fmt.Errorf("perturb: unknown perturbation %q", name)
	}
	if err := validateParams(name, params); err != nil {
		return "", err
	}
	id := name + "-" + strconv.Itoa(l.seq)
	l.seq++
	a := &Active{
		ID:      id,
		Name:    name,
		Params:  params,
		FromNS:  fromNS,
		UntilNS: untilNS,
		rng:     randutil.Substream(l.seed, l.worldID+"/perturb/"+id),
	}
	if name == Reorder {
		disp := int(paramInt(params, "max_displacement", 2))
		a.window = newReorderWindow(disp)
	}
	l.active[id] = a
	l.order = append(l.order, id)
	return id, nil
}

// Clear deactivates a perturbation.
func (l *Layer) Clear(id string) error {
	a, ok := l.active[id]
	if !ok {
		return fmt.Errorf("perturb: unknown perturbation id %q", id)
	}
	for _, r := range a.buffer {
		r.Delivered = false
		r.Reason = model.DeliveryDroppedByPerturb
		l.pending = append(l.pending, r)
	}
	if a.window != nil {
		for _, r := range a.window.flush(0) {
			r.Delivered = false
			r.Reason = model.DeliveryDroppedByPerturb
			l.pending = append(l.pending, r)
		}
	}
	delete(l.active, id)
	return nil
}

// ActiveIDs lists the active perturbation ids in application order.
func (l *Layer) ActiveIDs() []string {
	var out []string
	for _, id := range l.order {
		if _, ok := l.active[id]; ok {
			out = append(out, id)
		}
	}
	return out
}

// Process applies every active perturbation to one native event at world
// time atNS, returning the delivered records. An event may produce zero
// (drop), one, or several records (duplicate, storm).
func (l *Layer) Process(ev model.SimEvent, atNS int64) []Delivered {
	// Unchanged by default; each active perturbation transforms the list.
	l.nextID++
	recs := []Delivered{{DeliveryID: l.nextID, Event: ev, Reason: model.DeliveryOK, Delivered: true}}
	for _, id := range l.ActiveIDs() {
		a := l.active[id]
		if atNS < a.FromNS || (a.UntilNS > 0 && atNS >= a.UntilNS) {
			continue
		}
		recs = l.applyOne(a, recs, atNS)
	}
	// Post-pass: flaps and reorder windows may hold records back.
	return recs
}

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

// Flush returns records held by windowing perturbations (reorder windows,
// producer flaps, gross backfill) at the end of an advance. Callers must
// process them exactly once, at the boundary where the perturbation's
// window closes.
func (l *Layer) Flush(atNS int64) []Delivered {
	var out []Delivered
	for _, id := range l.ActiveIDs() {
		a := l.active[id]
		switch a.Name {
		case Reorder:
			flushed := a.window.flush(atNS)
			for i := range flushed {
				if flushed[i].Delivered && flushed[i].Reason == model.DeliveryOK {
					flushed[i].Reason = model.DeliveryReordered
				}
			}
			out = append(out, flushed...)
		case ProducerFlap:
			// Birth burst: republish every held event at one observed time,
			// with event times spread across the outage.
			if len(a.buffer) > 0 && a.UntilNS > 0 && atNS >= a.UntilNS {
				recovery := atNS
				for i := range a.buffer {
					r := a.buffer[i]
					ev := r.Event
					ev.Birth = true
					ev.ObservedTime = model.FormatTime(recovery)
					r.Event = ev
					r.Reason = model.DeliveryDelayed
					out = append(out, r)
				}
				a.buffer = nil
			}
		case GrossBackfill:
			if len(a.buffer) > 0 && a.UntilNS > 0 && atNS >= a.UntilNS {
				recovery := atNS
				for i := range a.buffer {
					r := a.buffer[i]
					ev := r.Event
					ev.ObservedTime = model.FormatTime(recovery)
					r.Event = ev
					r.Reason = model.DeliveryDelayed
					out = append(out, r)
				}
				a.buffer = nil
			}
		}
	}
	if len(l.pending) > 0 {
		out = append(l.pending, out...)
		l.pending = nil
	}
	return out
}

func mapRecs(recs []Delivered, f func(Delivered) []Delivered) []Delivered {
	var out []Delivered
	for _, r := range recs {
		out = append(out, f(r)...)
	}
	return out
}

func isName(name string) bool {
	for _, n := range Names {
		if n == name {
			return true
		}
	}
	return false
}

func paramFloat(params map[string]any, key string, def float64) float64 {
	if params == nil {
		return def
	}
	switch v := params[key].(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	case int:
		return float64(v)
	case json.Number:
		f, err := v.Float64()
		if err == nil {
			return f
		}
	}
	return def
}

func paramInt(params map[string]any, key string, def int) int {
	return int(paramFloat(params, key, float64(def)))
}

func paramStr(params map[string]any, key, def string) string {
	if params == nil {
		return def
	}
	if s, ok := params[key].(string); ok {
		return s
	}
	return def
}

// paramCheck validates one declared parameter value. Perturbation parameters
// are fail-closed: a misspelled or unknown key is an error, not a silently
// ignored default, and a declared key must have the declared type and range.
type paramCheck func(v any) error

func numInRange[T int64 | float64](lo, hi T) paramCheck {
	return func(v any) error {
		f, ok := asFloat(v)
		if !ok {
			return fmt.Errorf("expected a number, got %T", v)
		}
		if f < float64(lo) || f > float64(hi) {
			return fmt.Errorf("must be in [%v,%v], got %v", lo, hi, f)
		}
		return nil
	}
}

func strEnum(values ...string) paramCheck {
	return func(v any) error {
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("expected a string, got %T", v)
		}
		for _, want := range values {
			if s == want {
				return nil
			}
		}
		return fmt.Errorf("must be one of %v, got %q", values, s)
	}
}

// payloadsCheck validates the injection_probe payload list: a non-empty array
// of strings. Payloads are the adversarial text families for the probe.
func payloadsCheck(v any) error {
	arr, ok := v.([]any)
	if !ok {
		return fmt.Errorf("expected an array of strings, got %T", v)
	}
	if len(arr) == 0 {
		return fmt.Errorf("must contain at least one payload")
	}
	for i, p := range arr {
		if _, ok := p.(string); !ok {
			return fmt.Errorf("payload %d: expected a string, got %T", i, p)
		}
	}
	return nil
}

// paramRules maps each perturbation to its declared parameter keys. A
// perturbation with no entry accepts no parameters. Every key here is read
// by applyOne; anything else is rejected so a misspelled option can never be
// silently approximated by a default.
var paramRules = map[string]map[string]paramCheck{
	Drop:           {"rate": numInRange(0.0, 1.0)},
	DuplicateBurst: {"rate": numInRange(0.0, 1.0)},
	IDReuse:        {"rate": numInRange(0.0, 1.0)},
	OutOfEnum:      {"rate": numInRange(0.0, 1.0)},
	OutOfRange: {
		"rate":      numInRange(0.0, 1.0),
		"magnitude": numInRange(-1e9, 1e9),
	},
	Reorder: {"max_displacement": numInRange(0, int64(1<<30))},
	DelayTail: {
		"mean_s":  numInRange(0.0, 1e9),
		"sigma_s": numInRange(0.0, 1e9),
	},
	ClockSkew: {
		"offset_s": numInRange(0.0, 1e9),
		"sign":     strEnum("positive", "negative"),
	},
	Oversize:       {"bytes": numInRange(1, int64(1<<62))},
	Storm:          {"multiplier": numInRange(1, int64(1<<20))},
	InjectionProbe: {"payloads": payloadsCheck},
}

func validateParams(name string, params map[string]any) error {
	if params == nil {
		return nil
	}
	rules, declared := paramRules[name]
	for k := range params {
		if !declared {
			return fmt.Errorf("perturb: %s accepts no parameters (got %q)", name, k)
		}
		check, ok := rules[k]
		if !ok {
			return fmt.Errorf("perturb: %s has no parameter %q", name, k)
		}
		if err := check(params[k]); err != nil {
			return fmt.Errorf("perturb: %s %s: %w", name, k, err)
		}
	}
	return nil
}

// asFloat coerces a JSON number to float64 without accepting non-numeric
// types. JSON numbers arrive as float64 or int64; Go callers may pass int;
// strings are never numeric.
func asFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int64:
		return float64(x), true
	case int:
		return float64(x), true
	}
	return 0, false
}
