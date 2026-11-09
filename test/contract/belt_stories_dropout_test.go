package contract

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// S5b is S5-A with D-3's vibration channel dropped out instead of frozen: the world stops delivering that channel at
// the injection, and the loss is a label of its own on D-3, observable at once, with nothing to act on.
const (
	vibrationDropout = "conveyor_drive__vibration__dropout"
	vibrationChannel = "conveyor_drive.vibration"
	afterDropoutS    = 600
)

func dropoutOnsetNS(t *testing.T, scenario *model.Scenario) (int, int64) {
	t.Helper()
	at := slices.IndexFunc(scenario.Commands, func(c model.ScenarioCommand) bool {
		return c.Op == model.OpFaultInject && c.Args["fault"] == vibrationDropout && c.Args["entity_id"] == "D-3"
	})
	if at < 0 {
		t.Fatalf("%s injects no %s on D-3", scenario.ID, vibrationDropout)
	}
	onset, err := scenario.Commands[at].Args["onset_ns"].(json.Number).Int64()
	if err != nil {
		t.Fatal(err)
	}
	return at, onset
}

type deliveredReading struct {
	EntityID  string `json:"entity_id"`
	Channel   string `json:"channel"`
	EventTime string `json:"event_time"`
}

// readingsAround counts each entity's readings of a channel delivered before and at or after a time.
func readingsAround(t *testing.T, trace string, channel string, at int64) (before, after map[string]int) {
	t.Helper()
	file, err := os.Open(trace)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	before, after = map[string]int{}, map[string]int{}
	lines := bufio.NewScanner(file)
	for lines.Scan() {
		var reading deliveredReading
		if err := json.Unmarshal(lines.Bytes(), &reading); err != nil {
			t.Fatal(err)
		}
		if reading.Channel != channel {
			continue
		}
		if eventTime, err := time.Parse(time.RFC3339Nano, reading.EventTime); err != nil {
			t.Fatal(err)
		} else if eventTime.UnixNano() < at {
			before[reading.EntityID]++
		} else {
			after[reading.EntityID]++
		}
	}
	if err := lines.Err(); err != nil {
		t.Fatal(err)
	}
	return before, after
}

func TestS5bDeliversNoVibrationReadingOfD3AfterItsDropout(t *testing.T) {
	t.Parallel()
	scenario := committedScenario(t, "s5b-a.do-nothing.scenario.json")
	at, onset := dropoutOnsetNS(t, scenario)
	scenario.Commands = append(scenario.Commands[:at+1], model.ScenarioCommand{Op: model.OpClockAdvance,
		Args: map[string]any{"to_ns": onset + afterDropoutS*1e9}})
	dir := t.TempDir()
	summary, err := run.RunScenario(t.Context(), scenario, run.ScenarioRunOptions{DomainsDir: filepath.Join(testsupport.RepositoryRoot(), "examples"),
		AdaptersDir: testsupport.AdaptersDir(), SinkName: model.SinkFile, SinkTarget: filepath.Join(dir, model.FileTrace), OutDir: dir, CheckDomain: beltKindsCheck})
	if err != nil || !summary.Reproducible {
		t.Fatalf("RunScenario = %+v, %v", summary, err)
	}
	before, after := readingsAround(t, filepath.Join(dir, model.FileTrace), vibrationChannel, onset)
	t.Logf("vibration readings before the dropout %v, in the %d s after it %v", before, afterDropoutS, after)
	if before["D-3"] == 0 || after["D-3"] != 0 {
		t.Errorf("D-3 delivered %d vibration readings before its dropout and %d after, want some and none", before["D-3"], after["D-3"])
	}
	for _, peer := range survivors {
		if after[peer] < afterDropoutS/60 {
			t.Errorf("%s delivered %d vibration readings in the %d s after D-3's dropout, want one a minute", peer, after[peer], afterDropoutS)
		}
	}
}

func TestS5bLabelsTheDropoutOnD3ObservableAtItsInjectionWithNothingToDo(t *testing.T) {
	t.Parallel()
	scenario := committedScenario(t, "s5b-a.do-nothing.scenario.json")
	_, onset := dropoutOnsetNS(t, scenario)
	spec := beltPlant(t, "a")
	solver, err := truth.NewSolver(spec, scenario.Seed, 120*1e9, 0)
	if err != nil {
		t.Fatal(err)
	}
	label, err := truth.BuildRecord(spec, solver, truth.Injection{ScenarioID: "belt-conveyor-drives-a/0001", Seed: scenario.Seed, EntityID: "D-3",
		FaultID: vibrationDropout, OnsetNS: onset, StartNS: model.DefaultStartTimeNS, EntityIDs: world.PopulationIDs(spec, 0)})
	if err != nil {
		t.Fatal(err)
	}
	if !label.IsPositive() || label.FirstObservableTimeNS != onset || len(label.ExpectedActions) != 0 || label.ExpectedEffector != "" {
		t.Fatalf("dropout label: positive %v, first observable %d (injected %d), expected actions %v %q; want a positive label observable at its injection with no action",
			label.IsPositive(), label.FirstObservableTimeNS, onset, label.ExpectedActions, label.ExpectedEffector)
	}
}
