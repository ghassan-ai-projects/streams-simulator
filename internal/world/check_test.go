package world_test

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// The facade's checks judge a command without carrying it out.
func TestTheFacadeChecksCommandsWithoutActing(t *testing.T) {
	t.Parallel()
	w := newPondWorld(t, world.Options{})
	pond := w.EntityIDs()[0]
	if err := w.CheckEffectorArgs("start_aerator", map[string]any{"pond_id": pond}); err != nil {
		t.Fatalf("valid effector args refused: %v", err)
	}
	if err := w.CheckEffectorArgs("start_aerator", map[string]any{}); err == nil || !strings.Contains(err.Error(), `"pond_id"`) {
		t.Fatalf("effector args without the required pond_id: %v", err)
	}
	if err := w.CheckFaultParams("aerator_failure", map[string]any{"severity": 0.5}); err != nil {
		t.Fatalf("declared fault severity refused: %v", err)
	}
	if err := w.CheckFaultParams("aerator_failure", map[string]any{"bogus": 1}); err == nil || !strings.Contains(err.Error(), `no parameter "bogus"`) {
		t.Fatalf("an undeclared fault parameter: %v", err)
	}
	if len(w.EffectorCalls()) != 0 || w.ActiveFaultsCount() != 0 {
		t.Fatal("a check acted on the world")
	}
	if !world.KnownFailureMode(world.ModeSilentNoEffect) || world.KnownFailureMode("bogus_mode") {
		t.Fatal("KnownFailureMode does not follow the vocabulary")
	}
}
