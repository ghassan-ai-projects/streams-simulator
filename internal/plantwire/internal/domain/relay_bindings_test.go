package domain

import (
	"reflect"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func pumpDigest(t *testing.T) string {
	t.Helper()
	digest, err := KindDigest(birthKinds()[0].(map[string]any))
	if err != nil {
		t.Fatal(err)
	}
	return digest
}

func boundPump(t *testing.T, edit func(*Bound)) []Bound {
	t.Helper()
	bound := Bound{Index: 0, Instance: "pump-a", Board: "b-1", Device: "P1", Kind: "pump", KindDigest: pumpDigest(t)}
	if edit != nil {
		edit(&bound)
	}
	return []Bound{bound}
}

func notesOf(notes []Note, kind string) []map[string]any {
	var found []map[string]any
	for _, note := range notes {
		if note.Type == kind {
			found = append(found, note.Fields)
		}
	}
	return found
}

func plannedRig(t *testing.T, bound []Bound) *mixedRig {
	t.Helper()
	g := newMixedRig(t)
	g.relay.Plan(map[string]string{"S1": "sim-one"}, bound)
	g.relay.GatewayAttached()
	return g
}

func TestAGatewayThatAnnouncesTheBoundDeviceWithItsKindFileIsMatchedAndItsBirthsCarryTheirOrigin(t *testing.T) {
	t.Parallel()
	g := plannedRig(t, boundPump(t, nil))
	admitted := g.gw(hardwareBirth(t, nil))
	bindings := notesOf(admitted.Notes, NoteBinding)
	if len(bindings) != 1 || bindings[0]["outcome"] != "matched" || len(bindings[0]["issues"].([]model.Issue)) != 0 || !g.relay.GatewayAdmitted() {
		t.Fatalf("notes = %+v, want one matched binding and the gateway admitted", admitted.Notes)
	}
	g.relay.HostAttached()
	g.hostSends(hostState(7, 0))
	pump, valve := g.gw(deviceBirthOf("P1", "boot-1")), g.gw(deviceBirthOf("V1", "boot-7"))
	wantPump := map[string]any{"instance": "pump-a", "device": "P1", "board": "b-1", "environment": EnvironmentHardware, "boot_id": "boot-1"}
	if got := notesOf(pump.Notes, NoteOrigin); len(got) != 1 || !reflect.DeepEqual(got[0], wantPump) {
		t.Fatalf("origin of P1 = %+v, want %+v", got, wantPump)
	}
	if got := notesOf(valve.Notes, NoteOrigin); len(got) != 1 || got[0]["instance"] != nil || got[0]["environment"] != EnvironmentHardware {
		t.Fatalf("origin of V1 = %+v, want a hardware device the plan does not bind, with no instance", got)
	}
	g.simSends(Record{Type: "NBIRTH", Lane: LaneControl, Fields: simulatedBirthRecord(t, "s-1", []string{"S1", "S2"}, birthKinds()).Fields})
	world := g.simSends(simDeviceBirth("S1"))
	if got := notesOf(world.Notes, NoteOrigin); len(got) != 1 || got[0]["instance"] != "sim-one" || got[0]["environment"] != EnvironmentSimulated {
		t.Fatalf("origin of S1 = %+v, want the world's device as the instance sim-one, simulated", got)
	}
}

func TestAGatewayThatBreaksABindingIsRefusedWithTheIssueAtThePlansMember(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		edit             func(*Bound)
		pointer, keyword string
	}{
		"no such board":     {func(b *Bound) { b.Board = "b-9" }, "/real/0/board", "board_missing"},
		"no such device":    {func(b *Bound) { b.Device = "X9" }, "/real/0/device", "device_missing"},
		"no such kind":      {func(b *Bound) { b.Kind = "fan" }, "/real/0/instance", "kind_missing"},
		"another interface": {func(b *Bound) { b.KindDigest = "sha256:" + strings.Repeat("0", 64) }, "/real/0/instance", "kind_mismatch"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := plannedRig(t, boundPump(t, tc.edit))
			moves := g.gw(hardwareBirth(t, nil))
			bindings := notesOf(moves.Notes, NoteBinding)
			if len(moves.Down) != 1 || moves.Down[0]["refusal"] != "inventory_mismatch" || g.relay.GatewayAdmitted() || len(bindings) != 1 {
				t.Fatalf("down = %v notes = %+v, want an inventory_mismatch refusal and one binding note", moves.Down, moves.Notes)
			}
			issues := bindings[0]["issues"].([]model.Issue)
			if bindings[0]["outcome"] != "refused" || len(issues) != 1 || issues[0].Pointer != tc.pointer || issues[0].Keyword != tc.keyword {
				t.Fatalf("binding = %+v, want refused with one %s at %s", bindings[0], tc.keyword, tc.pointer)
			}
			if stats := g.relay.Stats(); stats.Refused != "inventory_mismatch" || stats.Reason != issues[0].Message || bindings[0]["reason"] != stats.Reason {
				t.Fatalf("stats = %+v, want the issue as the refusal's reason", stats)
			}
		})
	}
}

func TestABoundDeviceBornAsAnotherKindIsCarriedUntouchedAndReportedNotAsTheInstance(t *testing.T) {
	t.Parallel()
	g := plannedRig(t, boundPump(t, nil))
	g.gw(hardwareBirth(t, nil))
	g.relay.HostAttached()
	g.hostSends(hostState(7, 0))
	before := g.relay.Stats().Dropped
	moves := g.gwEdited(deviceBirthOf("P1", "boot-1"), func(record map[string]any) { record["kind"] = "valve" })
	bindings := notesOf(moves.Notes, NoteBinding)
	if len(moves.Up) != 1 || g.relay.Stats().Dropped != before || len(bindings) != 1 || bindings[0]["outcome"] != "refused" {
		t.Fatalf("up = %v notes = %+v, want the birth carried as the gateway sent it and a refused binding", typesOf(moves.Up), moves.Notes)
	}
	sameContent(t, moves.Up[0], g.sent)
	if origins := notesOf(moves.Notes, NoteOrigin); len(origins) != 1 || origins[0]["instance"] != nil {
		t.Fatalf("origins = %+v, want the device not taken for the instance it was bound to", origins)
	}
	if issues := bindings[0]["issues"].([]model.Issue); len(issues) != 1 || issues[0].Keyword != "device_kind_mismatch" || issues[0].Pointer != "/real/0/device" {
		t.Fatalf("issues = %+v, want device_kind_mismatch at /real/0/device", issues)
	}
}

func TestARebirthThatNoLongerHoldsTheBindingIsDroppedAndReported(t *testing.T) {
	t.Parallel()
	g := plannedRig(t, boundPump(t, nil))
	g.gw(hardwareBirth(t, nil))
	moves := g.gw(Record{Type: "NBIRTH", Lane: LaneControl, Fields: hardwareBirthMembers(t, "b-1", "V1")})
	bindings := notesOf(moves.Notes, NoteBinding)
	if len(bindings) != 1 || bindings[0]["outcome"] != "refused" || bindings[0]["refusal"] != "inventory_mismatch" {
		t.Fatalf("notes = %+v, want the rebirth refused against the plan", moves.Notes)
	}
}

func TestARelayWithoutAPlanNotesNoBindingAndAHardwareRelayNoOrigin(t *testing.T) {
	t.Parallel()
	mixed := newMixedRig(t)
	mixed.relay.GatewayAttached()
	if moves := mixed.gw(hardwareBirth(t, nil)); len(notesOf(moves.Notes, NoteBinding)) != 0 {
		t.Fatalf("notes = %+v, want no binding note without a plan", moves.Notes)
	}
	hardware := newRelayRig(t)
	hardware.relay.GatewayAttached()
	hardware.gw(hardwareBirth(t, nil))
	hardware.relay.HostAttached()
	hardware.hostSends(hostState(7, 0))
	if moves := hardware.gw(deviceBirthOf("P1", "boot-1")); len(moves.Up) != 1 || len(moves.Notes) != 0 {
		t.Fatalf("up = %v notes = %+v, want the birth carried and no origin note outside a mixed session", typesOf(moves.Up), moves.Notes)
	}
}
