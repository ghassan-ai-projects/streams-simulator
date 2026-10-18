package plantwire

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func pumpKindText(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(testsupport.PumpKindFile())
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func twoIssues(map[string]any) []model.Issue {
	return []model.Issue{model.RuleIssue("/metrics/0/resolution", "twin_member_required", "resolution", "needs it"),
		model.RuleIssue("/commands/0/safe_stop", "twin_member_required", "safe_stop", "needs it")}
}

func TestValidateKindStreamSummarisesOneDocumentAsStdin(t *testing.T) {
	t.Parallel()
	summary, err := ValidateKindStream(strings.NewReader(pumpKindText(t)), nil)
	if err != nil || summary.Path != "<stdin>" || summary.Kind != "pump" || !strings.HasPrefix(summary.KindDigest, "sha256:") {
		t.Fatalf("ValidateKindStream = %+v, %v, want the pump kind read from <stdin>", summary, err)
	}
}

func TestValidateKindStreamRefusesWhatIsNotAKindDocumentOrIsTooLarge(t *testing.T) {
	t.Parallel()
	_, err := ValidateKindStream(strings.NewReader("{"), nil)
	if invalid, ok := model.AsValidationError(err); !ok || invalid.Source != "<stdin>" || invalid.Issues[0].Keyword != "json" {
		t.Fatalf("a document that is not JSON = %v, want a violation of <stdin> with the keyword json", err)
	}
	_, err = ValidateKindStream(strings.NewReader(pumpKindText(t)+strings.Repeat(" ", 1<<20)), nil)
	if err == nil || !strings.Contains(err.Error(), "<stdin> holds more than 1048576 bytes") {
		t.Fatalf("a document over 1 MiB = %v, want the limit named", err)
	}
}

func TestTwinRulesGivenToTheFacadeRefuseAKindDocumentWithEveryIssue(t *testing.T) {
	t.Parallel()
	_, err := ValidateKindStream(strings.NewReader(pumpKindText(t)), twoIssues)
	invalid, ok := model.AsValidationError(err)
	if !ok || !errors.Is(err, ErrTwinRequirements) || len(invalid.Issues) != 2 || invalid.Issues[1].Member != "safe_stop" {
		t.Fatalf("ValidateKindStream with rules = %v, want ErrTwinRequirements with both issues", err)
	}
	path := filepath.Join(pumpKindsDirectory(t), "pump.kind.json")
	if _, err := ValidateKindFilesWith([]string{path}, twoIssues); !errors.Is(err, ErrTwinRequirements) {
		t.Fatalf("ValidateKindFilesWith with rules = %v, want ErrTwinRequirements", err)
	}
}

func TestValidateKindFilesWithoutRulesAcceptsAFileAndRefusesAMissingOne(t *testing.T) {
	t.Parallel()
	dir := pumpKindsDirectory(t)
	summaries, err := ValidateKindFilesWith([]string{filepath.Join(dir, "pump.kind.json")}, nil)
	if err != nil || len(summaries) != 1 || summaries[0].Kind != "pump" {
		t.Fatalf("ValidateKindFilesWith = %+v, %v, want the pump kind", summaries, err)
	}
	if _, err := ValidateKindFilesWith([]string{filepath.Join(dir, "absent.kind.json")}, nil); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a missing file = %v, want fs.ErrNotExist", err)
	}
}

func TestWithKindFilesHandsTheKindFilesOfADirectoryToTheCaller(t *testing.T) {
	t.Parallel()
	dir := pumpKindsDirectory(t)
	names, err := WithKindFiles(nil, []string{dir}, func(kinds map[string]map[string]any) ([]string, error) {
		found := []string{}
		for name := range kinds {
			found = append(found, name)
		}
		return found, nil
	})
	if err != nil || len(names) != 1 || names[0] != "pump" {
		t.Fatalf("WithKindFiles = %v, %v, want the pump kind", names, err)
	}
	refusal := errors.New("the caller refuses")
	if _, err := WithKindFiles([]string{filepath.Join(dir, "pump.kind.json")}, nil, func(map[string]map[string]any) (int, error) { return 0, refusal }); !errors.Is(err, refusal) {
		t.Fatalf("WithKindFiles = %v, want the caller's refusal", err)
	}
	if _, err := WithKindFiles(nil, []string{filepath.Join(dir, "absent")}, func(map[string]map[string]any) (int, error) { return 1, nil }); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a missing directory = %v, want fs.ErrNotExist", err)
	}
}
