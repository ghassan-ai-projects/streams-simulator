package domain

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func marshalCanonical(t *testing.T, value any) ([]byte, error) {
	t.Helper()
	raw, err := canonical.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw, nil
}

const fanMapping = `{"board_type": "%s", "kinds": {"fan": {"commands": {"stop_fan": {"raw": "stop_fan"}}, "metrics": {}}}}`

func TestABoardMappingIsRefusedAsAWholeWhenItCannotBeRead(t *testing.T) {
	t.Parallel()
	unreadable := errors.New("unreadable")
	cases := []struct {
		name   string
		source KindSource
		want   error
	}{
		{"read error", KindSource{Path: "a.board.json", Err: unreadable}, unreadable},
		{"member twice", KindSource{Path: "a.board.json", Raw: []byte(`{"board_type": "a", "board_type": "b", "kinds": {}}`)}, ErrDuplicateMember},
		{"number beyond a double", KindSource{Path: "a.board.json",
			Raw: []byte(`{"board_type": "a", "kinds": {"k": {"commands": {"c": {"raw": "r", "args": {"x": {"value": 1e999}}}}, "metrics": {}}}}`)}, errBoardSchema},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, _, err := ParseBoardFiles([]KindSource{c.source})
			if errors.Is(c.want, errBoardSchema) && strings.Contains(err.Error(), "board mapping fails its schema") {
				return
			}
			if !errors.Is(err, c.want) {
				t.Fatalf("ParseBoardFiles = %v, want %v", err, c.want)
			}
		})
	}
}

var errBoardSchema = errors.New("board schema")

func TestTwoBoardTypesMayMapOneKindButOneBoardTypeIsMappedOnce(t *testing.T) {
	t.Parallel()
	first := KindSource{Path: "a.board.json", Raw: []byte(strings.Replace(fanMapping, "%s", "a", 1))}
	second := KindSource{Path: "b.board.json", Raw: []byte(strings.Replace(fanMapping, "%s", "b", 1))}
	both, _, err := ParseBoardFiles([]KindSource{first, second})
	if err != nil || len(both["a"].Kinds) != 1 || len(both["b"].Kinds) != 1 {
		t.Fatalf("the fan on board types a and b: %v, %v, want both mappings, each read by its board type", both, err)
	}
	led := KindSource{Path: "a2.board.json", Raw: []byte(`{"board_type": "a", "kinds": {"led": {"commands": {}, "metrics": {}}}}`)}
	if _, _, err := ParseBoardFiles([]KindSource{first, led}); !errors.Is(err, ErrBoardTypeDuplicate) {
		t.Fatalf("two mappings of board type a: %v, want ErrBoardTypeDuplicate", err)
	}
	files, summaries, err := ParseBoardFiles([]KindSource{first})
	if err != nil || len(files) != 1 || summaries[0].BoardType != "a" || !strings.HasPrefix(summaries[0].BoardDigest, "sha256:") {
		t.Fatalf("one mapping of fan: %v, %+v, %v", files, summaries, err)
	}
}

func TestASchemaFailureNamesTheBoardFileAndItsContract(t *testing.T) {
	t.Parallel()
	_, err := ParseBoardMapping(KindSource{Path: "x.board.json", Raw: []byte(`{"board_type": "x"}`)})
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || invalid.Source != "x.board.json" || invalid.Contract != model.ContractBoardFile ||
		!strings.Contains(err.Error(), "board mapping fails its schema") {
		t.Fatalf("a mapping without kinds: %v", err)
	}
	if code, _ := model.FailureCodeOf(err); code != "kind_failed" {
		t.Errorf("a board mapping failure has code %q, want kind_failed", code)
	}
}

func TestTheBoardSchemaIsNotARecordType(t *testing.T) {
	t.Parallel()
	if err := Check(map[string]any{"type": "BOARD", "board_type": "b", "kinds": map[string]any{}}); err == nil || !strings.Contains(err.Error(), "unknown record type") {
		t.Fatalf("the board mapping schema was taken for a record type: %v", err)
	}
}

func TestTheContractVersionIsTheVersionFileOfTheContract(t *testing.T) {
	t.Parallel()
	folder := os.DirFS(filepath.Join(testsupport.RepositoryRoot(), "contracts", "plant-wire"))
	raw, err := fs.ReadFile(folder, "VERSION")
	if err != nil || string(raw) != ContractVersion+"\n" {
		t.Fatalf("contracts/plant-wire/VERSION holds %q (%v), this binary speaks %s", raw, err, ContractVersion)
	}
}
