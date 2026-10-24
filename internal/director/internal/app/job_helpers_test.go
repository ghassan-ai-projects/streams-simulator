package app

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"
	"sync"
	"testing"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const pondEntity = "site-a/pond-1"

type anchor struct {
	elapsedNS int64
	rate      float64
}

type gatedPacer struct {
	asked chan int64
	gate  chan struct{}

	mu      sync.Mutex
	anchors []anchor
}

func newGatedPacer() *gatedPacer {
	return &gatedPacer{asked: make(chan int64, 1024), gate: make(chan struct{}, 1024)}
}

func (p *gatedPacer) Reanchor(simElapsedNS int64, rate float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.anchors = append(p.anchors, anchor{simElapsedNS, rate})
}

func (p *gatedPacer) reanchored() []anchor {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]anchor{}, p.anchors...)
}

func (p *gatedPacer) WaitUntil(ctx context.Context, simElapsedNS int64) error {
	p.asked <- simElapsedNS
	select {
	case <-p.gate:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("gated pacer: %w", ctx.Err())
	}
}

func (p *gatedPacer) release(n int) {
	for range n {
		p.gate <- struct{}{}
	}
}

type freePacer struct{}

func (freePacer) Reanchor(int64, float64) {}

func (freePacer) WaitUntil(ctx context.Context, _ int64) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("free pacer: %w", err)
	}
	return nil
}

func pacedBy(d *Director, pacer Pacer) *Director {
	d.jobs.newPacer = func(float64) Pacer { return pacer }
	return d
}

func shippedDirector(t *testing.T, ctx context.Context) *Director {
	t.Helper()
	d, err := LoadDirector(ctx, testsupport.DomainsDir(), testsupport.AdaptersDir(), t.TempDir(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func pondJob() rules.JobRequest {
	return rules.JobRequest{Domain: "aquaculture-pond", Seed: 5, DurationS: 3600, Rate: 60,
		Faults: []rules.JobFault{{EntityID: pondEntity, Fault: "aerator_failure", AtS: 600}},
		Calls:  []rules.JobCall{{EntityID: pondEntity, Effector: "start_aerator", Args: map[string]any{"pond_id": pondEntity}, AtS: 1800}}}
}

func startJob(t *testing.T, d *Director, req rules.JobRequest) rules.JobStatus {
	t.Helper()
	status, err := d.StartJob(req)
	if err != nil {
		t.Fatal(err)
	}
	return status
}

func awaitJob(t *testing.T, d *Director, worldID string) rules.JobStatus {
	t.Helper()
	j := d.jobs.find(worldID)
	if j == nil {
		t.Fatalf("no job for %s", worldID)
	}
	<-j.done
	status, err := d.JobStatus(worldID)
	if err != nil {
		t.Fatal(err)
	}
	return status
}

func loadRunArtifact(t *testing.T, d *Director, worldID string) *model.RunArtifact {
	t.Helper()
	art, err := run.LoadArtifact(filepath.Join(d.OutDir, worldID, model.FileRunArtifact))
	if err != nil {
		t.Fatal(err)
	}
	return art
}

func commandsJSON(t *testing.T, commands any) string {
	t.Helper()
	raw, err := json.Marshal(commands)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func loggedCommands(art *model.RunArtifact) []model.ScenarioCommand {
	out := make([]model.ScenarioCommand, len(art.CommandLog))
	for i, command := range art.CommandLog {
		args := maps.Clone(command.Args)
		delete(args, "await_consumer")
		out[i] = model.ScenarioCommand{Op: command.Op, Args: args}
	}
	return out
}

func requireVerifies(t *testing.T, d *Director, worldID string) {
	t.Helper()
	res, err := d.VerifyRun(filepath.Join(d.OutDir, worldID, model.FileRunArtifact))
	if err != nil || res["matches"] != true {
		t.Fatalf("verify = %v, %v, want the run to replay to the same trace", res, err)
	}
}

func requireReplaysFromTheRunFile(t *testing.T, d *Director, worldID string) *model.RunArtifact {
	t.Helper()
	path := filepath.Join(d.OutDir, worldID, model.FileRunArtifact)
	res, err := run.ReplayFile(t.Context(), path, run.Directories{Domains: testsupport.DomainsDir(), Adapters: testsupport.AdaptersDir()})
	art := loadRunArtifact(t, d, worldID)
	if err != nil || !res.Matches || res.GotDigest != art.ExpectedTraceDigest {
		t.Fatalf("replay from run.json = %+v, %v, want the recorded trace digest %s", res, err, art.ExpectedTraceDigest)
	}
	return art
}

func contextWithCancel(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithCancel(t.Context())
}
