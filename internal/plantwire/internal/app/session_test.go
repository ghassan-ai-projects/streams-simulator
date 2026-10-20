package app

import (
	"errors"
	"io/fs"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func TestASessionBirthsEveryDeviceBeforeItsDataAndSyncsEachSlice(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t}
	if _, err := session(t, host, 5); err != nil {
		t.Fatal(err)
	}
	types := host.types()
	if types[0] != "NBIRTH" || types[len(types)-1] != "NDEATH" || countOf(types, "NSYNC") != 5 || countOf(types, "DBIRTH") != 4 {
		t.Fatalf("record types %v", types)
	}
	born := map[string]bool{}
	for i, record := range host.sent {
		if wire.Seq(record["seq"]) != uint64(i) {
			t.Fatalf("record %d has seq %v", i, record["seq"])
		}
		device, _ := record["device"].(string)
		switch record["type"] {
		case "DBIRTH":
			born[device] = true
		case "DDATA":
			if !born[device] {
				t.Fatalf("data of %s before its birth", device)
			}
		}
	}
}

func TestARefusedSessionEndsAfterItsBirth(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t, refusal: "environment_mismatch"}
	if _, err := session(t, host, 2); !errors.Is(err, wire.ErrRefused) || len(host.sent) != 1 {
		t.Fatalf("err %v after %v", err, host.types())
	}
}

func TestARebirthRepeatsEveryBirthInTheSameSession(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			h.reply("NCMD", map[string]any{"command": "rebirth"})
		}
	}}
	if _, err := session(t, host, 2); err != nil {
		t.Fatal(err)
	}
	if births := birthsOf(host); len(births) != 2 || wire.Seq(births[1]["bd_seq"]) != 1 || births[1]["session_id"] != births[0]["session_id"] {
		t.Fatalf("births %v", births)
	}
}

func TestAHostSequenceGapEndsTheSession(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) { h.seq += 2 }}
	if _, err := session(t, host, 3); !errors.Is(err, ErrHostGap) {
		t.Fatalf("err = %v, want a host gap", err)
	}
}

func TestThreeMissedBarriersAreHostLoss(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t, silent: true}
	if _, err := session(t, host, 5); !errors.Is(err, ErrHostLost) || countOf(host.types(), "NSYNC") != 3 {
		t.Fatalf("err %v after %v", err, host.types())
	}
}

func TestAStateQueryIsAnsweredWithSelfReportedValues(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			h.reply("DQUERY", map[string]any{"device": "meter-1"})
		}
	}}
	if _, err := session(t, host, 2); err != nil {
		t.Fatal(err)
	}
	for _, record := range host.sent {
		if record["type"] == "DSTATE" && record["device"] == "meter-1" && record["self_reported"] == true && len(record["values"].([]any)) > 0 {
			return
		}
	}
	t.Fatalf("no state answer in %v", host.types())
}

func TestNoWireRecordCarriesAFaultOrItsOnset(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t}
	r := utilityRun(t)
	onset := model.DefaultStartTimeNS + 90e9
	if _, err := r.InjectFault("meter-1", "load_surge", onset, nil); err != nil {
		t.Fatal(err)
	}
	options := Options{Environment: "simulated", ProtocolMax: 1, SessionID: "s-1", SliceNS: 60e9, UntilNS: model.DefaultStartTimeNS + 300e9}
	if err := NewSession(r, host, options).Run(t.Context()); err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"load_surge", "meter_jitter", "flow_loss", model.FormatTime(onset)} {
		if strings.Contains(string(host.transcript), secret) {
			t.Fatalf("the wire carries %q", secret)
		}
	}
}

func countOf(types []string, kind string) int {
	return len(slices.DeleteFunc(slices.Clone(types), func(t string) bool { return t != kind }))
}

func birthsOf(host *scriptedHost) []map[string]any {
	var births []map[string]any
	for _, record := range host.sent {
		if record["type"] == "NBIRTH" {
			births = append(births, record)
		}
	}
	return births
}

func TestAHostGoodbyeEndsTheSessionEarly(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			h.reply("STATE", map[string]any{"host_id": "host-a", "online": false, "heartbeat": 1, "heartbeat_interval_ms": 1000,
				"authority_epoch": 1, "protocol_min": 1, "protocol_max": 1})
		}
	}}
	if _, err := session(t, host, 5); err != nil || countOf(host.types(), "NSYNC") != 1 || host.types()[len(host.types())-1] != "NDEATH" {
		t.Fatalf("err %v after %v", err, host.types())
	}
}

func TestServingAMissingSocketFails(t *testing.T) {
	t.Parallel()
	if err := Serve(t.Context(), utilityRun(t), "/nonexistent/host.sock", time.Second, Options{}); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("serve = %v, want the missing socket", err)
	}
}

func TestAHostThatNeverAnswersTheBirthIsLost(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t}
	host.sent = []map[string]any{{"type": "placeholder"}}
	if _, err := session(t, host, 1); !errors.Is(err, ErrHostLost) {
		t.Fatalf("err = %v, want host loss", err)
	}
}

func TestDuplicateHostRecordsAreDroppedAndTheEpochOnlyRises(t *testing.T) {
	t.Parallel()
	var s *Session
	host := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			h.reply("STATE", map[string]any{"host_id": "host-a", "online": true, "heartbeat": 1, "heartbeat_interval_ms": 1000, "authority_epoch": 3, "protocol_min": 1, "protocol_max": 1})
			h.pending = append(h.pending, h.pending[len(h.pending)-1])
			h.reply("STATE", map[string]any{"host_id": "host-a", "online": true, "heartbeat": 2, "heartbeat_interval_ms": 1000, "authority_epoch": 2, "protocol_min": 1, "protocol_max": 1})
		}
	}}
	r := utilityRun(t)
	s = NewSession(r, host, Options{Environment: "simulated", ProtocolMax: 1, SessionID: "s-1", SliceNS: 60e9, UntilNS: model.DefaultStartTimeNS + 120e9})
	if err := s.Run(t.Context()); err != nil || s.epoch != 3 {
		t.Fatalf("err %v, epoch floor %d, want 3", err, s.epoch)
	}
}

func TestDevicesOnABoardShareItsBoot(t *testing.T) {
	t.Parallel()
	d := newDevices(structuredRun(t), nil, nil)
	if d.board("T-main") == "T-main" || d.boot("T-main") != d.boot(d.spec.Spec.Boards[0].Members[1]) {
		t.Fatalf("board of T-main %q, boots %q and %q", d.board("T-main"), d.boot("T-main"), d.boot(d.spec.Spec.Boards[0].Members[1]))
	}
	if d.board("pump-9") != "pump-9" {
		t.Fatal("a device on no board is its own board")
	}
}
