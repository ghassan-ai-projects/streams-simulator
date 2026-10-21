package plantwire

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

func pumpKindsDirectory(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(testsupport.PumpKindFile())
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	testsupport.WriteFileIn(t, dir, "pump.kind.json", raw)
	return dir
}

func TestCheckImplementsAcceptsAMatchingKindFileAndRefusesTheMissingOne(t *testing.T) {
	t.Parallel()
	compiled, err := domain.Load(testsupport.PumpBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	compiled.Spec.Kinds[0].Implements = "pump"
	checked, err := CheckImplements(compiled.Spec, []string{pumpKindsDirectory(t)})
	if err != nil || len(checked) != 1 || checked[0].Kind != "pump" || !strings.HasPrefix(checked[0].KindDigest, "sha256:") {
		t.Fatalf("CheckImplements = %v, %v, want the pump kind accepted with its digest", checked, err)
	}
	var invalid *model.ValidationError
	_, err = CheckImplements(compiled.Spec, nil)
	if !errors.As(err, &invalid) || len(invalid.Issues) != 1 || invalid.Issues[0].Pointer != "/kinds/0/implements" || !errors.Is(err, ErrKindFileMissing) {
		t.Fatalf("CheckImplements without kind files = %v, want a domain validation error at /kinds/0/implements", err)
	}
}

func TestValidateKindFilesReturnsTheKindAndItsDigest(t *testing.T) {
	t.Parallel()
	path := filepath.Join(pumpKindsDirectory(t), "pump.kind.json")
	validated, err := ValidateKindFilesWith([]string{path}, nil)
	summaries := validated.Kinds
	if err != nil || len(summaries) != 1 || summaries[0].Kind != "pump" || summaries[0].Path != path || !strings.HasPrefix(summaries[0].KindDigest, "sha256:") {
		t.Fatalf("ValidateKindFilesWith = %+v, %v, want the pump kind file with its digest", summaries, err)
	}
}
