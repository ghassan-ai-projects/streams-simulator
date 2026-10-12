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

func stamped(t *testing.T, records ...Record) []map[string]any {
	t.Helper()
	out := NewTransmitter("s-1")
	for _, record := range records {
		out.Queue(record)
	}
	drained := out.Drain()
	for _, record := range drained {
		raw, err := json.Marshal(record)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Parse(raw); err != nil {
			t.Fatalf("%v: %v", record["type"], err)
		}
	}
	return drained
}

func TestBirthsOfEveryDomainMatchTheContract(t *testing.T) {
	t.Parallel()
	for _, path := range []string{testsupport.StructuredDomain(), testsupport.UtilityDomain(), testsupport.Domain("aquaculture-pond")} {
		spec, err := domain.Load(path)
		if err != nil {
			t.Fatal(err)
		}
		kinds := make([]*model.Kind, len(spec.Spec.Kinds))
		for i := range spec.Spec.Kinds {
			kinds[i] = &spec.Spec.Kinds[i]
		}
		records := []Record{NodeBirth(NodeIdentity{Environment: "simulated", ProtocolMax: 1, ContractVersion: "1.0.0", PriorLastSeq: -1}, kinds, spec.Spec.Boards)}
		for set := range spec.Spec.Instances {
			effective := spec.InstanceKind(set)
			id := DeviceIdentity{Device: "dev-" + effective.Name, Kind: effective.Name, Board: "board-1", BootID: "board-1-boot-1"}
			records = append(records, DeviceBirth(id, spec.Kind(effective.Name), effective, nil))
		}
		stamped(t, records...)
	}
}

func TestTheContractDigestIsThePinnedOne(t *testing.T) {
	t.Parallel()
	pinned, err := os.ReadFile(filepath.Join(testsupport.RepositoryRoot(), "contracts", "plant-wire", "DIGEST"))
	if err != nil {
		t.Fatal(err)
	}
	if got := ContractDigest(); got != strings.TrimSpace(string(pinned)) {
		t.Fatalf("contract digest %s, pinned %s", got, pinned)
	}
}

func TestTheCapabilityDigestMatchesTheContractFixture(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join(testsupport.RepositoryRoot(), "contracts", "plant-wire", "fixtures", "valid", "NBIRTH.json"))
	if err != nil {
		t.Fatal(err)
	}
	birth, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := CapabilityDigest(birth["kinds"].([]any)); got != birth["capability_digest"] {
		t.Fatalf("capability digest %s, fixture %v", got, birth["capability_digest"])
	}
}

func TestOverridesAddAndNarrowOnlyWhatTheInstanceChanges(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.StructuredDomain())
	if err != nil {
		t.Fatal(err)
	}
	named := spec.InstanceKind(spec.InstanceSetFor("T-main"))
	overrides := overridesOf(spec.Kind(named.Name), named)
	if len(overrides["add"].([]any)) != 1 || overrides["narrow"] != nil {
		t.Fatalf("overrides of T-main = %v, want its one added channel", overrides)
	}
	plain := spec.InstanceKind(spec.InstanceSetFor("tank-1"))
	if got := overridesOf(spec.Kind(plain.Name), plain); len(got) != 0 {
		t.Fatalf("overrides of tank-1 = %v, want none", got)
	}
}
