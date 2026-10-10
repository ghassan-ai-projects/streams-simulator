package app

import (
	"context"
	"sync"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
)

// WorldRecord is one created world under the director's control.
type WorldRecord struct {
	Run       *run.Run
	Token     string
	Nameplate *Nameplate
	Operator  *OperatorView
	Started   bool
	RunEnded  bool
}

// Director holds the director-role state: the catalog, the installed
// adapters, the world registry, and the truth store.
type Director struct {
	mu               sync.Mutex
	Catalog          *domain.Catalog
	Adapters         map[string]*model.Adapter
	Worlds           map[string]*WorldRecord
	byRun            map[string]string // run id -> world id
	byToken          map[string]*WorldRecord
	OperatorEndpoint string // operator HTTP endpoint served by this process (CLI)
	Truth            *truth.Store
	OutDir           string
	ctx              context.Context
	seq              int
}

// NewDirector builds the director with its registries.
func NewDirector(ctx context.Context, cat *domain.Catalog, adapters map[string]*model.Adapter, outDir string) *Director {
	d := &Director{
		Catalog: cat, Adapters: adapters, Worlds: map[string]*WorldRecord{},
		byRun: map[string]string{}, byToken: map[string]*WorldRecord{},
		OutDir: outDir, ctx: ctx,
	}
	d.Truth = truth.NewStore(d.runIsOpen)
	return d
}

// ResolveOperator returns the operator view for a capability token. This is
// what lets one operator endpoint serve every world: the token, not the
// connection, names the world.
func (d *Director) ResolveOperator(token string) (*OperatorView, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	rec := d.byToken[token]
	if rec == nil {
		return nil, errTool(CodeCapabilityDenied, "capability token required")
	}
	return rec.Operator, nil
}

// SetOperatorEndpoint records the operator HTTP endpoint this process serves
// (CLI wiring); sim.world.create includes it in its response.
func (d *Director) SetOperatorEndpoint(endpoint string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.OperatorEndpoint = endpoint
}

// World returns the record for a world id, or nil.
func (d *Director) World(worldID string) *WorldRecord {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.Worlds[worldID]
}

// openRun marks the world's run opened once its truth is sealed. The check
// and the mark are one step, so two callers cannot both open it.
func (d *Director) openRun(w *WorldRecord, worldID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if w.Started {
		return errTool(CodeDomainInvalid, "a run is already open for %q", worldID)
	}
	if err := d.requireSealedRun(w.Run.ID); err != nil {
		return err
	}
	w.Started = true
	return nil
}

// runLifecycle reports whether the world's run has been opened and ended.
func (d *Director) runLifecycle(w *WorldRecord) (started, ended bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return w.Started, w.RunEnded
}

// markRunEnded records that the world's run is over, whether or not it
// ended cleanly.
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
		return true // a run the director does not know is never treated as closed
	}
	w := d.Worlds[worldID]
	return w != nil && !w.RunEnded
}
