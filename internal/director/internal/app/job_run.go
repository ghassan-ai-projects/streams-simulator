package app

import (
	"fmt"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func (d *Director) runJob(j *job) {
	defer close(j.done)
	defer j.cancel()
	outcome, err := j.run.ExecuteScenarioWith(j.ctx, &j.scenario, j.hooks())
	j.closeImmediate()
	if err != nil {
		d.failJob(j, fmt.Errorf("job failed: %w", err))
		return
	}
	d.concludeJob(j, outcome)
}

func (d *Director) concludeJob(j *job, outcome run.ScenarioOutcome) {
	reason, state := model.TerminationCompleted, rules.JobCompleted
	if outcome.Stopped {
		reason, state = model.TerminationStoppedByOperator, rules.JobStopped
	}
	termination := model.Termination{Reason: reason, AtNS: j.run.Status().ClockNS}
	if err := j.run.SetTermination(termination); err != nil {
		d.failJob(j, fmt.Errorf("job failed: %w", err))
		return
	}
	d.endJob(j, state)
}

func (d *Director) failJob(j *job, cause error) {
	j.run.Abandon(cause.Error())
	if _, err := d.EndRun(j.worldID); err != nil {
		cause = fmt.Errorf("%w; ending the run: %w", cause, err)
	}
	j.finish(rules.JobFailed, cause.Error())
}

func (d *Director) endJob(j *job, state rules.JobState) {
	if _, err := d.EndRun(j.worldID); err != nil {
		j.finish(rules.JobFailed, fmt.Sprintf("job failed: ending the run: %v", err))
		return
	}
	j.finish(state, "")
}
