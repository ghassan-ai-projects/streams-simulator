package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func semanticKindsDir() string {
	return filepath.Join(testsupport.RepositoryRoot(), "internal", "plantwire", "testdata", "semantic")
}

func mappedThermalDomains(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(testsupport.ThermalBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	for _, kind := range document["kinds"].([]any) {
		if fields := kind.(map[string]any); fields["name"] == "fan" || fields["name"] == "indicator" {
			fields["board_type"] = "arduino-thermal"
		}
	}
	dir := t.TempDir()
	out, _ := json.Marshal(document)
	testsupport.WriteFileIn(t, dir, filepath.Base(testsupport.ThermalBoardDomain()), out)
	return dir
}

func mappedThermalRun(t *testing.T) (*run.Run, wire.Interfaces) {
	t.Helper()
	r := runOf(t, filepath.Join(mappedThermalDomains(t), filepath.Base(testsupport.ThermalBoardDomain())))
	_, faces, err := loadInterfaces(r.Domain().Spec, []string{semanticKindsDir()})
	if err != nil {
		t.Fatal(err)
	}
	return r, faces
}

func mappedCommand(h *scriptedHost, id, device, command string, args map[string]any) map[string]any {
	return withKey(set(commandFields(id, device, command, args, sliceEnd(1)), "expected_boot_id", bootOf(h.sent, device)))
}

func TestAMappedNodeSpeaksTheKindFilesAndDrivesTheRawTwin(t *testing.T) {
	t.Parallel()
	r, faces := mappedThermalRun(t)
	host := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) {
		if syncs != 1 {
			return
		}
		h.reply("DCMD", mappedCommand(h, "c-cool", "fan-01", "set_cooling", map[string]any{"level_percent": 45}))
		h.reply("DCMD", mappedCommand(h, "c-watch", "led-01", "show_state", map[string]any{"state": "watch"}))
		h.reply("DCMD", mappedCommand(h, "c-high", "fan-01", "set_cooling", map[string]any{"level_percent": 70}))
		h.reply("DCMD", mappedCommand(h, "c-panic", "led-01", "show_state", map[string]any{"state": "panic"}))
		h.reply("DCMD", mappedCommand(h, "c-raw", "fan-01", "set_pwm_lease", map[string]any{"duty_permille": 450, "lease_ms": 5000}))
	}}
	options := sliceOptions(3)
	options.Policies, options.Interfaces = []string{policy}, faces
	if err := NewSession(r, host, options).Run(t.Context()); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"c-cool": "DRECEIPT:accepted", "c-watch": "DRECEIPT:accepted", "c-high": "DRECEIPT:refusedout_of_range",
		"c-panic": "DRECEIPT:refusedout_of_range", "c-raw": "DRECEIPT:refusedunknown_command"}
	for id, answer := range want {
		if got := answersOf(host.sent, id); len(got) == 0 || got[0] != answer {
			t.Errorf("%s: answers %v, want first %s", id, got, answer)
		}
	}
	assertTheTwinRanTheRawCommands(t, r)
	assertTheWireSpeaksTheKindFiles(t, host.sent)
}

func answersOf(sent []map[string]any, id string) []string {
	var answers []string
	for _, record := range sent {
		if record["command_id"] == id {
			answers = append(answers, fmt.Sprint(record["type"])+":"+str(record["status"])+str(record["reason"]))
		}
	}
	return answers
}

func assertTheTwinRanTheRawCommands(t *testing.T, r *run.Run) {
	t.Helper()
	calls := map[string]map[string]any{}
	for _, call := range r.World.EffectorCalls() {
		calls[call.Effector] = call.Args
	}
	if len(calls) != 2 || !reflect.DeepEqual(calls["set_pwm_lease"], map[string]any{"duty_permille": int64(450), "lease_ms": int64(5000)}) ||
		!reflect.DeepEqual(calls["set_led"], map[string]any{"brightness_permille": int64(500), "pattern": int64(1)}) {
		t.Fatalf("the twin ran %v, want set_pwm_lease 450/5000 and set_led 500/1", calls)
	}
}

func assertTheWireSpeaksTheKindFiles(t *testing.T, sent []map[string]any) {
	t.Helper()
	kinds, _, err := wire.ParseKindFiles(semanticSourcesOf(t))
	if err != nil {
		t.Fatal(err)
	}
	birth := birthsOfRecords(sent)[0]
	for _, entry := range birth["kinds"].([]any) {
		kind := entry.(map[string]any)
		digest, _ := wire.KindDigest(kinds[kind["kind"].(string)])
		if kind["kind_digest"] != digest {
			t.Errorf("NBIRTH announces %v with %v, its kind file's digest is %s", kind["kind"], kind["kind_digest"], digest)
		}
	}
	levels := readingsOf(sent, "fan-01", "level_percent")
	states := readingsOf(sent, "led-01", "state")
	if !slices.ContainsFunc(levels, func(r reading) bool { return r.value == 45 }) || len(readingsOf(sent, "fan-01", "duty_permille")) != 0 {
		t.Errorf("fan readings %v, want level_percent 45 after the command and no raw duty", levels)
	}
	if len(states) == 0 || !slices.ContainsFunc(sent, func(r map[string]any) bool { return containsWord(r, "watch") }) {
		t.Errorf("indicator readings %v, want the state word watch on the wire", states)
	}
}
