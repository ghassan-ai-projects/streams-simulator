package app

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestKindValidateReportsABoardMappingWithItsBoardDigest(t *testing.T) {
	t.Parallel()
	board := filepath.Join(testsupport.ThermalBoardKindsDir(), "arduino-thermal.board.json")
	fan := filepath.Join(testsupport.ThermalBoardKindsDir(), "fan.kind.json")
	doc := invoke(t, "kind", "validate", fan, board).mustSucceed(t).document(t, "kind_validate")
	boards, _ := doc["boards"].([]any)
	if len(boards) != 1 || len(doc["kinds"].([]any)) != 1 {
		t.Fatalf("kind validate = %v, want the fan kind file and the board mapping", doc)
	}
	mapping, _ := boards[0].(map[string]any)
	digest, _ := mapping["board_digest"].(string)
	if mapping["board_type"] != "arduino-thermal" || !strings.HasPrefix(digest, "sha256:") || mapping["path"] != board {
		t.Fatalf("board = %v, want arduino-thermal with its board_digest", mapping)
	}
	alone := invoke(t, "kind", "validate", board).mustSucceed(t).document(t, "kind_validate")
	if kinds, _ := alone["kinds"].([]any); kinds == nil || len(kinds) != 0 {
		t.Fatalf("a board mapping alone = %v, want an empty kinds list beside it", alone)
	}
	broken := writeFile(t, "broken.board.json", `{"board_type": "x", "kinds": {}, "extra": 1}`)
	if got := invoke(t, "kind", "validate", broken); got.code != 1 || got.failure(t)["code"] != "kind_failed" {
		t.Fatalf("a broken board mapping = %d %v, want kind_failed", got.code, got.failure(t))
	}
}
