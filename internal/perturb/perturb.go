// Package perturb implements the delivery perturbation layer, which sits
// between the world and the adapter: the world produces what physically
// happened, the perturbation layer produces what the observer got. It
// changes nothing physical and tests a consumer's ingest, time handling and
// deduplication. The delivery ledger records what each perturbation did, so
// a scenario the consumer never saw is scored as a transport miss, not a
// reasoning miss.
package perturb

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
	"strconv"
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
