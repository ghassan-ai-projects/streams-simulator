package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestComposeCopiesTheKindFileAndBoardMappingAnEntryImplementsOntoItsKind(t *testing.T) {
	t.Parallel()
	doc := composeOK(t, baseLibrary(t), withEquipment(t, kilnEntry(map[string]any{"implements": "kiln_face", "board_type": "kiln-plc"})))
	kind := member(t, doc, "kinds", 0).(map[string]any)
	if kind["implements"] != "kiln_face" || kind["board_type"] != "kiln-plc" {
		t.Fatalf("kind implements %v through %v, want kiln_face through kiln-plc", kind["implements"], kind["board_type"])
	}
}

func TestComposeLeavesAKindWithoutAnInterfaceWhenItsEntryNamesNone(t *testing.T) {
	t.Parallel()
	cases := map[string]map[string]any{
		"no interface":            nil,
		"a kind file, no mapping": {"implements": "kiln_face"},
	}
	for name, extra := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			kind := member(t, composeOK(t, baseLibrary(t), withEquipment(t, kilnEntry(extra))), "kinds", 0).(map[string]any)
			if _, mapped := kind["board_type"]; mapped || kind["implements"] != extra["implements"] {
				t.Fatalf("kind implements %v through %v, want %v and no board mapping", kind["implements"], kind["board_type"], extra["implements"])
			}
		})
	}
}

func TestComposeRefusesABoardTypeWithoutTheKindFileItImplements(t *testing.T) {
	t.Parallel()
	_, err := compose(t, baseLibrary(t), withEquipment(t, kilnEntry(map[string]any{"board_type": "kiln-plc"})))
	wantIssue(t, err, model.ContractComposition, "/equipment/0/implements", "missing required property")
}

func TestComposeRefusesAnInterfaceNameThatIsNotAnID(t *testing.T) {
	t.Parallel()
	for _, member := range []string{"implements", "board_type"} {
		t.Run(member, func(t *testing.T) {
			t.Parallel()
			entry := kilnEntry(map[string]any{"implements": "kiln_face", member: "a kiln"})
			_, err := compose(t, baseLibrary(t), withEquipment(t, entry))
			wantIssue(t, err, model.ContractComposition, "/equipment/0/"+member, "")
		})
	}
}
