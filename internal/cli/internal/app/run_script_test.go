package app

import (
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestRunOptionsPreserveDefaultsEpochZeroAndFileTarget(t *testing.T) {
	t.Parallel()
	options, err := parseRunOptions([]string{"--domain", "rotating-machinery", "--start-time", "0", "--sink", "file", "--out", "results"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if options.startTime != 0 || options.seed != 1 || options.durationS != 6*3600 || options.adapterID != "native-jsonl" || options.sinkTarget != filepath.Join("results", "trace.jsonl") {
		t.Fatalf("options=%+v", options)
	}
	options, err = parseRunOptions([]string{"--domain", "rotating-machinery", "--sink", "file"}, io.Discard)
	if err == nil || err.Error() != "--sink=file requires --sink-target or --out" {
		t.Fatalf("missing file destination: options=%+v err=%v", options, err)
	}
}

func TestFlagsDescribeFaultsThenPerturbationsThenEffectorsThenTheAdvance(t *testing.T) {
	t.Parallel()
	options, err := parseRunOptions([]string{"--domain", "rotating-machinery", "--start-time", "1000", "--duration", "10",
		"--fault", "site-a/pump-1=bearing_wear@1", "--perturb", "drop@2@3", "--effector", "stop@site-a/pump-1@4"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	scenario, origins, err := scenarioFromFlags(options)
	if err != nil {
		t.Fatal(err)
	}
	wantOrigins := []string{`--fault "site-a/pump-1=bearing_wear@1"`, `--perturb "drop@2@3"`, `--effector "stop@site-a/pump-1@4"`, "--duration 10"}
	if strings.Join(origins, "|") != strings.Join(wantOrigins, "|") {
		t.Fatalf("origins = %q, want %q", origins, wantOrigins)
	}
	var ops []string
	for _, command := range scenario.Commands {
		ops = append(ops, command.Op)
	}
	want := []string{model.OpFaultInject, model.OpPerturbApply, model.OpEffectorInvoke, model.OpClockAdvance}
	if strings.Join(ops, ",") != strings.Join(want, ",") {
		t.Fatalf("ops = %v, want %v", ops, want)
	}
	if onset := scenario.Commands[0].Args["onset_ns"]; onset != int64(1000+1e9) {
		t.Fatalf("fault onset = %v, want start + 1 s", onset)
	}
	if end := scenario.Commands[3].Args["to_ns"]; end != int64(1000+10e9) {
		t.Fatalf("advance = %v, want start + duration", end)
	}
}

func TestScriptedEffectorArgumentsMustBeAJSONObject(t *testing.T) {
	t.Parallel()
	_, err := parseEffectorSpec(`start_aerator@site-a/pond-1@5@[1,2]`)
	if err == nil || !strings.Contains(err.Error(), "arguments must be a JSON object") {
		t.Fatalf("err = %v", err)
	}
	spec, err := parseEffectorSpec("stop@e-1")
	if err != nil || spec.offsetNS != 0 || len(spec.args) != 0 {
		t.Fatalf("spec = %+v, err = %v", spec, err)
	}
}
