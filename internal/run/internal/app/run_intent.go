package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/domain"
)

func (r *Run) SetPlan(plan model.RunPlan) error {
	if err := rules.CheckRunPlan(plan); err != nil {
		return err
	}
	return r.recordIntent(func() { r.plan = &plan })
}

func (r *Run) SetTermination(termination model.Termination) error {
	if err := rules.CheckTermination(termination); err != nil {
		return err
	}
	return r.recordIntent(func() { r.termination = &termination })
}

func (r *Run) recordIntent(record func()) error {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	if r.finished {
		return fmt.Errorf("run: intent recorded after the run finished")
	}
	record()
	return nil
}

func (r *Run) attachArtifactIntent(artifact *model.RunArtifact) {
	artifact.Plan, artifact.Termination = r.plan, r.termination
}
