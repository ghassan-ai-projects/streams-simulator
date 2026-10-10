package perturb

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/perturb/internal/domain"
)

// Apply activates a perturbation. fromNS/untilNS bound its activity window
// (0 = from now / no end). It returns the id of the applied perturbation.
func (l *Layer) Apply(name string, params map[string]any, fromNS, untilNS int64) (string, error) {
	return l.layer.Apply(name, params, fromNS, untilNS)
}

// Clear deactivates a perturbation by id.
func (l *Layer) Clear(id string) error {
	return l.layer.Clear(id)
}

// Process applies every active perturbation to one native event at world
// time atNS, returning the delivered records. An event may produce zero
// (drop), one, or several records (duplicate, storm).
func (l *Layer) Process(ev model.SimEvent, atNS int64) []Delivered {
	return l.layer.Process(ev, atNS)
}

// Flush returns records held by windowing perturbations (reorder windows,
// producer flaps, gross backfill) at the end of an advance. Callers must
// process them exactly once, at the boundary where the perturbation's window
// closes.
func (l *Layer) Flush(atNS int64) []Delivered {
	return l.layer.Flush(atNS)
}

// Names returns the perturbation catalog in a stable order. The slice is a
// copy: changing it does not change which perturbations Apply admits.
func Names() []string {
	return layer.CatalogNames()
}
