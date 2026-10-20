package domain

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func semanticDir() string {
	return filepath.Join(testsupport.RepositoryRoot(), "internal", "plantwire", "testdata", "semantic")
}

func semanticSources(t *testing.T, suffix string, edit func(name, raw string) string) []KindSource {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(semanticDir(), "*"+suffix))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no %s files: %v", suffix, err)
	}
	sources := make([]KindSource, len(paths))
	for i, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		sources[i] = KindSource{Path: path, Raw: []byte(edit(filepath.Base(path), string(raw)))}
	}
	return sources
}

func asIs(_, raw string) string { return raw }

func mappedThermal(t *testing.T) *model.DomainSpec {
	t.Helper()
	compiled, err := domain.Load(testsupport.ThermalBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	for i := range compiled.Spec.Kinds {
		if name := compiled.Spec.Kinds[i].Name; name == "fan" || name == "indicator" {
			compiled.Spec.Kinds[i].BoardType = "arduino-thermal"
		}
	}
	return compiled.Spec
}

func thermalInterfaces(t *testing.T, editKinds, editBoard func(name, raw string) string) ([]model.ImplementsChecked, Interfaces, error) {
	t.Helper()
	kinds, summaries, err := ParseKindFiles(semanticSources(t, KindFileSuffix, editKinds))
	if err != nil {
		t.Fatal(err)
	}
	boards, _, err := ParseBoardFiles(semanticSources(t, BoardFileSuffix, editBoard))
	if err != nil {
		t.Fatal(err)
	}
	checked, faces, err := CheckInterfaces(mappedThermal(t), kinds, boards)
	if err == nil && len(checked) != len(summaries) {
		t.Fatalf("checked %v, kind files %v", checked, summaries)
	}
	return checked, faces, err
}

func TestTheRawThermalKindsProjectThroughTheBoardMappingOntoTheSemanticKindFiles(t *testing.T) {
	t.Parallel()
	checked, faces, err := thermalInterfaces(t, asIs, asIs)
	if err != nil {
		t.Fatal(err)
	}
	kinds, summaries, _ := ParseKindFiles(semanticSources(t, KindFileSuffix, asIs))
	for _, summary := range summaries {
		want, _ := KindDigest(kinds[summary.Kind])
		if !containsChecked(checked, summary.Kind, want) {
			t.Errorf("%s: checked %v, want the kind file's own digest %s", summary.Kind, checked, want)
		}
	}
	fan, mapped := faces.Of(specKind(mappedThermal(t), "fan"))
	if mapped == nil || fan.Channels[0].Name != "level_percent" || fan.Effectors[0].Name != "set_cooling" || fan.Channels[0].Observes != "fan_duty" {
		t.Fatalf("projected fan %+v", fan)
	}
	if sensor, mapped := faces.Of(specKind(mappedThermal(t), "thermal_sensor")); mapped != nil || sensor.Channels[0].Name != "temperature" {
		t.Errorf("an unmapped kind is its own interface, got %+v", sensor)
	}
}

func specKind(spec *model.DomainSpec, name string) *model.Kind {
	for i := range spec.Kinds {
		if spec.Kinds[i].Name == name {
			return &spec.Kinds[i]
		}
	}
	return nil
}

func containsChecked(checked []model.ImplementsChecked, kind, digest string) bool {
	for _, entry := range checked {
		if entry.Kind == kind && entry.KindDigest == digest {
			return true
		}
	}
	return false
}

func TestAMappingThatDisagreesWithItsKindFileOrItsRawKindIsDriftAtTheMember(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		kind, board func(name, raw string) string
		want        string
	}{
		"a scale that misses the raw range": {asIs, swap("arduino-thermal", `"scale": 10}`, `"scale": 9}`), "/metrics/level_percent/eu_range"},
		"a kind range the raw argument cannot hold": {swap("fan.kind", `"maximum": 60, "minimum": 0`, `"maximum": 70, "minimum": 0`), asIs,
			"/commands/set_cooling/args/duty_permille"},
		"a word the kind does not offer": {swap("indicator.kind", `["off", "watch", "alert"], "type"`, `["off", "watch", "panic"], "type"`), asIs,
			"/commands/show_state/table"},
		"a reading word the kind does not offer": {swap("indicator.kind", `["off", "watch", "alert"], "resolution"`, `["off", "watch"], "resolution"`), asIs,
			"/metrics/state/table"},
		"a raw command the twin lacks":        {asIs, swap("arduino-thermal", `"raw": "set_pwm_lease"`, `"raw": "set_pwm"`), "/commands/set_cooling/raw"},
		"a command the kind file lacks":       {swap("fan.kind", `"name": "stop_fan"`, `"name": "halt_fan"`), asIs, "/commands/halt_fan"},
		"a from that names no argument":       {asIs, swap("arduino-thermal", `"from": "level_percent"`, `"from": "level"`), "/commands/set_cooling/args/duty_permille/from"},
		"a lease the board refuses":           {asIs, swap("arduino-thermal", `"ms": 5000`, `"ms": 20000`), "/commands/set_cooling/lease"},
		"a risk class the twin does not have": {swap("fan.kind", `"risk_class": "medium"`, `"risk_class": "low"`), asIs, "commands"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, _, err := thermalInterfaces(t, c.kind, c.board)
			var invalid *model.ValidationError
			if !errors.Is(err, ErrKindDrift) || !errors.As(err, &invalid) || invalid.Issues[0].Keyword != "implements_drift" || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err %v, want implements_drift at %s", err, c.want)
			}
		})
	}
}

func swap(file, old, replacement string) func(name, raw string) string {
	return func(name, raw string) string {
		if !strings.HasPrefix(name, file) {
			return raw
		}
		return strings.ReplaceAll(raw, old, replacement)
	}
}

func TestAKindThatNamesABoardMappingNobodyLoadedIsMissing(t *testing.T) {
	t.Parallel()
	kinds, _, err := ParseKindFiles(semanticSources(t, KindFileSuffix, asIs))
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = CheckInterfaces(mappedThermal(t), kinds, nil)
	var invalid *model.ValidationError
	if !errors.Is(err, ErrBoardMissing) || !errors.As(err, &invalid) || invalid.Issues[0].Keyword != "implements_missing" {
		t.Fatalf("no board mapping loaded: %v, want implements_missing", err)
	}
}
