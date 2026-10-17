package domain

import (
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
		{"a metric unit that drifted", func(raw string) string { return strings.Replace(raw, `"m3/h"`, `"l/s"`, 1) }, "pump",
			func(*model.Kind) {}, ErrKindDrift, "/metrics/0/unit"},
		{"a metric resolution that drifted", func(raw string) string { return strings.Replace(raw, `"resolution": 0.1`, `"resolution": 0.5`, 1) }, "pump",
			func(*model.Kind) {}, ErrKindDrift, "/metrics/0/resolution"},
		{"a command the kind lacks", unchanged, "pump", func(kind *model.Kind) { kind.Effectors = slices.Delete(kind.Effectors, 1, 2) },
			ErrKindDrift, "/commands/1/name"},
		{"a command the kind file lacks", unchanged, "pump", func(kind *model.Kind) { kind.Effectors = append(kind.Effectors, model.Effector{Name: "extra"}) },
			ErrKindDrift, "/commands/3"},
		{"a kind under another name", unchanged, "pump", func(kind *model.Kind) { kind.Name = "pump2" }, ErrKindDrift, "/kind"},
		{"a safe-stop flag, which the node configures and the kind file carries", func(raw string) string { return strings.Replace(raw, `"safe_stop": true`, `"safe_stop": false`, 1) },
			"pump", func(*model.Kind) {}, nil, ""},
		{"a kind_digest member left in the kind file", func(raw string) string {
			return strings.Replace(raw, `"kind": "pump",`, `"kind": "pump", "kind_digest": "sha256:00",`, 1)
		}, "pump", func(*model.Kind) {}, nil, ""},
		{"an implements naming no kind file", unchanged, "no-such", func(*model.Kind) {}, ErrKindFileMissing, `"no-such"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := CheckImplements(pumpDomain(t, tc.implements, tc.shape), pumpKindFile(t, tc.edit))
			if tc.want == nil && err != nil {
				t.Fatalf("CheckImplements = %v, want the kind accepted", err)
			}
			if tc.want != nil && (!errors.Is(err, tc.want) || !strings.Contains(err.Error(), tc.at)) {
				t.Fatalf("CheckImplements = %v, want %v naming %q", err, tc.want, tc.at)
			}
		})
	}
}

func TestADomainWithoutImplementsNeedsNoKindFile(t *testing.T) {
	t.Parallel()
	if err := CheckImplements(pumpDomain(t, "", func(*model.Kind) {}), nil); err != nil {
		t.Fatalf("CheckImplements = %v, want a domain without implements unchanged", err)
	}
}

func TestADomainThatImplementsFailsClosedWithoutKindFiles(t *testing.T) {
	t.Parallel()
	err := CheckImplements(pumpDomain(t, "pump", func(*model.Kind) {}), nil)
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
			err := CheckImplements(tc.spec, tc.files)
			if err == nil || !strings.Contains(err.Error(), "canonical form of a kind") {
				t.Fatalf("CheckImplements = %v, want a canonical-form refusal", err)
			}
		})
	}
}

func TestAnImplementsMismatchIsADomainValidationErrorAtTheKind(t *testing.T) {
	t.Parallel()
	edit := func(raw string) string { return strings.Replace(raw, `"m3/h"`, `"l/s"`, 1) }
	err := CheckImplements(pumpDomain(t, "pump", func(*model.Kind) {}), pumpKindFile(t, edit))
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || invalid.Contract != model.ContractDomainSpec || len(invalid.Issues) != 1 ||
		invalid.Issues[0].Pointer != "/kinds/0/implements" || !invalid.Issues[0].Located || !errors.Is(invalid.Cause, ErrKindDrift) {
		t.Fatalf("CheckImplements = %+v, want a located domain-spec issue at /kinds/0/implements caused by ErrKindDrift", err)
	}
	if want := `domain: pump-board does not implement its kind files: kind "pump" differs`; !strings.HasPrefix(invalid.Error(), want) {
		t.Fatalf("summary = %q, want it to start %q", invalid.Error(), want)
	}
}
