package app

import (
	"context"
	"errors"
	"io"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

type failingField struct {
	err   error
	added int
}

func (f *failingField) Add(map[string]any) error {
	f.added++
	return f.err
}

func (f *failingField) Flush() error { return nil }

func (f *failingField) Close() error { return nil }

func (f *failingField) Digest() string { return "sha256:" + strings.Repeat("a", 64) }

func TestAFieldTraceThatCannotTakeARecordStopsRecordingAndLeavesTheRunUncovered(t *testing.T) {
	t.Parallel()
	for name, cause := range map[string]error{"full": run.ErrFieldTraceFull, "too long a line": run.ErrFieldTrace} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			field := &failingField{err: cause}
			rec := recording{field: field}
			line := domain.Traced{Direction: domain.TraceGateway, Record: map[string]any{"type": "DDATA"}, Field: true}
			if err := rec.note(line); err != nil || !rec.incomplete {
				t.Fatalf("note = %v incomplete %v, want the relay kept and the trace marked incomplete", err, rec.incomplete)
			}
			if err := rec.note(line); err != nil || field.added != 1 {
				t.Fatalf("second note = %v after %d adds, want nothing more recorded", err, field.added)
			}
			if rec.digest() != "" {
				t.Fatalf("digest = %q, want none for a trace that is not whole", rec.digest())
			}
		})
	}
}

func TestAFieldTraceThatFailsToWriteStillEndsTheRelay(t *testing.T) {
	t.Parallel()
	disk := errors.New("disk gone")
	rec := recording{field: &failingField{err: disk}}
	err := rec.note(domain.Traced{Direction: domain.TraceGateway, Record: map[string]any{"type": "DDATA"}, Field: true})
	if !errors.Is(err, disk) || rec.incomplete {
		t.Fatalf("note = %v incomplete %v, want the write failure reported", err, rec.incomplete)
	}
}

func TestOnlyARecordMarkedForTheFieldTraceIsAddedToIt(t *testing.T) {
	t.Parallel()
	field := &failingField{}
	rec := recording{field: field}
	for _, line := range []domain.Traced{{Direction: domain.TraceGateway}, {Direction: domain.TraceNode, Field: true}} {
		if err := rec.note(line); err != nil {
			t.Fatal(err)
		}
	}
	if field.added != 1 {
		t.Fatalf("adds = %d, want only the record marked for the field trace", field.added)
	}
}

func TestARelayThatCannotTraceTheGatewayAnyMoreDisconnectsItAndRefusesTheNext(t *testing.T) {
	t.Parallel()
	node := relayGateway(t, "s-1")
	birth, err := node.Birth()
	if err != nil {
		t.Fatal(err)
	}
	in, err := domain.InboundOf(birth)
	if err != nil {
		t.Fatal(err)
	}
	relay := domain.NewRelay(relayTick.Milliseconds())
	relay.GatewayAttached()
	client, server := net.Pipe()
	t.Cleanup(func() { _ = server.Close() })
	loop := &relayLoop{relay: relay, rec: recording{field: &failingField{err: run.ErrFieldTraceFull}}, gateway: uds.NewAccepted(client, time.Second)}
	if err := loop.apply(relay.FromGateway(in)); err != nil {
		t.Fatalf("apply = %v, want the relay to go on", err)
	}
	if loop.gateway != nil || relay.GatewayAdmitted() || !loop.rec.incomplete {
		t.Fatalf("gateway %v admitted %v incomplete %v, want the gateway disconnected and the trace incomplete", loop.gateway, relay.GatewayAdmitted(), loop.rec.incomplete)
	}
	later, laterServer := net.Pipe()
	t.Cleanup(func() { _ = laterServer.Close() })
	if err := loop.attach(uds.NewAccepted(later, time.Second)); err != nil || loop.gateway != nil {
		t.Fatalf("attach = %v gateway %v, want a new gateway turned away", err, loop.gateway)
	}
	if _, err := laterServer.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
		t.Fatalf("read on the turned-away gateway = %v, want it hung up", err)
	}
}

func TestEveryRecordTheLinkAcceptsFitsAFieldTraceLine(t *testing.T) {
	t.Parallel()
	numbers := strings.TrimSuffix(strings.Repeat("1E20,", (domain.MaxBirthBytes-200)/5), ",")
	raw := `{"type":"DDATA","seq":1,"metrics":[` + numbers + `]}`
	in, ok := domain.Decode([]byte(raw))
	if !ok || len(raw) > domain.MaxBirthBytes {
		t.Fatalf("decoded %v for %d bytes, want a record the link accepts", ok, len(raw))
	}
	trace, err := run.NewFieldTrace(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := trace.Add(in.Record); err != nil {
		t.Fatalf("Add of the largest record the link accepts = %v, want it to fit", err)
	}
	if err := trace.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestAnInterruptHidesOnlyTheErrorsItCaused(t *testing.T) {
	t.Parallel()
	broken := errors.New("plantwire: send: the host went away")
	cases := map[string]struct {
		stopped bool
		served  error
		want    error
	}{
		"a cancellation after an interrupt":      {true, context.Canceled, nil},
		"a lost host caused by the interrupt":    {true, errors.Join(ErrHostLost, context.Canceled), nil},
		"a closed connection after an interrupt": {true, net.ErrClosed, nil},
		"a real error after an interrupt":        {true, broken, broken},
		"a host gap after an interrupt":          {true, ErrHostGap, ErrHostGap},
		"a cancellation nobody asked for":        {false, context.Canceled, context.Canceled},
		"nothing":                                {true, nil, nil},
	}
	for name, tc := range cases {
		if got := withoutInterrupt(tc.stopped, tc.served); !errors.Is(got, tc.want) || (tc.want == nil && got != nil) {
			t.Errorf("%s: withoutInterrupt = %v, want %v", name, got, tc.want)
		}
	}
}

func TestASimulatedNodeThatOverflowsItsMailboxFailsTheRelayInsteadOfEndingQuietly(t *testing.T) {
	t.Parallel()
	node := &simNode{}
	node.complete(uds.ErrSimOverflow)
	if !errors.Is(node.failure, uds.ErrSimOverflow) {
		t.Fatalf("failure = %v, want the overflow kept as a failure", node.failure)
	}
}

func TestADeadSimulatedNodeIsReapedOnceAndItsDevicesAreReportedAtOnce(t *testing.T) {
	t.Parallel()
	done := make(chan error, 1)
	node := &simNode{done: done}
	if node.reap() {
		t.Fatal("a node still running was reaped")
	}
	done <- uds.ErrSimOverflow
	if !node.reap() || node.reap() {
		t.Fatal("a node that ended must be reaped exactly once")
	}
}

func TestThePlanOfASessionFindsTheKindFileDigestOfEveryRealInstance(t *testing.T) {
	t.Parallel()
	if got := checkedByKind([]model.ImplementsChecked{{Kind: "pump", KindDigest: "sha256:p"}})["pump"].KindDigest; got != "sha256:p" {
		t.Fatalf("digest of pump = %q, want the one checked", got)
	}
	check, err := CheckConnect(connectOptions(t, planText("/tmp/h.sock", "/tmp/g.sock", "zone-01", "uno-07", "zone-01"), ""))
	if err != nil {
		t.Fatal(err)
	}
	want, err := domain.KindDigest(thermalKindFile(t, "thermal_sensor.kind.json"))
	for _, instance := range check.Instances {
		if instance.Instance == "zone-01" && (err != nil || instance.KindDigest != want) {
			t.Fatalf("kind_digest of zone-01 = %q, want the digest of its kind file %q (%v)", instance.KindDigest, want, err)
		}
	}
}

func TestAPlanWhoseSocketsAreNotPlainAbsolutePathsIsRefusedBeforeAnythingStarts(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "c-1")
	for name, plan := range map[string]string{
		"a parent component": planText("/tmp/x/../h.sock", ""),
		"the same socket":    planText("/tmp/h.sock", "/tmp//h.sock"),
	} {
		_, err := Connect(t.Context(), connectOptions(t, plan, out))
		if !errors.Is(err, ErrPlanRefused) && !isPlanViolation(err) {
			t.Errorf("%s: Connect = %v, want the plan refused", name, err)
		}
	}
}

func isPlanViolation(err error) bool {
	invalid, ok := model.AsValidationError(err)
	return ok && invalid.Contract == model.ContractSessionPlan
}
