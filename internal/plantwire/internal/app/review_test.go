package app

import (
	"bytes"
	"errors"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func shippedDomains(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(testsupport.DomainsDir(), "*.domain.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no shipped domains: %v", err)
	}
	return files
}

func TestEveryShippedDomainSpeaksValidWire(t *testing.T) {
	t.Parallel()
	for _, path := range shippedDomains(t) {
		host := &scriptedHost{t: t}
		if err := NewSession(runOf(t, path), host, sliceOptions(3)).Run(t.Context()); err != nil {
			t.Fatalf("%s: %v", filepath.Base(path), err)
		}
	}
}

func TestTheWireCarriesNoFaultFailureModeOrOnsetInAnyShippedDomain(t *testing.T) {
	t.Parallel()
	for _, path := range append(shippedDomains(t), testsupport.PumpBoardDomain()) {
		r := runOf(t, path)
		spec, entity := r.Domain(), r.World.EntityIDs()[0]
		kind := spec.InstanceKind(spec.InstanceSetFor(entity))
		onset := model.DefaultStartTimeNS + 90_123_456_789
		if _, err := r.InjectFault(entity, kind.Faults[0].ID, onset, nil); err != nil {
			t.Fatal(err)
		}
		host := &scriptedHost{t: t}
		if err := NewSession(r, host, sliceOptions(3)).Run(t.Context()); err != nil {
			t.Fatal(err)
		}
		for _, secret := range append(wireSecrets(spec.Spec.Kinds, onset), hiddenFacts(spec.Spec)...) {
			if bytes.Contains(host.transcript, []byte(secret)) {
				t.Errorf("%s: the wire carries %q", filepath.Base(path), secret)
			}
		}
	}
}

func wireSecrets(kinds []model.Kind, onset int64) []string {
	secrets := []string{strconv.FormatInt(onset, 10), "silent_no_effect", "confirmed_no_effect", "ack_lost"}
	for _, kind := range kinds {
		for _, fault := range kind.Faults {
			secrets = append(secrets, `"`+fault.ID+`"`, fault.ID+" ")
		}
	}
	return secrets
}

func TestAFaultChangesNoBirth(t *testing.T) {
	t.Parallel()
	births := func(inject bool) string {
		r := utilityRun(t)
		if inject {
			if _, err := r.InjectFault("meter-1", "load_surge", model.DefaultStartTimeNS+1, nil); err != nil {
				t.Fatal(err)
			}
		}
		host := &scriptedHost{t: t}
		if err := NewSession(r, host, sliceOptions(1)).Run(t.Context()); err != nil {
			t.Fatal(err)
		}
		var out strings.Builder
		for _, line := range strings.Split(string(host.transcript), "\n") {
			if strings.Contains(line, `"type":"NBIRTH"`) || strings.Contains(line, `"type":"DBIRTH"`) {
				out.WriteString(line)
			}
		}
		return out.String()
	}
	if births(false) != births(true) {
		t.Fatal("an injected fault changed a birth")
	}
}

func TestAMalformedCommandIsAnsweredAndTheSessionGoesOn(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			h.replyInvalid(map[string]any{"type": "DCMD", "command_id": "c-bad", "device": "meter-1", "fallback_on_lease_los": "safe_off"})
		}
	}}
	if _, err := session(t, host, 3); err != nil {
		t.Fatal(err)
	}
	for _, record := range host.sent {
		if record["type"] == "DRECEIPT" && record["command_id"] == "c-bad" && record["reason"] == "malformed" {
			return
		}
	}
	t.Fatalf("no malformed receipt in %v", host.types())
}

func TestAGoodbyeThenHangUpEndsCleanlyAndAHangUpAloneIsHostLoss(t *testing.T) {
	t.Parallel()
	goodbye := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) {
		h.silent, h.hangUp = true, io.EOF
		h.reply("STATE", map[string]any{"host_id": "host-a", "online": false, "heartbeat": 1, "heartbeat_interval_ms": 1000, "authority_epoch": 1, "protocol_min": 1, "protocol_max": 1})
	}}
	if _, err := session(t, goodbye, 3); err != nil {
		t.Fatalf("goodbye then hang-up: %v", err)
	}
	dropped := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) { h.silent, h.hangUp = true, io.EOF }}
	if _, err := session(t, dropped, 3); !errors.Is(err, ErrHostLost) {
		t.Fatalf("hang-up: %v", err)
	}
}

func TestALowerEpochGoodbyeIsIgnored(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			h.reply("STATE", map[string]any{"host_id": "host-a", "online": false, "heartbeat": 1, "heartbeat_interval_ms": 1000, "authority_epoch": 0, "protocol_min": 1, "protocol_max": 1})
		}
	}}
	if _, err := session(t, host, 3); err != nil || countOf(host.types(), "NSYNC") != 3 {
		t.Fatalf("err %v after %v", err, host.types())
	}
}

func TestInvalidOptionsAreRefused(t *testing.T) {
	t.Parallel()
	for _, options := range []Options{{}, {SliceNS: 1, UntilNS: 1, ProtocolMax: 1, Environment: "simulated", SessionID: "a b"}, {SliceNS: 1, UntilNS: 1, ProtocolMax: 1, Environment: "lab", SessionID: "s"}} {
		if err := NewSession(utilityRun(t), &scriptedHost{t: t}, options).Run(t.Context()); !errors.Is(err, ErrBadOptions) {
			t.Fatalf("options %+v: %v", options, err)
		}
	}
	if wire.WireID("site-a/pond-1") != "site-a:pond-1" || wire.WireID("/x") != "d:x" {
		t.Fatal("wire ids")
	}
}
