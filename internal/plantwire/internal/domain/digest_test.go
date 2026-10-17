package domain

import (
	"encoding/json"
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
	if got := KindDigest(entry); got != "sha256:540c36216dd6364d21e22923590e1cfc5791f7a06c51774cd509214b0451a177" {
		t.Fatalf("kind digest %s", got)
	}
	entry["kind_digest"] = "sha256:" + strings.Repeat("0", 64)
	if got := KindDigest(entry); got != KindDigest(map[string]any{"kind": "k", "metrics": []any{}, "commands": []any{}}) {
		t.Fatalf("the digest covers its own member: %s", got)
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
	raw, err := json.Marshal(stamped(t, NodeBirth(NodeIdentity{Environment: "simulated", ProtocolMax: 1, PriorLastSeq: -1}, kinds, nil))[0])
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
	if err := Check(map[string]any{"type": "KIND"}); err == nil || !strings.Contains(err.Error(), "unknown record type") {
		t.Fatalf("the kind file schema was taken for a record type: %v", err)
	}
}
