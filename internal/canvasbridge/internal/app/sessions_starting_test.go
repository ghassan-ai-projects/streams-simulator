package app

import (
	"context"
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func TestASecondStartWhileTheFirstIsStillStartingIsRefused(t *testing.T) {
	t.Parallel()
	rig := newSessionRig(t)
	rig.launched, rig.proceed = make(chan struct{}), make(chan struct{})
	first := rig.later("session.start", plan())
	<-rig.launched
	_, err := rig.bridge.RunAction(context.Background(), "session.start", plan())
	var failure *domain.Failure
	if !errors.As(err, &failure) || failure.Code != "session_running" || rig.children() != 0 {
		t.Fatalf("a second start while the first starts = %v after %d children, want session_running and no second child", err, rig.children())
	}
	close(rig.proceed)
	if started := <-first; started.State != domain.SessionRunning || started.Run != "c-1" || rig.children() != 1 {
		t.Fatalf("the first start = %+v after %d children, want it running alone as c-1", started, rig.children())
	}
}
