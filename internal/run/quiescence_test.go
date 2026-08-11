package run

// Slice B (G5): the quiescence barrier is fully deterministic. The deadline
// comes from an injectable clock, the wait is request-cancellable, a stale
// report cannot satisfy a later wait, and a timed-out advance is logged and
// replayed as incomplete — never silently successful. No test here sleeps.

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// fakeQuiescenceClock fires deadlines on demand.
type fakeQuiescenceClock struct {
	mu     sync.Mutex
	timers []*fakeQuiescenceTimer
}

func (f *fakeQuiescenceClock) NewTimer(time.Duration) QuiescenceTimer {
	t := &fakeQuiescenceTimer{ch: make(chan time.Time, 1)}
	f.mu.Lock()
	f.timers = append(f.timers, t)
	f.mu.Unlock()
	return t
}

func (f *fakeQuiescenceClock) fire() {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, t := range f.timers {
		t.fire()
	}
}

type fakeQuiescenceTimer struct {
	ch   chan time.Time
	once sync.Once
}

func (t *fakeQuiescenceTimer) C() <-chan time.Time { return t.ch }
func (t *fakeQuiescenceTimer) Stop() bool          { return true }

func (t *fakeQuiescenceTimer) fire() {
	t.once.Do(func() { t.ch <- time.Now() })
}

// quiesceHarness builds a run with a fake clock and a park hook; Advance
// runs in the background and parks deterministically.
func quiesceHarness(t *testing.T, start, to int64) (*Run, *fakeQuiescenceClock, chan error, chan struct{}) {
	t.Helper()
	spec, a := testBase(t)
	fc := &fakeQuiescenceClock{}
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 77, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, QuiescenceClock: fc,
	})
	if err != nil {
		t.Fatal(err)
	}
	parked := make(chan struct{})
	r.SetQuiesceParkedHook(func() { parked <- struct{}{} })
	done := make(chan error, 1)
	go func() {
		_, err := r.Advance(context.Background(), to, true)
		done <- err
	}()
	<-parked // the waiter is blocked on its deadline, not polling
	return r, fc, done, parked
}

func TestQuiescenceTimeoutMarksRunIncomplete(t *testing.T) {
	start := model.DefaultStartTimeNS + 4*3600*1e9
	to := start + int64(time.Hour)
	r, fc, done, _ := quiesceHarness(t, start, to)
	fc.fire()
	err := <-done
	if !errors.Is(err, ErrConsumerNotQuiesced) {
		t.Fatalf("expected consumer_not_quiesced, got %v", err)
	}
	if r.Reproducible() {
		t.Fatal("a timed-out await must mark the run non-reproducible")
	}
	if len(r.commandLog) == 0 || r.commandLog[len(r.commandLog)-1].Op != model.OpClockAdvance {
		t.Fatal("the timed-out advance must still be logged for replay")
	}
	if !r.commandLog[len(r.commandLog)-1].Args["await_consumer"].(bool) {
		t.Fatal("logged advance must retain the await flag")
	}
	// The artifact records the incomplete state. End returns the artifact
	// alongside the retained failure for an aborted run.
	dir := t.TempDir()
	art, endErr := r.End(dir)
	if art == nil || !art.Incomplete {
		t.Fatalf("timed-out run must be recorded incomplete: art=%v endErr=%v", art, endErr)
	}
}

func TestQuiescenceCancelStopsTheWait(t *testing.T) {
	start := model.DefaultStartTimeNS + 4*3600*1e9
	to := start + int64(time.Hour)
	spec, a := testBase(t)
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 77, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, QuiescenceClock: &fakeQuiescenceClock{},
	})
	if err != nil {
		t.Fatal(err)
	}
	parked := make(chan struct{})
	r.SetQuiesceParkedHook(func() { parked <- struct{}{} })
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := r.Advance(ctx, to, true)
		done <- err
	}()
	<-parked
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	// A canceled wait is a caller-side abandonment, not a simulator
	// failure: the run is not marked incomplete.
	if r.Reproducible() == false {
		t.Fatal("ctx cancel must not mark the run non-reproducible")
	}
}

func TestQuiescenceStaleReportCannotSatisfyLaterWait(t *testing.T) {
	start := model.DefaultStartTimeNS + 4*3600*1e9
	to := start + int64(time.Hour)
	r, fc, done, parked := quiesceHarness(t, start, to)
	// A report for an earlier instant wakes the waiter but must not satisfy
	// the later target: the waiter re-parks.
	r.ReportQuiesced(to - int64(time.Minute))
	select {
	case <-parked:
	case <-done:
		t.Fatal("stale report satisfied the wait")
	case <-time.After(5 * time.Second):
		t.Fatal("waiter did not re-park after a stale report")
	}
	fc.fire()
	if err := <-done; !errors.Is(err, ErrConsumerNotQuiesced) {
		t.Fatalf("expected timeout after stale report, got %v", err)
	}
}

func TestQuiescenceReportAtTargetSatisfies(t *testing.T) {
	start := model.DefaultStartTimeNS + 4*3600*1e9
	to := start + int64(time.Hour)
	r, _, done, _ := quiesceHarness(t, start, to)
	r.ReportQuiesced(to)
	if err := <-done; err != nil {
		t.Fatalf("report at target must satisfy the wait: %v", err)
	}
}

func TestQuiescenceFastPathHonorsReportBetweenAdvances(t *testing.T) {
	start := model.DefaultStartTimeNS + 4*3600*1e9
	spec, a := testBase(t)
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 77, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, QuiescenceClock: &fakeQuiescenceClock{},
	})
	if err != nil {
		t.Fatal(err)
	}
	parked := make(chan struct{}, 4)
	r.SetQuiesceParkedHook(func() { parked <- struct{}{} })

	// The consumer reported quiescence through T2 before the harness issued
	// the advance: the wait returns immediately without blocking.
	t2 := start + int64(time.Hour)
	r.ReportQuiesced(t2)
	if _, err := r.Advance(context.Background(), t2, true); err != nil {
		t.Fatalf("fast path failed: %v", err)
	}
	select {
	case <-parked:
		t.Fatal("fast path must not park")
	default:
	}

	// A later target beyond the reported watermark blocks again.
	t3 := t2 + int64(time.Hour)
	done := make(chan error, 1)
	go func() {
		_, err := r.Advance(context.Background(), t3, true)
		done <- err
	}()
	select {
	case <-parked:
	case <-time.After(5 * time.Second):
		t.Fatal("waiter did not park for a target beyond the watermark")
	}
	r.ReportQuiesced(t3)
	if err := <-done; err != nil {
		t.Fatalf("report at the new target failed: %v", err)
	}
}

// TestReplayOfTimedOutAdvanceReportsIncomplete: a run whose advance timed
// out reproduces byte-for-byte, but replay reports the original run as
// incomplete — never silent success.
func TestReplayOfTimedOutAdvanceReportsIncomplete(t *testing.T) {
	start := model.DefaultStartTimeNS + 4*3600*1e9
	to := start + int64(time.Hour)
	spec, a := testBase(t)
	fc := &fakeQuiescenceClock{}
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 77, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, QuiescenceClock: fc,
	})
	if err != nil {
		t.Fatal(err)
	}
	parked := make(chan struct{})
	r.SetQuiesceParkedHook(func() { parked <- struct{}{} })
	done := make(chan error, 1)
	go func() {
		_, err := r.Advance(context.Background(), to, true)
		done <- err
	}()
	// No consumer reports: fire the injected deadline deterministically. The
	// park signal guarantees the waiter created its deadline first.
	<-parked
	fc.fire()
	if err := <-done; !errors.Is(err, ErrConsumerNotQuiesced) {
		t.Fatal(err)
	}
	dir := t.TempDir()
	art, _ := r.End(dir)
	if art == nil || !art.Incomplete {
		t.Fatalf("artifact must be incomplete: art=%v", art)
	}
	loaded, err := LoadArtifact(filepath.Join(dir, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := ReplayArtifact(context.Background(), loaded, spec, a, "")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Matches {
		t.Fatalf("timed-out run must still replay byte-identically: %s vs %s", res.GotDigest, res.WantDigest)
	}
	if !res.Incomplete {
		t.Fatal("replay must report the original run as incomplete")
	}
	if res.Detail == "" {
		t.Fatal("replay must surface the original failure detail")
	}
}
