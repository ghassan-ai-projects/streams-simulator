package plantwire

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func arduinoSpec(t *testing.T) *domain.Compiled {
	t.Helper()
	spec, err := domain.Load(testsupport.ThermalBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func arduinoKindFiles(t *testing.T) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(testsupport.ThermalBoardKindsDir(), "*.kind.json"))
	if err != nil || len(paths) != 3 {
		t.Fatalf("kind files %v: %v, want the sensor, the indicator and the fan", paths, err)
	}
	return paths
}

func driftedKinds(t *testing.T, file, from, to string) string {
	t.Helper()
	dir := t.TempDir()
	for _, path := range arduinoKindFiles(t) {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		content := string(raw)
		if filepath.Base(path) == file {
			if !strings.Contains(content, from) {
				t.Fatalf("%s holds no %s to change", file, from)
			}
			content = strings.Replace(content, from, to, 1)
		}
		testsupport.WriteFileIn(t, dir, filepath.Base(path), []byte(content))
	}
	return dir
}

func TestEveryKindOfTheArduinoTwinImplementsItsKindFile(t *testing.T) {
	t.Parallel()
	spec := arduinoSpec(t)
	for i := range spec.Spec.Kinds {
		if kind := &spec.Spec.Kinds[i]; kind.Implements != kind.Name {
			t.Errorf("kind %s implements %q, want the kind file of its own name", kind.Name, kind.Implements)
		}
	}
	checked, err := CheckImplements(spec.Spec, []string{testsupport.ThermalBoardKindsDir()})
	if err != nil || len(checked) != 3 {
		t.Fatalf("CheckImplements = %v, %v, want the twin to equal its three kind files", checked, err)
	}
	for _, entry := range checked {
		if want := kindDigests(t)[entry.Kind]; entry.KindDigest != want {
			t.Errorf("checked kind %s has digest %s, its kind file digests to %s", entry.Kind, entry.KindDigest, want)
		}
	}
}

func TestAKindFileThatDriftsFromTheTwinIsRefusedAtTheMemberItNames(t *testing.T) {
	t.Parallel()
	spec := arduinoSpec(t)
	cases := []struct{ name, file, from, to, pointer, at string }{
		{"a sensor unit", "thermal_sensor.kind.json", `"Cel"`, `"degC"`, "/kinds/0/implements", "/metrics/0/unit"},
		{"a command bound", "indicator.kind.json", `"maximum": 1000`, `"maximum": 1023`, "/kinds/1/implements", "/commands/0/args_schema/properties/brightness_permille/maximum"},
		{"the duty range", "fan.kind.json", `"max": 600}`, `"max": 700}`, "/kinds/2/implements", "/metrics/0/eu_range/max"},
		{"the lease floor", "fan.kind.json", `"minimum": 5000`, `"minimum": 1000`, "/kinds/2/implements", "/commands/0/args_schema/properties/lease_ms/minimum"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := CheckImplements(spec.Spec, []string{driftedKinds(t, tc.file, tc.from, tc.to)})
			var invalid *model.ValidationError
			if !errors.As(err, &invalid) || len(invalid.Issues) != 1 || invalid.Issues[0].Pointer != tc.pointer {
				t.Fatalf("CheckImplements = %v, want one domain issue at %s", err, tc.pointer)
			}
			if !strings.Contains(invalid.Issues[0].Message, "at "+tc.at+":") {
				t.Fatalf("issue %q, want it to name the member %s", invalid.Issues[0].Message, tc.at)
			}
		})
	}
}

func arduinoRun(t *testing.T) *run.Run {
	t.Helper()
	a, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := run.New(context.Background(), run.Config{Domain: arduinoSpec(t), Adapter: a, Seed: 5, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func recordingHost(t *testing.T) (string, <-chan []map[string]any) {
	t.Helper()
	path := testsupport.SocketPath(t)
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "unix", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	received := make(chan []map[string]any, 1)
	go recordSession(listener, received)
	return path, received
}

func recordSession(listener net.Listener, received chan<- []map[string]any) {
	conn, err := listener.Accept()
	if err != nil {
		return
	}
	defer func() { _ = conn.Close() }()
	var records []map[string]any
	seq := 0
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		var record map[string]any
		_ = json.Unmarshal(scanner.Bytes(), &record)
		records = append(records, record)
		if answer := hostAnswer(record, seq); answer != nil {
			raw, _ := json.Marshal(answer)
			_, _ = conn.Write(append(raw, '\n'))
			seq++
		}
	}
	received <- records
}

func kindDigests(t *testing.T) map[string]string {
	t.Helper()
	summaries, err := ValidateKindFiles(arduinoKindFiles(t))
	if err != nil {
		t.Fatal(err)
	}
	digests := map[string]string{}
	for _, summary := range summaries {
		digests[summary.Kind] = summary.KindDigest
	}
	return digests
}

func arduinoOptions() Options {
	return Options{Environment: "simulated", ProtocolMax: 1, SessionID: "s-1", SliceNS: 60e9, UntilNS: model.DefaultStartTimeNS + 120e9,
		KindsDirs: []string{testsupport.ThermalBoardKindsDir()}}
}

func TestANodeWhoseSafeStopsDisagreeWithItsKindFilesIsRefused(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, effector, at string
		kind               int
		flagged            bool
	}{
		{"a safe stop the kind file flags but the domain does not", "stop_fan", "/commands/1/safe_stop", 2, false},
		{"a command the domain flags as a safe stop but the kind file does not", "set_led", "/commands/0/safe_stop", 1, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := arduinoRun(t)
			r.Domain().Spec.Kinds[tc.kind].EffectorDecl(tc.effector).SafeStop = tc.flagged
			path, _ := recordingHost(t)
			err := Serve(t.Context(), r, path, 5*time.Second, arduinoOptions())
			if !errors.Is(err, ErrKindDrift) || !strings.Contains(err.Error(), "at "+tc.at+":") {
				t.Fatalf("Serve = %v, want ErrKindDrift naming %s", err, tc.at)
			}
		})
	}
}

func TestTheNodeBirthsTheArduinoBoardAnnouncingTheDigestsOfItsKindFiles(t *testing.T) {
	t.Parallel()
	path, received := recordingHost(t)
	if err := Serve(t.Context(), arduinoRun(t), path, 5*time.Second, arduinoOptions()); err != nil {
		t.Fatal(err)
	}
	records := <-received
	if records[0]["type"] != "NBIRTH" {
		t.Fatalf("first record %v, want NBIRTH", records[0]["type"])
	}
	want, announced := kindDigests(t), map[string]any{}
	for _, entry := range records[0]["kinds"].([]any) {
		kind := entry.(map[string]any)
		announced[kind["kind"].(string)] = kind["kind_digest"]
	}
	if len(announced) != 3 {
		t.Fatalf("NBIRTH announces %v, want the three kinds", announced)
	}
	for name, digest := range want {
		if announced[name] != digest {
			t.Errorf("kind %s announced as %v, its kind file digests to %s", name, announced[name], digest)
		}
	}
	assertBoardBirths(t, records)
}

func assertBoardBirths(t *testing.T, records []map[string]any) {
	t.Helper()
	boards := records[0]["boards"].([]any)
	if len(boards) != 1 || boards[0].(map[string]any)["board"] != "mega-01" || len(boards[0].(map[string]any)["members"].([]any)) != 3 {
		t.Fatalf("NBIRTH boards %v, want mega-01 with its three devices", boards)
	}
	kinds, data := map[string]string{}, 0
	for _, record := range records {
		switch record["type"] {
		case "DBIRTH":
			kinds[record["device"].(string)] = record["kind"].(string)
		case "DDATA":
			data++
		}
	}
	if kinds["zone-01"] != "thermal_sensor" || kinds["led-01"] != "indicator" || kinds["fan-01"] != "fan" || data == 0 {
		t.Fatalf("devices born %v with %d DDATA records, want zone-01, led-01 and fan-01 reporting", kinds, data)
	}
}
