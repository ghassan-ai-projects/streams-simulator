package domain

import (
	"encoding/json"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// fakeFacts is a domain with fault f-1 (parameter severity only),
// perturbation p-1 (parameter rate only, window must not end before it
// starts), profile nominal and effector act, whose argument "level" is
// required.
type fakeFacts struct{}

func (fakeFacts) HasProfile(name string) bool      { return name == "nominal" }
func (fakeFacts) HasFault(id string) bool          { return id == "f-1" }
func (fakeFacts) HasPerturbation(name string) bool { return name == "p-1" }
func (fakeFacts) HasEffector(name string) bool     { return name == "act" }

func (fakeFacts) CheckFaultParams(_ string, params map[string]any) string {
	return onlyParameter(params, "severity")
}

func (facts fakeFacts) CheckPerturbationParams(name string, params map[string]any) string {
	if !facts.HasPerturbation(name) {
		return "unknown perturbation"
	}
	return onlyParameter(params, "rate")
}

func (fakeFacts) CheckPerturbationWindow(_ string, fromNS, untilNS int64) string {
	if untilNS != 0 && untilNS < fromNS {
		return "window ends before it starts"
	}
	return ""
}

func (fakeFacts) CheckEffectorArgs(_ string, args map[string]any) string {
	if _, ok := args["level"]; !ok {
		return "missing level"
	}
	return ""
}

func onlyParameter(params map[string]any, name string) string {
	for key := range params {
		if key != name {
			return "no parameter " + key
		}
	}
	return ""
}

func command(op string, args map[string]any) model.ScenarioCommand {
	return model.ScenarioCommand{Op: op, Args: args}
}

func at(ns int64) map[string]any { return map[string]any{"at_ns": ns} }

func with(args map[string]any, more map[string]any) map[string]any {
	for key, value := range more {
		args[key] = value
	}
	return args
}

func TestAValidScenarioHasNoIssues(t *testing.T) {
	t.Parallel()
	scenario := &model.Scenario{World: model.ScenarioWorld{StartTimeNS: 10, Profile: "nominal"}, Commands: []model.ScenarioCommand{
		command(model.OpEntityAdd, with(at(11), map[string]any{"entity_id": "e-2"})),
		command(model.OpFaultInject, map[string]any{"entity_id": "e-2", "fault": "f-1", "onset_ns": json.Number("12"), "params": map[string]any{"severity": 1}}),
		command(model.OpPerturbApply, map[string]any{"perturbation": "p-1", "params": nil, "from_ns": int64(12), "until_ns": int64(0)}),
		command(model.OpPerturbApply, map[string]any{"perturbation": "p-1", "params": map[string]any{"rate": 0.1}, "from_ns": int64(12), "until_ns": int64(12)}),
		command(model.OpEffectorInvoke, with(at(13), map[string]any{"entity_id": "e-1", "effector": "act", "args": map[string]any{"level": 1}})),
		command(model.OpClockAdvance, map[string]any{"to_ns": int64(20)}),
		command(model.OpClockAdvance, map[string]any{"to_ns": int64(20)}),
		command(model.OpClockAdvance, map[string]any{"to_ns": 10 + MaxRunSpanNS}),
		command(model.OpEntityRetire, with(at(21), map[string]any{"entity_id": "e-2"})),
	}}
	if issues := CheckScenario(scenario, fakeFacts{}); len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}

func TestEveryViolationIsReportedAtItsArgument(t *testing.T) {
	t.Parallel()
	scenario := &model.Scenario{World: model.ScenarioWorld{StartTimeNS: 10, Profile: "rush"}, Commands: []model.ScenarioCommand{
		command(model.OpFaultInject, map[string]any{"entity_id": "e-1", "fault": "f-9", "onset_ns": int64(10)}),
		command(model.OpPerturbApply, map[string]any{"perturbation": "p-9", "from_ns": int64(0), "until_ns": int64(0)}),
		command(model.OpEffectorInvoke, with(at(10), map[string]any{"entity_id": "e-1", "effector": "act", "args": map[string]any{}})),
		command(model.OpClockAdvance, map[string]any{"to_ns": int64(30)}),
		command(model.OpClockAdvance, map[string]any{"to_ns": int64(20)}),
		command(model.OpEffectorInvoke, with(at(10), map[string]any{"entity_id": "e-1", "effector": "nope"})),
		command(model.OpFaultInject, map[string]any{"entity_id": "e-1", "fault": "f-1", "onset_ns": int64(10), "params": map[string]any{"bogus": 1}}),
		command(model.OpPerturbApply, map[string]any{"perturbation": "p-1", "params": map[string]any{"bogus": 1}, "from_ns": int64(20), "until_ns": int64(25)}),
		command(model.OpPerturbApply, map[string]any{"perturbation": "p-1", "from_ns": int64(20), "until_ns": int64(15)}),
		command(model.OpClockAdvance, map[string]any{"to_ns": json.Number("1.7e18")}),
		command(model.OpClockAdvance, map[string]any{"to_ns": 11 + MaxRunSpanNS}),
	}}
	want := []string{"/world/profile", "/commands/0/args/fault", "/commands/1/args/perturbation", "/commands/2/args/args",
		"/commands/4/args/to_ns", "/commands/5/args/effector", "/commands/6/args/params", "/commands/7/args/params",
		"/commands/8/args/until_ns", "/commands/9/args/to_ns", "/commands/10/args/to_ns"}
	assertPointers(t, CheckScenario(scenario, fakeFacts{}), want)
}

// Entities are not the static check's business: which exist depends on the
// run as it unfolds (churn births and deaths), so the dry run judges them.
func TestTheStaticCheckLeavesEntitiesToTheRun(t *testing.T) {
	t.Parallel()
	scenario := &model.Scenario{World: model.ScenarioWorld{StartTimeNS: 10}, Commands: []model.ScenarioCommand{
		command(model.OpEntityRetire, with(at(11), map[string]any{"entity_id": "ghost"})),
		command(model.OpEntityAdd, with(at(12), map[string]any{"entity_id": "ghost"})),
		command(model.OpEffectorInvoke, with(at(12), map[string]any{"entity_id": "never-born", "effector": "act", "args": map[string]any{"level": 1}})),
	}}
	if issues := CheckScenario(scenario, fakeFacts{}); len(issues) != 0 {
		t.Fatalf("issues = %+v, want none from the static check", issues)
	}
}

// A time that is not an exact 64-bit integer is reported at each such
// argument, and the command is not checked further.
func TestEveryTimeMustBeAnExactInteger(t *testing.T) {
	t.Parallel()
	for _, value := range []any{json.Number("1.5"), json.Number("1e3"), json.Number("99999999999999999999"), "10", nil, 1.5} {
		scenario := &model.Scenario{Commands: []model.ScenarioCommand{
			command(model.OpPerturbApply, map[string]any{"perturbation": "p-9", "from_ns": value, "until_ns": value}),
		}}
		assertPointers(t, CheckScenario(scenario, fakeFacts{}), []string{"/commands/0/args/from_ns", "/commands/0/args/until_ns"})
	}
	if got, exact := ExactTime(map[string]any{"t": float64(1 << 40)}, "t"); !exact || got != 1<<40 {
		t.Fatalf("ExactTime(integral float) = %d, %v", got, exact)
	}
}

func assertPointers(t *testing.T, issues []model.Issue, want []string) {
	t.Helper()
	if len(issues) != len(want) {
		t.Fatalf("issues = %+v, want one at each of %v", issues, want)
	}
	for i, issue := range issues {
		if !issue.Located || issue.Pointer != want[i] || issue.Message == "" {
			t.Errorf("issue %d = %+v, want a message at %s", i, issue, want[i])
		}
	}
}

func TestJournalCommandDropsPacingFromAnAdvanceAndLeavesTheRest(t *testing.T) {
	t.Parallel()
	advance := model.Command{Seq: 2, AtNS: 5, Op: model.OpClockAdvance, Args: map[string]any{"to_ns": int64(9), "await_consumer": true}}
	got := JournalCommand(advance)
	if _, awaited := got.Args["await_consumer"]; awaited || got.Args["to_ns"] != int64(9) || got.Seq != 2 || got.AtNS != 5 {
		t.Fatalf("JournalCommand(advance) = %+v", got)
	}
	if _, kept := advance.Args["await_consumer"]; !kept {
		t.Fatal("the command log's own entry must keep its pacing")
	}
	invoke := model.Command{Op: model.OpEffectorInvoke, Args: map[string]any{"await_consumer": true}}
	if _, kept := JournalCommand(invoke).Args["await_consumer"]; !kept {
		t.Fatal("only an advance drops await_consumer")
	}
}
