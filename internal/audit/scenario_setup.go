package audit

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func (p *Panel) prepareWorld(start int64, entities []string, setup []model.SetupCall, perturbations []Perturbation) (*world.World, *perturb.Layer, error) {
	w, err := world.New(p.spec, p.seed, "w-audit", start, world.Options{InitialEntities: entities, ForceEffectorOK: true})
	if err != nil {
		return nil, nil, fmt.Errorf("build: %w", err)
	}
	layer := perturb.New(w.ID, p.seed, p.spec)
	if err := applyAuditPerturbations(layer, perturbations); err != nil {
		return nil, nil, err
	}
	if err := applyAuditSetup(w, setup); err != nil {
		return nil, nil, err
	}
	return w, layer, nil
}

func applyAuditPerturbations(layer *perturb.Layer, perturbations []Perturbation) error {
	for _, item := range perturbations {
		if _, err := layer.Apply(item.Name, item.Params, item.FromNS, item.UntilNS); err != nil {
			return fmt.Errorf("build: perturb %s: %w", item.Name, err)
		}
	}
	return nil
}

func applyAuditSetup(w *world.World, setup []model.SetupCall) error {
	for _, call := range setup {
		if _, err := w.InvokeEffector(call.Effector, call.EntityID, call.CommandID, call.Args, call.AtNS); err != nil {
			return fmt.Errorf("build: %w", err)
		}
	}
	return nil
}

func (capture *auditCapture) bindEmitter(w *world.World, layer *perturb.Layer) {
	w.SetEmitter(func(ev model.SimEvent) {
		t, _ := model.ParseTime(ev.EventTime)
		for _, delivery := range layer.Process(ev, t) {
			capture.deliver(delivery)
		}
	})
}

func injectAuditFaults(w *world.World, entity string, faults map[string]int64) error {
	for fault, at := range faults {
		if fault == "" {
			continue
		}
		if _, err := w.InjectFault(entity, fault, at, nil); err != nil {
			return fmt.Errorf("build: %w", err)
		}
	}
	return nil
}

func auditHorizon(start, duration int64) int64 {
	// The audit is bounded by the scenario window or 24 hours.
	horizon := start + 24*3600*1e9
	if duration > 0 && start+duration < horizon {
		horizon = start + duration
	}
	return horizon
}

func (capture *auditCapture) finishWorld(w *world.World, layer *perturb.Layer, horizon int64) error {
	if _, _, err := w.Advance(horizon); err != nil {
		return fmt.Errorf("build: %w", err)
	}
	for _, delivery := range layer.Flush(horizon) {
		capture.deliver(delivery)
	}
	return nil
}

func prepareAuditCapture(entity string, w *world.World, layer *perturb.Layer) *auditCapture {
	capture := &auditCapture{entityID: entity, log: emissionLog{}, series: map[string][]float64{}, times: map[string][]int64{}}
	capture.bindEmitter(w, layer)
	return capture
}
