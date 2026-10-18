package standin

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func TestAHardwareNodeBirthsItsKindFilesWithTheirDigestsAndTheHostAcceptsThem(t *testing.T) {
	t.Parallel()
	b := newBench(t, boardNode(t, Script{}), hostOptions("hardware"))
	birth, err := b.node.Birth()
	if err != nil {
		t.Fatal(err)
	}
	if err := wireErr(birth); err != nil {
		t.Fatalf("NBIRTH fails the contract: %v", err)
	}
	if birth["environment"] != "hardware" || birth["seq"] != uint64(0) || birth["node_metrics"].(map[string]any)["clock_mode"] != "wall" {
		t.Fatalf("NBIRTH = %v, want the first record of a hardware node on a wall clock", birth)
	}
	for _, kind := range birth["kinds"].([]any) {
		announced := kind.(map[string]any)
		want, err := domain.KindDigest(kindFile(t, map[string]string{"thermal_sensor": sensorKind, "fan": fanKind}[announced["kind"].(string)]))
		if err != nil || announced["kind_digest"] != want {
			t.Errorf("kind %v announces %v, its kind file digests to %s (%v)", announced["kind"], announced["kind_digest"], want, err)
		}
	}
	if replies := b.deliver(birth); len(replies) != 1 || replies[0]["type"] != "STATE" || replies[0]["online"] != true || !b.host.Accepted() {
		t.Fatalf("host answered %v, want it to accept the session with an online STATE", replies)
	}
}

func TestTheBoardIsBornDeviceByDeviceWithAliasesAndWarmingUpReadings(t *testing.T) {
	t.Parallel()
	b := newBench(t, boardNode(t, Script{}), hostOptions("hardware"))
	births := b.handshake()
	if got := types(births); !reflect.DeepEqual(got, []string{"DBIRTH", "DBIRTH"}) {
		t.Fatalf("born as %v, want one DBIRTH per device", got)
	}
	if report := b.host.Report(); len(report.Quarantined) != 0 || report.Rebirths != 0 {
		t.Fatalf("host report %+v, want nothing quarantined and no rebirth asked", report)
	}
	want := []any{
		map[string]any{"name": "temperature", "alias": 1, "value": 21.5, "quality": "good"},
		map[string]any{"name": "humidity", "alias": 2, "value": nil, "quality": "warming_up"},
	}
	if births[0]["device"] != "zone-01" || births[0]["board"] != "mega-01" || births[0]["boot_id"] != "boot-1" || !reflect.DeepEqual(births[0]["metrics"], want) {
		t.Fatalf("DBIRTH of zone-01 = %v, want its readings %v", births[0], want)
	}
	for _, birth := range births {
		if err := wireErr(birth); err != nil {
			t.Errorf("%v fails the contract: %v", birth["device"], err)
		}
	}
}

func TestAHostThatRefusesTheSessionEndsTheHandshakeWithTheReason(t *testing.T) {
	t.Parallel()
	b := newBench(t, boardNode(t, Script{}), hostOptions("simulated"))
	birth, err := b.node.Birth()
	if err != nil {
		t.Fatal(err)
	}
	replies := b.deliver(birth)
	_, err = b.node.Accept(replies[0])
	if !errors.Is(err, domain.ErrRefused) || !strings.Contains(err.Error(), "environment_mismatch") {
		t.Fatalf("Accept = %v, want domain.ErrRefused naming environment_mismatch", err)
	}
}

func TestAcceptRefusesAnythingThatIsNoValidRecord(t *testing.T) {
	t.Parallel()
	node := boardNode(t, Script{})
	for name, record := range map[string]map[string]any{"a record of no type": {"seq": 1}, "a value JSON cannot hold": {"type": "STATE", "bad": make(chan int)}} {
		if _, err := node.Accept(record); !errors.Is(err, ErrRecord) {
			t.Errorf("%s: Accept = %v, want ErrRecord", name, err)
		}
	}
}

func TestAShutdownIsANodeDeathInSequence(t *testing.T) {
	t.Parallel()
	b := newBench(t, boardNode(t, Script{}), hostOptions("hardware"))
	b.handshake()
	death := b.node.Shutdown()
	if err := wireErr(death); err != nil || death["type"] != "NDEATH" || death["reason"] != "shutdown" || death["seq"] != uint64(3) {
		t.Fatalf("Shutdown = %v, %v, want NDEATH shutdown as the fourth record of the session", death, err)
	}
	if replies := b.deliver(death); len(replies) != 0 {
		t.Fatalf("host answered a node death with %v", replies)
	}
}

func TestTwoNodesOfOneScriptSpeakTheSameBytes(t *testing.T) {
	t.Parallel()
	script := Script{Steps: []Step{{AtNS: 5, Device: "zone-01", Samples: []Sample{{Metric: "temperature", Value: 22.0}}}, {AtNS: 9, Device: "zone-01", Link: LinkDown}}}
	speak := func() string {
		b := newBench(t, boardNode(t, script), hostOptions("hardware"))
		records := append(b.handshake(), b.node.Play(10)...)
		raw, err := json.Marshal(records)
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	if first, second := speak(), speak(); first != second || !strings.Contains(first, "DDEATH") {
		t.Fatalf("two runs differ:\n%s\n%s", first, second)
	}
}
