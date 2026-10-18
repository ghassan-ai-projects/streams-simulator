package domain

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func pumpKindFile(t *testing.T, edit func(raw string) string) KindFiles {
	t.Helper()
	raw, err := os.ReadFile(testsupport.PumpKindFile())
	if err != nil {
		t.Fatal(err)
	}
	files, _, err := ParseKindFiles([]KindSource{{Path: "pump.kind.json", Raw: []byte(edit(string(raw)))}})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func pumpDomain(t *testing.T, implements string, shape func(kind *model.Kind)) *model.DomainSpec {
	t.Helper()
	compiled, err := domain.Load(testsupport.PumpBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	kind := &compiled.Spec.Kinds[0]
	kind.Implements = implements
	shape(kind)
	return compiled.Spec
}

func unchanged(raw string) string { return raw }

func replacing(from, to string) func(string) string {
	return func(raw string) string { return strings.Replace(raw, from, to, 1) }
}

func driftAt(pointer string) string { return " at " + pointer + ":" }

func TestADomainKindMustEqualTheKindFileItImplements(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		edit       func(raw string) string
		implements string
		shape      func(kind *model.Kind)
		want       error
		at         string
	}{
		{"a match", unchanged, "pump", func(*model.Kind) {}, nil, ""},
		{"a metric unit that drifted", replacing(`"m3/h"`, `"l/s"`), "pump", func(*model.Kind) {}, ErrKindDrift, driftAt("/metrics/0/unit")},
		{"a metric resolution that drifted", replacing(`"resolution": 0.1`, `"resolution": 0.5`), "pump",
			func(*model.Kind) {}, ErrKindDrift, driftAt("/metrics/0/resolution")},
		{"an argument bound that drifted", replacing(`"minimum": 1000`, `"minimum": 2000`), "pump",
			func(*model.Kind) {}, ErrKindDrift, driftAt("/commands/0/args_schema/properties/lease_ms/minimum")},
		{"an argument schema the domain loosened", unchanged, "pump", func(kind *model.Kind) { kind.Effectors[1].ArgsSchema = map[string]any{"type": "object"} },
			ErrKindDrift, driftAt("/commands/1/args_schema/additionalProperties")},
		{"a risk class that drifted", replacing(`"risk_class": "medium"`, `"risk_class": "high"`), "pump",
			func(*model.Kind) {}, ErrKindDrift, driftAt("/commands/0/risk_class")},
		{"a risk class the domain declares differently", unchanged, "pump", func(kind *model.Kind) { kind.Effectors[2].RiskClass = "low" },
			ErrKindDrift, driftAt("/commands/2/risk_class")},
		{"a metric the kind file does not call writable", replacing(`"writable": true`, `"writable": false`), "pump",
			func(*model.Kind) {}, ErrKindDrift, driftAt("/metrics/0/writable")},
		{"a measuring range that drifted", replacing(`"max": 120`, `"max": 130`), "pump", func(*model.Kind) {}, ErrKindDrift, driftAt("/metrics/0/eu_range/max")},
		{"metrics in another order", unchanged, "pump", func(kind *model.Kind) { slices.Reverse(kind.Channels) },
			ErrKindDrift, driftAt("/metrics/0/eu_range/max")},
		{"a metric the kind file lacks", unchanged, "pump", func(kind *model.Kind) {
			kind.Channels = append(kind.Channels, model.Channel{Name: "extra", ValueType: "number"})
		},
			ErrKindDrift, driftAt("/metrics/2")},
		{"a metric the domain lacks", unchanged, "pump", func(kind *model.Kind) { kind.Channels = slices.Delete(kind.Channels, 1, 2) },
			ErrKindDrift, driftAt("/metrics/1")},
		{"a command the kind lacks", unchanged, "pump", func(kind *model.Kind) { kind.Effectors = slices.Delete(kind.Effectors, 1, 2) },
			ErrKindDrift, driftAt("/commands/1/name")},
		{"a command the kind file lacks", unchanged, "pump", func(kind *model.Kind) { kind.Effectors = append(kind.Effectors, model.Effector{Name: "extra"}) },
			ErrKindDrift, driftAt("/commands/3")},
		{"a kind under another name", unchanged, "pump", func(kind *model.Kind) { kind.Name = "pump2" }, ErrKindDrift, driftAt("/kind")},
		{"a safe stop the kind file no longer flags", replacing(`"safe_stop": true`, `"safe_stop": false`),
			"pump", func(*model.Kind) {}, ErrKindDrift, driftAt("/commands/2/safe_stop")},
		{"a command the domain flags as a safe stop and the kind file does not", unchanged, "pump",
			func(kind *model.Kind) { kind.Effectors[0].SafeStop = true }, ErrKindDrift, driftAt("/commands/0/safe_stop")},
		{"a safe stop the domain no longer flags", unchanged, "pump",
			func(kind *model.Kind) { kind.Effectors[2].SafeStop = false }, ErrKindDrift, driftAt("/commands/2/safe_stop")},
		{"numbers written in another form", replacing(`"max": 120}`, `"max": 1.2e2}`), "pump", func(*model.Kind) {}, nil, ""},
		{"an implements naming no kind file", unchanged, "no-such", func(*model.Kind) {}, ErrKindFileMissing, `"no-such"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checked, err := CheckImplements(pumpDomain(t, tc.implements, tc.shape), pumpKindFile(t, tc.edit))
			if tc.want == nil {
				if err != nil || len(checked) != 1 || checked[0].Kind != "pump" {
					t.Fatalf("CheckImplements = %v, %v, want the pump kind accepted and reported", checked, err)
				}
				return
			}
			if checked != nil || !refusedAt(err, tc.want, tc.at) {
				t.Fatalf("CheckImplements = %v, %v, want %v naming %q and no checked kind", checked, err, tc.want, tc.at)
			}
		})
	}
}

func refusedAt(err, want error, at string) bool {
	return errors.Is(err, want) && strings.Contains(err.Error(), at)
}

func TestACheckedKindCarriesTheDigestOfItsKindFile(t *testing.T) {
	t.Parallel()
	files := pumpKindFile(t, unchanged)
	want, err := KindDigest(files["pump"])
	if err != nil {
		t.Fatal(err)
	}
	checked, err := CheckImplements(pumpDomain(t, "pump", func(*model.Kind) {}), files)
	if err != nil || !slices.Equal(checked, []model.ImplementsChecked{{Kind: "pump", KindDigest: want}}) {
		t.Fatalf("CheckImplements = %v, %v, want the pump kind with digest %s", checked, err, want)
	}
}

func TestADomainWithoutImplementsNeedsNoKindFileAndChecksNothing(t *testing.T) {
	t.Parallel()
	checked, err := CheckImplements(pumpDomain(t, "", func(*model.Kind) {}), nil)
	if err != nil || checked == nil || len(checked) != 0 {
		t.Fatalf("CheckImplements = %v, %v, want an empty list and no error", checked, err)
	}
}

func TestADomainThatImplementsFailsClosedWithoutKindFiles(t *testing.T) {
	t.Parallel()
	_, err := CheckImplements(pumpDomain(t, "pump", func(*model.Kind) {}), nil)
	if !errors.Is(err, ErrKindFileMissing) || !strings.Contains(err.Error(), "(0 loaded)") {
		t.Fatalf("CheckImplements = %v, want ErrKindFileMissing counting no loaded kind files", err)
	}
}

func TestAKindThatHasNoCanonicalFormIsRefusedOnEitherSide(t *testing.T) {
	t.Parallel()
	unencodable := func(kind *model.Kind) { kind.Effectors[0].ArgsSchema = map[string]any{"const": make(chan int)} }
	unencodableFile := KindFiles{"pump": {"kind": "pump", "metrics": []any{}, "commands": []any{make(chan int)}}}
	cases := []struct {
		name  string
		spec  *model.DomainSpec
		files KindFiles
	}{
		{"the kind", pumpDomain(t, "pump", unencodable), pumpKindFile(t, unchanged)},
		{"the kind file", pumpDomain(t, "pump", func(*model.Kind) {}), unencodableFile},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := CheckImplements(tc.spec, tc.files)
			var invalid *model.ValidationError
			if !errors.As(err, &invalid) || !errors.Is(err, ErrNotCanonical) || len(invalid.Issues) != 1 || !invalid.Issues[0].Located {
				t.Fatalf("CheckImplements = %v, want a located validation issue caused by ErrNotCanonical", err)
			}
		})
	}
}

func TestTheTwoSidesOfTheComparisonMeetInCanonicalForm(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		a, b    any
		differs bool
	}{
		{"an exponent against a plain integer", json.Number("1e3"), float64(1000), false},
		{"a decimal against its exponent form", json.Number("0.5"), json.Number("5e-1"), false},
		{"equal integers above 2^53", json.Number("9007199254740993"), json.Number("9007199254740993"), false},
		{"integers above 2^53 that differ", json.Number("9007199254740993"), json.Number("9007199254740992"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			want, errWant := canonicalTree(map[string]any{"n": tc.a})
			got, errGot := canonicalTree(map[string]any{"n": tc.b})
			if _, differs := firstDifference(want, got, ""); errWant != nil || errGot != nil || differs != tc.differs {
				t.Fatalf("differs = %v, errors %v %v, want differs = %v", differs, errWant, errGot, tc.differs)
			}
		})
	}
}

func TestAnImplementsMismatchIsADomainValidationErrorAtTheKind(t *testing.T) {
	t.Parallel()
	_, err := CheckImplements(pumpDomain(t, "pump", func(*model.Kind) {}), pumpKindFile(t, replacing(`"m3/h"`, `"l/s"`)))
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || invalid.Contract != model.ContractDomainSpec || len(invalid.Issues) != 1 ||
		invalid.Issues[0].Pointer != "/kinds/0/implements" || !invalid.Issues[0].Located || !errors.Is(invalid.Cause, ErrKindDrift) {
		t.Fatalf("CheckImplements = %+v, want a located domain-spec issue at /kinds/0/implements caused by ErrKindDrift", err)
	}
	if want := `domain: pump-board does not implement its kind files: kind "pump" differs`; !strings.HasPrefix(invalid.Error(), want) {
		t.Fatalf("summary = %q, want it to start %q", invalid.Error(), want)
	}
}

func TestAKindFileHandedOverWithAKindDigestIsNeverTakenAsEqual(t *testing.T) {
	t.Parallel()
	files := pumpKindFile(t, unchanged)
	files["pump"]["kind_digest"] = "sha256:" + strings.Repeat("0", 64)
	checked, err := CheckImplements(pumpDomain(t, "pump", func(*model.Kind) {}), files)
	if checked != nil || !refusedAt(err, ErrKindDrift, driftAt("/kind_digest")) {
		t.Fatalf("CheckImplements = %v, %v, want ErrKindDrift at /kind_digest: the digest is not part of a kind file", checked, err)
	}
}
