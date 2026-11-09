package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const (
	policyNone      = "do-nothing"
	policyReference = "reference"
	placeholderHost = "/tmp/streamsim-host.sock"
)

func storyNS(seconds float64) int64 { return model.DefaultStartTimeNS + int64(math.Round(seconds*1e9)) }

func (s beltStory) policies() []string {
	if len(s.Policy) == 0 {
		return []string{policyNone}
	}
	return []string{policyNone, policyReference}
}

func (s beltStory) scenario(row map[string]any, policy string) *model.Scenario {
	scenario := &model.Scenario{SchemaVersion: model.ScenarioSchemaVersion, ID: s.ID + "-" + policy,
		Description: fmt.Sprintf("%s: %s, seed row 1 (%s).", s.Story, s.Title, policy), Domain: s.domain(),
		Adapter: "native-jsonl", Seed: uint64(row["seed"].(float64)), World: model.ScenarioWorld{StartTimeNS: model.DefaultStartTimeNS}}
	clock := 0.0
	for i, event := range s.events(row, policy == policyReference) {
		if event.atS > clock {
			scenario.Commands = append(scenario.Commands, advanceTo(event.atS))
			clock = event.atS
		}
		scenario.Commands = append(scenario.Commands, event.command(fmt.Sprintf("%s-%d", s.ID, i+1)))
	}
	return appendHorizon(scenario, s.HorizonS)
}

func advanceTo(seconds float64) model.ScenarioCommand {
	return model.ScenarioCommand{Op: model.OpClockAdvance, Args: map[string]any{"to_ns": storyNS(seconds)}}
}

func appendHorizon(scenario *model.Scenario, horizonS float64) *model.Scenario {
	scenario.Commands = append(scenario.Commands, advanceTo(horizonS))
	return scenario
}

func (e storyEvent) command(commandID string) model.ScenarioCommand {
	switch e.kind {
	case eventInject:
		var params map[string]any
		if e.severity != nil {
			params = map[string]any{"severity": *e.severity}
		}
		return model.ScenarioCommand{Op: model.OpFaultInject, Args: map[string]any{"entity_id": e.instance, "fault": e.name, "onset_ns": storyNS(e.atS), "params": params}}
	case eventClear:
		return model.ScenarioCommand{Op: model.OpFaultClear, Args: map[string]any{"fault_id": world.FaultID(e.ordinal), "at_ns": storyNS(e.atS)}}
	}
	return model.ScenarioCommand{Op: model.OpEffectorInvoke, Args: map[string]any{"effector": e.name, "entity_id": e.instance,
		"command_id": commandID, "args": effectorArgs(e.name, e.instance, e.args), "at_ns": storyNS(e.atS)}}
}

func (s beltStory) sessionPlan(row map[string]any) *model.SessionPlan {
	seed := uint64(row["seed"].(float64))
	horizon := s.HorizonS
	plan := &model.SessionPlan{SchemaVersion: model.SessionPlanSchemaVersion, Domain: s.domain(), Seed: &seed, HostSocket: placeholderHost, HorizonS: &horizon,
		Pace: s.Pace}
	for _, event := range s.events(row, false) {
		if event.kind == eventInject {
			plan.Disturbances = append(plan.Disturbances, model.Disturbance{AtS: event.atS, Instance: event.instance, Fault: event.name,
				DurationS: event.durationS, Severity: event.severity})
		}
	}
	return plan
}

type storyFile struct {
	path    string
	content []byte
}

func storyFiles(t *testing.T) []storyFile {
	t.Helper()
	dir := testsupport.BeltConveyorStoriesDir()
	var files []storyFile
	for _, story := range loadStories(t).Stories {
		for _, policy := range story.policies() {
			name := filepath.Join(dir, "scenarios", story.ID+"."+policy+".scenario.json")
			files = append(files, storyFile{name, indented(t, story.scenario(story.Rows[0], policy))})
		}
		files = append(files, storyFile{filepath.Join(dir, "sessions", story.ID+".session-plan.json"), indented(t, story.sessionPlan(story.Rows[0]))})
	}
	return files
}

func indented(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return append(raw, '\n')
}

// The committed scenario and session files are generated from stories.json; UPDATE_GOLDEN=1 writes them.
func TestTheCommittedStoryFilesAreWhatTheStoryTableGenerates(t *testing.T) {
	t.Parallel()
	for _, file := range storyFiles(t) {
		if os.Getenv("UPDATE_GOLDEN") != "" {
			if err := os.WriteFile(file.path, file.content, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		committed, err := os.ReadFile(file.path)
		if err != nil || !bytes.Equal(committed, file.content) {
			t.Errorf("%s differs from what stories.json generates (%v); regenerate with UPDATE_GOLDEN=1", filepath.Base(file.path), err)
		}
	}
}

func TestEveryStoryScenarioAndSessionPlanFollowsItsContract(t *testing.T) {
	t.Parallel()
	for _, file := range storyFiles(t) {
		if bytes.HasSuffix([]byte(file.path), []byte(".scenario.json")) {
			if _, err := model.DecodeScenario(file.content); err != nil {
				t.Errorf("%s: %v", filepath.Base(file.path), err)
			}
			continue
		}
		check, err := plantwire.CheckConnect(plantwire.ConnectOptions{Plan: bytes.NewReader(file.content), Source: filepath.Base(file.path),
			DomainsDir: filepath.Join(testsupport.RepositoryRoot(), "examples"), AdaptersDir: testsupport.AdaptersDir(),
			KindsDirs: []string{testsupport.BeltConveyorKindsDir()}, SliceNS: 1e9})
		if err != nil || len(check.Instances) != 5 {
			t.Errorf("%s: CheckConnect = %+v, %v", filepath.Base(file.path), check, err)
		}
	}
}

func committedScenario(t *testing.T, name string) *model.Scenario {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(testsupport.BeltConveyorStoriesDir(), "scenarios", name))
	if err != nil {
		t.Fatal(err)
	}
	scenario, err := model.DecodeScenario(raw)
	if err != nil {
		t.Fatal(err)
	}
	return scenario
}

func beltKindsCheck(spec *model.DomainSpec) ([]model.ImplementsChecked, error) {
	checked, err := plantwire.CheckImplements(spec, []string{testsupport.BeltConveyorKindsDir()})
	if err != nil {
		return nil, fmt.Errorf("check the conveyor kind files: %w", err)
	}
	return checked, nil
}

func runCommitted(t *testing.T, scenario *model.Scenario) {
	t.Helper()
	summary, err := run.RunScenario(t.Context(), scenario, run.ScenarioRunOptions{DomainsDir: filepath.Join(testsupport.RepositoryRoot(), "examples"),
		AdaptersDir: testsupport.AdaptersDir(), SinkName: model.SinkInproc, OutDir: t.TempDir(), CheckDomain: beltKindsCheck})
	if err != nil || !summary.Reproducible || summary.Emitted == 0 || len(summary.ImplementsChecked) != 3 {
		t.Fatalf("RunScenario = %+v, %v", summary, err)
	}
}

// The CI tier runs the committed file up to a minute after its first command to the plant; the long tier runs it whole.
func TestACommittedStoryScenarioRunsAsAUserRunsIt(t *testing.T) {
	t.Parallel()
	scenario := committedScenario(t, "s2-a.reference.scenario.json")
	if !testing.Short() {
		runCommitted(t, scenario)
		return
	}
	first := slices.IndexFunc(scenario.Commands, func(c model.ScenarioCommand) bool { return c.Op == model.OpEffectorInvoke })
	at, err := scenario.Commands[first].Args["at_ns"].(json.Number).Int64()
	if err != nil {
		t.Fatal(err)
	}
	cutoff := at + 60e9
	scenario.Commands = append(scenario.Commands[:first+1], model.ScenarioCommand{Op: model.OpClockAdvance, Args: map[string]any{"to_ns": cutoff}})
	runCommitted(t, scenario)
}
