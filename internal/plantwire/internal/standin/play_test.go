package standin

import (
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
)

func linkScript() Script {
	return Script{Steps: []Step{
		{AtNS: 30, Device: "zone-01", Link: LinkUp, BootID: "boot-2"},
		{AtNS: 10, Device: "zone-01", Samples: []Sample{{Metric: "temperature", Value: 22.0}, {Metric: "humidity", Value: 40.0, Quality: "suspect"}}},
		{AtNS: 5, Device: "fan-01", Samples: []Sample{{Metric: "level_percent", Value: 0}}},
		{AtNS: 20, Device: "zone-01", Link: LinkDown},
	}}
}

func TestScriptedStepsPlayInScriptTimeAndNeverBeforeTheirTime(t *testing.T) {
	t.Parallel()
	b := newBench(t, boardNode(t, linkScript()), hostOptions("hardware"))
	b.handshake()
	first := b.node.Play(9)
	if got := types(first); !reflect.DeepEqual(got, []string{"DDATA"}) || first[0]["device"] != "fan-01" {
		t.Fatalf("Play(9) = %v, want only the step at 5 ns, the fan's reading", first)
	}
	rest := b.node.Play(100)
	if got := types(rest); !reflect.DeepEqual(got, []string{"DDATA", "DDEATH", "DBIRTH"}) {
		t.Fatalf("Play(100) = %v, want the readings, the link loss and the rebirth in script order", got)
	}
	if again := b.node.Play(100); len(again) != 0 {
		t.Fatalf("a second Play(100) = %v, want nothing: every step plays once", again)
	}
	for _, record := range append(first, rest...) {
		if err := wireErr(record); err != nil {
			t.Errorf("%v fails the contract: %v", record["type"], err)
		}
	}
	b.deliver(append(first, rest...)...)
	if report := b.host.Report(); len(report.Quarantined) != 0 || report.Rebirths != 0 {
		t.Fatalf("host report %+v, want the whole script accepted in order", report)
	}
}

func TestReadingsCarryTheirAliasQualityAndScriptTime(t *testing.T) {
	t.Parallel()
	b := newBench(t, boardNode(t, linkScript()), hostOptions("hardware"))
	b.handshake()
	b.node.Play(5)
	data := b.node.Play(10)[0]
	want := []domain.Metric{{Alias: 1, Value: 22.0, Quality: "good"}, {Alias: 2, Value: 40.0, Quality: "suspect"}}
	metrics := data["metrics"].([]any)
	for i, expected := range want {
		sample := metrics[i].(map[string]any)
		if sample["alias"] != int(expected.Alias) || sample["value"] != expected.Value || sample["quality"] != expected.Quality ||
			sample["event_time_ns"] != int64(10) || sample["time_source"] != "device" {
			t.Errorf("sample %d = %v, want alias %d, value %v, quality %s at 10 ns from the device", i, sample, expected.Alias, expected.Value, expected.Quality)
		}
	}
}

func TestALinkLossIsADeviceDeathAndTheRebirthIsANewBoot(t *testing.T) {
	t.Parallel()
	b := newBench(t, boardNode(t, linkScript()), hostOptions("hardware"))
	b.handshake()
	played := b.node.Play(100)
	death, birth := played[2], played[3]
	if death["reason"] != "link_down" || death["boot_id"] != "boot-1" || death["device"] != "zone-01" || death["last_seq"] != uint64(4) {
		t.Fatalf("DDEATH = %v, want zone-01 link_down on boot-1 after record 4", death)
	}
	if birth["boot_id"] != "boot-2" || birth["birth_time_ns"] != int64(30) || birth["device"] != "zone-01" {
		t.Fatalf("DBIRTH = %v, want zone-01 reborn at 30 ns on boot-2", birth)
	}
}

func TestACommandIsAcknowledgedReportedAndSeenInTheNextReading(t *testing.T) {
	t.Parallel()
	step := host.Scripted{ID: "c1", AtNS: 0, Device: "fan-01", Command: "set_cooling", Args: map[string]any{"level_percent": 30},
		Verify: &host.Verify{Metric: "level_percent", Equals: 30, WithinNS: 10e9}}
	script := Script{
		Replies: []Reply{{Command: "set_cooling", Receipt: "accepted", Result: "succeeded"}},
		Steps:   []Step{{AtNS: 2e9, Device: "fan-01", Samples: []Sample{{Metric: "level_percent", Value: 30}}}},
	}
	b := newBench(t, boardNode(t, script), hostOptions("hardware", step))
	b.handshake()
	issued := b.deliver(b.node.Emit(domain.Sync(0, 1e9))...)
	if got := types(issued); !reflect.DeepEqual(got, []string{"DCMD", "HSYNC"}) {
		t.Fatalf("host answered the barrier with %v, want the scripted command and the barrier release", got)
	}
	answered := b.node.Handle(issued[0])
	if got := types(answered); !reflect.DeepEqual(got, []string{"DRECEIPT", "DRESULT"}) {
		t.Fatalf("the board answered with %v, want a receipt and a result", got)
	}
	b.deliver(answered...)
	b.deliver(b.node.Play(2e9)...)
	b.deliver(b.node.Emit(domain.Sync(0, 3e9))...)
	status := b.host.Report().Commands["c1"]
	if status == nil || !reflect.DeepEqual(status.Receipts, []string{"accepted"}) || status.Result != "succeeded" || status.Verified == nil || !*status.Verified {
		t.Fatalf("command c1 = %+v, want accepted, succeeded and verified by the reading that followed", status)
	}
}
