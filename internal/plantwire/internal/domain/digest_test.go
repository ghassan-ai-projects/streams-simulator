package domain

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestTheKindDigestOfAnEmptyKindIsTheKnownAnswer(t *testing.T) {
	t.Parallel()
	entry := map[string]any{"kind": "k", "metrics": []any{}, "commands": []any{}}
	got, err := KindDigest(entry)
	if err != nil || got != "sha256:540c36216dd6364d21e22923590e1cfc5791f7a06c51774cd509214b0451a177" {
		t.Fatalf("kind digest %s, %v", got, err)
	}
}

func decodedKind(t *testing.T, raw string) map[string]any {
	t.Helper()
	var entry map[string]any
	if err := decode([]byte(raw), &entry); err != nil {
		t.Fatal(err)
	}
	return entry
}

func TestTheKindDigestFollowsTheContentAndNeitherTheMemberOrderNorAnAnnouncedDigest(t *testing.T) {
	t.Parallel()
	const base = `{"kind": "k", "metrics": [{"name": "m", "data_type": "number", "unit": "Cel"}], "commands": []}`
	digest := func(raw string) string {
		t.Helper()
		got, err := KindDigest(decodedKind(t, raw))
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	want := digest(base)
	same := map[string]string{
		"members in another order": `{"commands": [], "metrics": [{"unit": "Cel", "data_type": "number", "name": "m"}], "kind": "k"}`,
		"an announced digest":      `{"kind": "k", "kind_digest": "sha256:` + strings.Repeat("0", 64) + `", "metrics": [{"name": "m", "data_type": "number", "unit": "Cel"}], "commands": []}`,
	}
	for name, raw := range same {
		if got := digest(raw); got != want {
			t.Errorf("%s: digest %s, want the base digest %s", name, got, want)
		}
	}
	changed := map[string]string{
		"a unit":        `{"kind": "k", "metrics": [{"name": "m", "data_type": "number", "unit": "degC"}], "commands": []}`,
		"a metric name": `{"kind": "k", "metrics": [{"name": "n", "data_type": "number", "unit": "Cel"}], "commands": []}`,
		"a kind name":   `{"kind": "j", "metrics": [{"name": "m", "data_type": "number", "unit": "Cel"}], "commands": []}`,
		"a command":     `{"kind": "k", "metrics": [{"name": "m", "data_type": "number", "unit": "Cel"}], "commands": [{"name": "go", "args_schema": {}}]}`,
	}
	for name, raw := range changed {
		if got := digest(raw); got == want {
			t.Errorf("changing %s left the digest at %s", name, got)
		}
	}
}

func TestTheKindDigestOfTheCommittedKindFixtureIsTheKnownAnswer(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join(testsupport.RepositoryRoot(), "contracts", "plant-wire", "fixtures", "valid", "KIND.json"))
	if err != nil {
		t.Fatal(err)
	}
	files, summaries, err := ParseKindFiles([]KindSource{{Path: "KIND.json", Raw: raw}})
	if err != nil || len(summaries) != 1 {
		t.Fatalf("ParseKindFiles = %v, %v, want the fixture accepted as a kind file", summaries, err)
	}
	const golden = "sha256:b5e93cadac03ad9537622cacf3e529d44de2271811d35455712b7eefb3f24fb1"
	if got, err := KindDigest(files["pump"]); err != nil || got != golden || summaries[0].KindDigest != golden {
		t.Fatalf("kind digest of valid/KIND.json = %s (summary %s), %v, want %s", got, summaries[0].KindDigest, err, golden)
	}
}

func TestAKindDigestRefusesWhatHasNoCanonicalForm(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		value any
	}{
		{"a number beyond the range of a double", json.Number("1e999")},
		{"a number that is no number", json.Number("one")},
		{"a value no JSON document holds", make(chan int)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			digest, err := KindDigest(map[string]any{"kind": "k", "metrics": []any{}, "commands": []any{tc.value}})
			if digest != "" || !errors.Is(err, ErrNotCanonical) {
				t.Fatalf("KindDigest = %q, %v, want ErrNotCanonical", digest, err)
			}
		})
	}
}

func TestAKindWithoutACanonicalFormIsNeverDeclaredWellDigested(t *testing.T) {
	t.Parallel()
	unencodable := map[string]any{"kind": "k", "kind_digest": "sha256:" + strings.Repeat("0", 64), "metrics": []any{}, "commands": []any{make(chan int)}}
	if !KindsMisdigested([]any{unencodable}) {
		t.Fatal("a kind with no canonical form passed the digest check")
	}
	if _, err := WithKindDigests([]any{unencodable}); !errors.Is(err, ErrNotCanonical) {
		t.Fatalf("WithKindDigests = %v, want ErrNotCanonical", err)
	}
}

func TestTheKindDigestsOfTheContractFixtureAreItsOwn(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join(testsupport.RepositoryRoot(), "contracts", "plant-wire", "fixtures", "valid", "NBIRTH.json"))
	if err != nil {
		t.Fatal(err)
	}
	birth, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	kinds := birth["kinds"].([]any)
	if KindsMisdigested(kinds) {
		t.Fatal("the fixture's kind_digest differs from the one computed here")
	}
	kinds[0].(map[string]any)["commands"] = []any{}
	if !KindsMisdigested(kinds) {
		t.Fatal("a changed kind kept its digest")
	}
}

func TestANodeBirthCarriesTheDigestOfEachKindItSendsAfterTheWire(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.PumpBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	kinds := []*model.Kind{&spec.Spec.Kinds[0]}
	birth, err := NodeBirth(NodeIdentity{Environment: "simulated", ProtocolMax: 1, PriorLastSeq: -1}, kinds, nil)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(stamped(t, birth)[0])
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if sent := parsed["kinds"].([]any); len(sent) != 1 || KindsMisdigested(sent) {
		t.Fatalf("kinds sent with wrong digests: %v", sent)
	}
}

func TestAKindFileHasNoKindDigestAndIsCheckedAgainstTheKindSchema(t *testing.T) {
	t.Parallel()
	entry := map[string]any{"kind": "k", "metrics": []any{}, "commands": []any{}}
	if err := CheckKind(entry); err != nil {
		t.Fatal(err)
	}
	if err := CheckKind(map[string]any{"kind": "k"}); err == nil || !strings.Contains(err.Error(), "kind file fails its schema") {
		t.Fatalf("a kind file without commands: %v", err)
	}
	unparsable := map[string]any{"kind": "k", "metrics": []any{}, "commands": []any{}, "extra": json.Number("1e999")}
	var invalid *model.ValidationError
	if err := CheckKind(unparsable); !errors.As(err, &invalid) || !errors.Is(err, ErrNotCanonical) || invalid.Contract != model.ContractKindFile {
		t.Fatalf("a kind file with a number beyond the range of a double: %v, want a kind-file validation error caused by ErrNotCanonical", err)
	}
	if err := Check(map[string]any{"type": "KIND"}); err == nil || !strings.Contains(err.Error(), "unknown record type") {
		t.Fatalf("the kind file schema was taken for a record type: %v", err)
	}
}
