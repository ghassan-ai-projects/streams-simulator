package app

import (
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestAPacedWorldStartsAtTheHostsAttachWithItsDisturbancesMeasuredFromThere(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t}
	pacer := &fakePacer{host: host, worldNow: sliceEnd(10) + 30e9}
	options := sliceOptions(13)
	options.Pace = pacer
	options.Disturbances = []domain.ScheduledFault{{Entity: "zone-01", Fault: "zone_heat_load", OnsetNS: model.DefaultStartTimeNS + 60e9}}
	r := runOf(t, testsupport.ThermalBoardDomain())
	if err := NewSession(r, host, options).Run(t.Context()); err != nil {
		t.Fatal(err)
	}
	if len(pacer.released) != 3 || pacer.released[0] != sliceEnd(11) {
		t.Fatalf("released %v, want the slices after the attach only, from %d: no catch-up burst", pacer.released, sliceEnd(11))
	}
	for _, record := range host.sent {
		if record["type"] == "DBIRTH" && record["birth_time_ns"] != float64(sliceEnd(10)) && record["birth_time_ns"] != sliceEnd(10) {
			t.Fatalf("DBIRTH at %v, want the attach slice %d", record["birth_time_ns"], sliceEnd(10))
		}
	}
	if faults := r.Faults(); len(faults) != 1 || faults[0].OnsetNS != sliceEnd(11) {
		t.Fatalf("faults %+v, want the heat load one minute after the attach", faults)
	}
}

func TestACommandWhoseDeadlinePassedOnThePacedClockIsExpiredWhileTheWorldLags(t *testing.T) {
	t.Parallel()
	script := func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			h.reply("DCMD", commandFields("c-late", "meter-1", "shed_load", map[string]any{"meter_id": "meter-1"}, sliceEnd(1)))
		}
	}
	_, unpaced := commandSession(t, 2, script, nil)
	_, lagging := commandSession(t, 2, script, func(o *Options) { o.Pace = &fakePacer{host: &scriptedHost{t: t}, lag: 300e9} })
	if first := unpaced["c-late"]; len(first) == 0 || first[0] != "DRECEIPT:accepted" {
		t.Fatalf("unpaced answers %v, want the command accepted on the world clock", first)
	}
	if late := lagging["c-late"]; len(late) != 1 || late[0] != "DRECEIPT:refusedexpired" {
		t.Fatalf("answers while the world lags the wall by 300 s: %v, want refused expired", late)
	}
}

func TestANodeKeepsReadingItsHostDuringALongPacedWaitAndActsOnWhatArrivedAtTheNextBarrier(t *testing.T) {
	t.Parallel()
	drains := 0
	host := &scriptedHost{t: t, duringDrain: func(h *scriptedHost) {
		drains++
		h.reply("STATE", map[string]any{"host_id": "host-a", "online": true, "heartbeat": drains, "heartbeat_interval_ms": 1000,
			"authority_epoch": 1, "protocol_min": 1, "protocol_max": 1})
		if drains == 3 {
			h.reply("STATE", map[string]any{"host_id": "host-a", "online": false, "heartbeat": drains, "heartbeat_interval_ms": 1000,
				"authority_epoch": 1, "protocol_min": 1, "protocol_max": 1})
		}
	}}
	options := sliceOptions(10)
	options.SliceNS = 200e9
	options.Pace = &fakePacer{host: host, wait: 200e9}
	if err := NewSession(utilityRun(t), host, options).Run(t.Context()); err != nil {
		t.Fatal(err)
	}
	if len(host.drained) != 3 || host.drained[0] != 200*time.Second || countOf(host.types(), "NSYNC") != 3 {
		t.Fatalf("drained %v with %d barriers, want the link read through every 200 s wait and the goodbye acted on at the third barrier",
			host.drained, countOf(host.types(), "NSYNC"))
	}
}
