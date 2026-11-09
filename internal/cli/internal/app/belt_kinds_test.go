package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func beltSessionPlan() string {
	return filepath.Join(testsupport.BeltConveyorStoriesDir(), "sessions", "s2-b.session-plan.json")
}

func TestDomainComposeOfTheConveyorFindsItsKindFilesAndBoardMappingInTheKindsDirectory(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "b.domain.json")
	invoke(t, "domain", "compose", testsupport.BeltConveyorComposition("b"), "--out", out, "--kinds-dir", testsupport.BeltConveyorKindsDir()).
		mustSucceed(t).document(t, "domain_compose")
	written, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	committed, err := os.ReadFile(testsupport.BeltConveyorDomain("b"))
	if err != nil || !bytes.Equal(written, committed) {
		t.Fatalf("the composed plant b differs from the committed example (%v)", err)
	}
	failure := invoke(t, "domain", "compose", testsupport.BeltConveyorComposition("b")).failure(t)
	if issue := firstIssue(failure); failure["code"] != "domain_invalid" || issue["keyword"] != "implements_missing" {
		t.Fatalf("compose without --kinds-dir = %v, want the missing kind files refused", failure)
	}
}

func TestAConnectCheckOfAConveyorSessionPlanChecksTheFeederThroughItsBoardMapping(t *testing.T) {
	t.Parallel()
	args := []string{"plantwire", "connect", "--plan", beltSessionPlan(), "--check", "--domains-dir", filepath.Join(testsupport.RepositoryRoot(), "examples"),
		"--adapters-dir", testsupport.AdaptersDir()}
	doc := invoke(t, append(args, "--kinds-dir", testsupport.BeltConveyorKindsDir())...).mustSucceed(t).document(t, "plantwire_connect_check")
	checked, _ := doc["implements_checked"].([]any)
	boards := map[any]any{}
	for _, entry := range checked {
		fields, _ := entry.(map[string]any)
		boards[fields["kind"]] = fields["board_type"]
	}
	if len(checked) != 3 || boards["bulk_feeder"] != "feeder-controller" || boards["belt_conveyor"] != nil || boards["conveyor_drive"] != nil {
		t.Fatalf("implements_checked = %v, want the feeder through feeder-controller and the belt and drives directly", checked)
	}
	if failure := invoke(t, args...).failure(t); failure["code"] != "domain_invalid" || firstIssue(failure)["keyword"] != "implements_missing" {
		t.Fatalf("check without --kinds-dir = %v, want the missing kind files refused", failure)
	}
}

func firstIssue(failure map[string]any) map[string]any {
	issues, _ := failure["issues"].([]any)
	if len(issues) == 0 {
		return nil
	}
	issue, _ := issues[0].(map[string]any)
	return issue
}
