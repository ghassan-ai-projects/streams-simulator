package score

import (
	"context"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func TestDroppedDetectionRequiresOneDetectionPerDrop(t *testing.T) {
	spec, a := testBase(t)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := run.New(context.Background(), run.Config{
		Domain: spec, Adapter: a, Seed: 22, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.ApplyPerturb("drop", map[string]any{"rate": 1.0}, 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(context.Background(), start+3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	if err := r.SubmitVerdict(&model.Verdict{
		SchemaVersion: "0.1", RunID: r.ID, Consumer: model.ConsumerInfo{Name: "test", Version: "1"},
		Detections: []model.Detection{{EntityID: pondIDs[0], DetectedAt: model.FormatTime(start + 1*1e9)}},
	}); err != nil {
		t.Fatal(err)
	}
	m := consumer(r, &model.GroundTruthRecord{EntityID: pondIDs[0]})
	if m.DroppedEventDetection {
		t.Fatal("one detection must not satisfy every dropped delivery")
	}
}

func TestActionFidelityChecksEffectorAndEntity(t *testing.T) {
	spec, a := testBase(t)
	start := model.DefaultStartTimeNS
	r, err := run.New(context.Background(), run.Config{
		Domain: spec, Adapter: a, Seed: 23, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, ForceFailureMode: "ok",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.InvokeEffector("start_aerator", pondIDs[0], "cmd", map[string]any{"pond_id": pondIDs[0], "level": 1.0}, start); err != nil {
		t.Fatal(err)
	}
	if err := r.SubmitVerdict(&model.Verdict{
		SchemaVersion: "0.1", RunID: r.ID, Consumer: model.ConsumerInfo{Name: "test", Version: "1"},
		Actions: []model.Action{{CommandID: "cmd", Effector: "halt_feeding", EntityID: pondIDs[1], IssuedAt: model.FormatTime(start), OutcomeBelieved: model.BelievedSucceeded}},
	}); err != nil {
		t.Fatal(err)
	}
	if consumer(r, &model.GroundTruthRecord{}).ActionFidelity {
		t.Fatal("wrong effector/entity must not receive action credit")
	}
}
