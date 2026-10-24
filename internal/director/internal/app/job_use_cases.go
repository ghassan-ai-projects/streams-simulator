package app

import (
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/runner"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

type jobPlan struct {
	world    WorldRequest
	scenario model.Scenario
}

func (d *Director) StartJob(req rules.JobRequest) (rules.JobStatus, error) {
	if err := rules.CheckJob(req); err != nil {
		return rules.JobStatus{}, err
	}
	d.jobs.startMu.Lock()
	defer d.jobs.startMu.Unlock()
	j, err := d.prepareJob(req)
	if err != nil {
		return rules.JobStatus{}, err
	}
	d.jobs.add(j)
	runner.Go(func() { d.runJob(j) })
	return j.snapshot(), nil
}

func (d *Director) prepareJob(req rules.JobRequest) (*job, error) {
	if err := d.jobs.refuseWhenActive(); err != nil {
		return nil, err
	}
	plan, err := d.planJob(req)
	if err != nil {
		return nil, err
	}
	return d.openJob(req, plan)
}

func (d *Director) planJob(req rules.JobRequest) (jobPlan, error) {
	world := WorldRequest{Domain: req.Domain, Seed: req.Seed, Explore: req.Explore}
	cfg, _, err := d.worldConfig(world)
	if err != nil {
		return jobPlan{}, err
	}
	scenario := rules.PlanScenario(req, worldAdapterID(world), cfg.StartTimeNS)
	if err := run.CheckScenario(d.ctx, cfg, &scenario); err != nil {
		return jobPlan{}, rules.ValidationToolError(err, rules.CodeInvalidArgs)
	}
	return jobPlan{world: world, scenario: scenario}, nil
}

func (d *Director) openJob(req rules.JobRequest, plan jobPlan) (*job, error) {
	created, err := d.CreateWorld(plan.world)
	if err != nil {
		return nil, err
	}
	worldID, _ := created["world_id"].(string)
	r := d.RunOf(worldID)
	status := newJobStatus(worldID, r.ID, req, plan.scenario.World.StartTimeNS)
	if err := r.SetPlan(model.RunPlan{RequestedEndNS: status.RequestedEndNS, Rate: req.Rate}); err != nil {
		_, _ = d.DestroyWorld(worldID)
		return nil, rules.ToolErrorf(rules.CodeDomainInvalid, "%v", err)
	}
	return newJob(d.ctx, r, status, plan.scenario, d.jobs.newPacer(req.Rate)), nil
}

func newJobStatus(worldID, runID string, req rules.JobRequest, startNS int64) rules.JobStatus {
	return rules.JobStatus{WorldID: worldID, RunID: runID, Domain: req.Domain, Seed: req.Seed, State: rules.JobRunning,
		StartNS: startNS, ReachedNS: startNS, RequestedEndNS: startNS + req.DurationS*nsPerSecond,
		Rate: req.Rate, Explore: req.Explore, Calls: []rules.JobCallOutcome{}}
}

func (d *Director) JobStatus(worldID string) (rules.JobStatus, error) {
	j, err := d.jobOf(worldID)
	if err != nil {
		return rules.JobStatus{}, err
	}
	return j.snapshot(), nil
}

func (d *Director) StopJob(worldID string) (rules.JobStatus, error) {
	j, err := d.jobOf(worldID)
	if err != nil {
		return rules.JobStatus{}, err
	}
	return j.requestStop(), nil
}

func (d *Director) ListJobs() []rules.JobStatus {
	jobs := d.jobs.all()
	listed := make([]rules.JobStatus, len(jobs))
	for i, j := range jobs {
		listed[i] = j.snapshot()
	}
	return listed
}

func (d *Director) stopJobAndWait(worldID string) {
	if j := d.jobs.find(worldID); j != nil {
		j.requestStop()
		<-j.done
	}
}
