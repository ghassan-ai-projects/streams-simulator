package app

import (
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const pumpBoard = "pump-board"

func pumpBoardDirector(t *testing.T, check func(*model.DomainSpec) error) *Director {
	t.Helper()
	compiled, err := domain.Load(testsupport.PumpBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	compiled.Spec.Kinds[0].Implements = "pump"
	native, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	d := NewDirector(t.Context(), domain.NewCatalog([]*domain.Compiled{compiled}), map[string]*model.Adapter{"native-jsonl": native}, t.TempDir())
	d.check = check
	return d
}

func TestAWorldOfADomainThatImplementsAKindFileIsRefusedWhenTheDirectorWasGivenNoKindFiles(t *testing.T) {
	t.Parallel()
	_, err := pumpBoardDirector(t, nil).CreateWorld(WorldRequest{Domain: pumpBoard, Seed: 1})
	if toolCode(err) != rules.CodeDomainInvalid || !strings.Contains(err.Error(), `implements kind file "pump"`) || !strings.Contains(err.Error(), "/kinds/0/implements") {
		t.Fatalf("CreateWorld = %v, want domain_invalid naming the kind file and /kinds/0/implements", err)
	}
}

func TestTheInjectedCheckDecidesWhetherAWorldOfAnImplementingDomainIsCreated(t *testing.T) {
	t.Parallel()
	var seen string
	accept := func(spec *model.DomainSpec) error {
		seen = spec.ID
		return nil
	}
	created, err := pumpBoardDirector(t, accept).CreateWorld(WorldRequest{Domain: pumpBoard, Seed: 1})
	if id, _ := created["world_id"].(string); err != nil || id == "" || seen != pumpBoard {
		t.Fatalf("CreateWorld = %v, %v (check saw %q), want a world after the check accepted %s", created, err, seen, pumpBoard)
	}
	drift := errors.New("kind pump differs from its kind file")
	refuse := func(*model.DomainSpec) error {
		return &model.ValidationError{Contract: model.ContractDomainSpec, Summary: drift.Error(), Cause: drift,
			Issues: []model.Issue{model.LocatedIssue("/kinds/0/implements", "differs at /metrics/0/unit")}}
	}
	_, err = pumpBoardDirector(t, refuse).CreateWorld(WorldRequest{Domain: pumpBoard, Seed: 1})
	if toolCode(err) != rules.CodeDomainInvalid || !strings.Contains(err.Error(), "differs at /metrics/0/unit") {
		t.Fatalf("CreateWorld = %v, want domain_invalid carrying the check's issue", err)
	}
}

func TestADomainWithoutImplementsNeedsNoKindFiles(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	created, err := d.CreateWorld(WorldRequest{Domain: "aquaculture-pond", Seed: 1})
	if id, _ := created["world_id"].(string); err != nil || id == "" {
		t.Fatalf("CreateWorld = %v, %v, want a world from a director with no check", created, err)
	}
}
