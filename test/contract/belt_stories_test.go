package contract

import (
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

type storyTable struct {
	SafeLimpFeedPct map[string]float64 `json:"safe_limp_feed_pct"`
	Stories         []beltStory        `json:"stories"`
}

type beltStory struct {
	ID            string             `json:"id"`
	Story         string             `json:"story"`
	Plant         string             `json:"plant"`
	Regime        string             `json:"regime"`
	Title         string             `json:"title"`
	NegativeClass bool               `json:"negative_class"`
	Unavoidable   bool               `json:"unavoidable"`
	HorizonS      float64            `json:"horizon_s"`
	Pace          string             `json:"pace"`
	Disturbances  []storyDisturbance `json:"disturbances"`
	Policy        []storyCall        `json:"policy"`
	StopLeadS     float64            `json:"planned_stop_lead_s"`
	Rows          []map[string]any   `json:"rows"`
}

type storyDisturbance struct {
	Instance  string   `json:"instance"`
	Fault     string   `json:"fault"`
	At        string   `json:"at"`
	AtS       float64  `json:"at_s"`
	Severity  any      `json:"severity"`
	DurationS *float64 `json:"duration_s"`
}

type storyCall struct {
	Anchor   string         `json:"anchor"`
	AfterS   float64        `json:"after_s"`
	Effector string         `json:"effector"`
	Instance string         `json:"instance"`
	Args     map[string]any `json:"args"`
}

const (
	eventInject = "inject"
	eventClear  = "clear"
	eventInvoke = "invoke"
)

type storyEvent struct {
	atS       float64
	kind      string
	instance  string
	name      string
	severity  *float64
	durationS *float64
	args      map[string]any
	ordinal   int
}

func loadStories(t *testing.T) storyTable {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(testsupport.BeltConveyorStoriesDir(), "stories.json"))
	if err != nil {
		t.Fatal(err)
	}
	var table storyTable
	if err := json.Unmarshal(raw, &table); err != nil {
		t.Fatal(err)
	}
	return table
}

func (s beltStory) domain() string { return "belt-conveyor-drives-" + s.Plant }

func (s beltStory) regime() beltRegime {
	name := cmp.Or(s.Regime, s.Plant)
	return beltRegimes[slices.IndexFunc(beltRegimes, func(r beltRegime) bool { return r.name == name })]
}

func rowNumbers(row map[string]any, field string) []float64 {
	switch value := row[field].(type) {
	case float64:
		return []float64{value}
	case []any:
		numbers := make([]float64, len(value))
		for i, item := range value {
			numbers[i] = item.(float64)
		}
		return numbers
	}
	panic(fmt.Sprintf("row field %q is missing", field))
}

func (d storyDisturbance) events(row map[string]any) []storyEvent {
	times := []float64{d.AtS}
	if d.At != "" {
		times = rowNumbers(row, d.At)
	}
	events := make([]storyEvent, len(times))
	for i, at := range times {
		events[i] = storyEvent{atS: at, kind: eventInject, instance: d.Instance, name: d.Fault, durationS: d.DurationS}
		switch severity := d.Severity.(type) {
		case float64:
			events[i].severity = &severity
		case string:
			events[i].severity = &rowNumbers(row, severity)[i]
		}
	}
	return events
}

// events lists a row's disturbances, their clears and, with the policy, its calls, in time order; a clear names its
// fault by the order in which the world numbers injected faults.
func (s beltStory) events(row map[string]any, withPolicy bool) []storyEvent {
	var events []storyEvent
	for _, d := range s.Disturbances {
		events = append(events, d.events(row)...)
	}
	if withPolicy {
		for _, call := range s.Policy {
			at := rowNumbers(row, call.Anchor)[0] + call.AfterS
			events = append(events, storyEvent{atS: at, kind: eventInvoke, instance: call.Instance, name: call.Effector, args: call.Args})
		}
	}
	slices.SortStableFunc(events, func(a, b storyEvent) int { return compareFloat(a.atS, b.atS) })
	return withClears(events)
}

func withClears(events []storyEvent) []storyEvent {
	var out []storyEvent
	injected := 0
	for _, event := range events {
		if event.kind == eventInject {
			event.ordinal = injected
			injected++
			if event.durationS != nil {
				out = append(out, storyEvent{atS: event.atS + *event.durationS, kind: eventClear, name: event.name, ordinal: event.ordinal})
			}
		}
		out = append(out, event)
	}
	slices.SortStableFunc(out, func(a, b storyEvent) int { return compareFloat(a.atS, b.atS) })
	return out
}

func compareFloat(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// effectorArgs gives a start or stop of a drive or the belt its run command, 1 or 0, unless the call names it.
func effectorArgs(effector string, args map[string]any) map[string]any {
	_, command, _ := strings.Cut(effector, "__")
	verb, _, _ := strings.Cut(command, "_")
	all := map[string]any{}
	if run, commanded := map[string]float64{"start": 1, "stop": 0}[verb]; commanded {
		all["run_command"] = run
	}
	maps.Copy(all, args)
	return all
}

func TestTheStoryTableHoldsTenRowsPerVariantAndResolvesEveryEvent(t *testing.T) {
	t.Parallel()
	table := loadStories(t)
	want := []string{"s0-a", "s1-a", "s1-b", "s2-a", "s2-b", "s3-a", "s4-a", "s4-b", "s5-a", "s6-a-stop", "s6-a-promote"}
	var ids []string
	for _, story := range table.Stories {
		ids = append(ids, story.ID)
		if len(story.Rows) != 10 {
			t.Errorf("%s: %d seed rows, want 10", story.ID, len(story.Rows))
		}
		for _, row := range story.Rows {
			for _, event := range story.events(row, true) {
				if event.atS < 0 || event.atS > story.HorizonS {
					t.Errorf("%s seed %v: %s %s at %.0f s, outside the horizon %.0f s", story.ID, row["seed"], event.kind, event.name, event.atS, story.HorizonS)
				}
			}
		}
	}
	if !slices.Equal(ids, want) {
		t.Fatalf("stories %v, want %v", ids, want)
	}
}
