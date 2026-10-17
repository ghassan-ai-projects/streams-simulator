// Package app implements the director's use cases: the registry of worlds and
// runs, the world commands (advance, fault, perturb, env), the run lifecycle
// (begin, end, seal and reveal truth, score, verify, audit) and the operator
// operations a capability token resolves to. The surface that calls it owns
// protocol framing and argument decoding; the simulator's rules live in the
// modules it composes.
package app

import (
	"context"
	"sync"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
)

type WorldRecord struct {
	Run       *run.Run
	Token     string
	Nameplate *rules.Nameplate
	Operator  *rules.OperatorView
	Started   bool
	RunEnded  bool

	DirectorCommands int
}

type Director struct {
	mu               sync.Mutex
	domains          *domain.Catalog
	catalogMu        sync.RWMutex
	library          *component.Library
	check            func(*model.DomainSpec) error
	Adapters         map[string]*model.Adapter
	Worlds           map[string]*WorldRecord
	byRun            map[string]string
	byToken          map[string]*WorldRecord
	OperatorEndpoint string
	Truth            *truth.Store
	OutDir           string
	ctx              context.Context
	seq              int
}

func NewDirector(ctx context.Context, cat *domain.Catalog, adapters map[string]*model.Adapter, outDir string) *Director {
	d := &Director{
		domains: cat, Adapters: adapters, Worlds: map[string]*WorldRecord{},
		byRun: map[string]string{}, byToken: map[string]*WorldRecord{},
		OutDir: outDir, ctx: ctx,
	}
	d.Truth = truth.NewStore(d.runIsOpen)
	return d
}

func (d *Director) ResolveOperator(token string) (*rules.OperatorView, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	rec := d.byToken[token]
	if rec == nil {
		return nil, rules.ToolErrorf(rules.CodeCapabilityDenied, "capability token required")
	}
	return rec.Operator, nil
}

func (d *Director) SetOperatorEndpoint(endpoint string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.OperatorEndpoint = endpoint
}

func (d *Director) World(worldID string) *WorldRecord {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.Worlds[worldID]
}

func (d *Director) RunOf(worldID string) *run.Run {
	if w := d.World(worldID); w != nil {
		return w.Run
	}
	return nil
}

func (d *Director) openRun(w *WorldRecord, worldID string) error {
	if err := d.requireSealedRun(w.Run.ID); err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if w.Started {
		return rules.ToolErrorf(rules.CodeDomainInvalid, "a run is already open for %q", worldID)
	}
	w.Started = true
	return nil
}

func (d *Director) runLifecycle(w *WorldRecord) (started, ended bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return w.Started, w.RunEnded
}

func (d *Director) markRunEnded(w *WorldRecord) {
	if !w.Run.Finished() {
		return
	}
	d.mu.Lock()
	w.RunEnded = true
	d.mu.Unlock()
}

func (d *Director) runIsOpen(runID string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	worldID, ok := d.byRun[runID]
	if !ok {
		return true
	}
	w := d.Worlds[worldID]
	return w != nil && !w.RunEnded
}
