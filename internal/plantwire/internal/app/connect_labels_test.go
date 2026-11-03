package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func readLabels(t *testing.T, out string) []model.GroundTruthRecord {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(out, model.FileLabels))
	if err != nil {
		t.Fatal(err)
	}
	var records []json.RawMessage
	if err := json.Unmarshal(raw, &records); err != nil {
		t.Fatalf("labels.json is not an array: %v\n%s", err, raw)
	}
	labels := make([]model.GroundTruthRecord, len(records))
	for i, record := range records {
		if err := model.ValidateGroundTruth(record); err != nil || json.Unmarshal(record, &labels[i]) != nil {
			t.Fatalf("labels.json[%d]: %v\n%s", i, err, record)
		}
	}
	return labels
}

func TestSeveralDisturbancesWriteOneSealedLabelEachInPlanOrder(t *testing.T) {
	t.Parallel()
	const plan = `,"disturbances":[{"at_s":0,"instance":"site-a/pump-1","fault":"bearing_wear"},` +
		`{"at_s":600,"instance":"site-a/pump-2","fault":"transient_none","duration_s":900}]`
	session := runStationSession(t, "pump-station", plan, "", nil, 0.05)
	labels := readLabels(t, session.out)
	if len(labels) != 2 {
		t.Fatalf("%d labels, want one per disturbance", len(labels))
	}
	wear, quiet := labels[0], labels[1]
	if wear.ScenarioID != "pump-station/0001" || wear.EntityID != "site-a/pump-1" || wear.Label != "bearing_wear" || wear.IsNegativeClass || !wear.ExpectedEpisode {
		t.Fatalf("first label = %+v, want the wear on pump-1", wear)
	}
	if quiet.ScenarioID != "pump-station/0002" || quiet.EntityID != "site-a/pump-2" || !quiet.IsNegativeClass ||
		quiet.InjectionTimeNS != model.DefaultStartTimeNS+600e9 {
		t.Fatalf("second label = %+v, want the negative on pump-2 at 600 s", quiet)
	}
	if _, err := os.Stat(filepath.Join(session.out, model.FileLabel)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("several disturbances also wrote label.json: %v", err)
	}
}

func TestASessionWithoutDisturbancesWritesAnEmptyLabelList(t *testing.T) {
	t.Parallel()
	session := runStationSession(t, "pump-station", "", "", nil, 0.05)
	if labels := readLabels(t, session.out); len(labels) != 0 {
		t.Fatalf("labels = %+v, want none", labels)
	}
}

func TestOneDisturbanceWritesOnlyLabelJSON(t *testing.T) {
	t.Parallel()
	session := runStationSession(t, "pump-station", `,"disturbances":[{"at_s":0,"instance":"`+stationPump+`","fault":"bearing_wear"}]`, "", nil, 0.05)
	assertLabelOfTheWear(t, session.out)
	if _, err := os.Stat(filepath.Join(session.out, model.FileLabels)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("one disturbance also wrote labels.json: %v", err)
	}
}
