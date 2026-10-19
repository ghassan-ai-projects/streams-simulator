package domain

import (
	"cmp"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

type GateAnswer struct {
	CommandID, Device, Reason string
}

type StopOrder struct {
	Device, Command string
}

type GateOutcome struct {
	Entries []InterruptEntry
	Proceed bool
	Queued  []GateAnswer
	Refused []GateAnswer
	Forward []Inbound
	Stops   []StopOrder
	Reply   map[string]any
}

type InterruptGate struct {
	machine   *Interrupts
	payloads  map[string]Inbound
	inventory string
}

func NewInterruptGate() *InterruptGate {
	return &InterruptGate{machine: NewInterrupts(), payloads: map[string]Inbound{}}
}

func (g *InterruptGate) View() map[string]any { return g.machine.View() }

func (g *InterruptGate) SetInventory(devices []InterruptDevice, clock string, nodeNS int64) GateOutcome {
	sorted := slices.SortedFunc(slices.Values(devices), func(a, b InterruptDevice) int { return cmp.Compare(a.Device, b.Device) })
	listed := make([]any, len(sorted))
	for i, d := range sorted {
		listed[i] = d.Fields()
	}
	digest, _ := canonical.Digest(listed)
	if digest == g.inventory {
		return GateOutcome{}
	}
	g.inventory = digest
	return g.run(newEntry(EventInventory, "devices", listed).at(clock, nodeNS), nil)
}

func (g *InterruptGate) Judge(in Inbound, clock string, nodeNS int64) GateOutcome {
	c := CommandOf(in.Message)
	args, truncated := boundedArgs(c.Args)
	input := newEntry(EventCommand, "command_id", c.ID, "key", c.Key, "content", c.Content(), "device", c.Device, "name", c.Name,
		"deadline_ns", c.DeadlineNS, "args", args, "args_truncated", truncated).at(clock, nodeNS)
	return g.run(input, &in)
}

func boundedArgs(args map[string]any) (map[string]any, bool) {
	if len(args) == 0 {
		return map[string]any{}, false
	}
	raw, err := canonical.Marshal(args)
	if err != nil || len(raw) > maxLoggedArgsBytes {
		return map[string]any{}, true
	}
	return args, false
}

func (g *InterruptGate) Covers(request InterruptEntry, c Command) bool {
	restricts := request.Op() == OpHold || request.Op() == OpEStop || (request.Op() == OpApproveMode && request["enabled"] == true)
	device := g.machine.deviceNamed(c.Device)
	return restricts && scopeOf(request["scope"]).covers(device) && !declaredStop(device, c.Name, c.Args, false)
}

func (g *InterruptGate) Control(request InterruptEntry, clock string, nodeNS int64) GateOutcome {
	return g.run(request.at(clock, nodeNS), nil)
}

func (g *InterruptGate) Advance(clock string, nodeNS int64) GateOutcome {
	return g.run(newEntry(EventClock).at(clock, nodeNS), nil)
}

func (g *InterruptGate) Flush(clock string, nodeNS int64) GateOutcome {
	return g.run(newEntry(EventFlush).at(clock, nodeNS), nil)
}

func (g *InterruptGate) run(input InterruptEntry, judged *Inbound) GateOutcome {
	outputs := g.machine.Step(input)
	outcome := GateOutcome{Entries: loggedEntries(input, outputs)}
	for _, out := range outputs {
		g.absorb(&outcome, out, judged)
	}
	return outcome
}

func loggedEntries(input InterruptEntry, outputs []InterruptEntry) []InterruptEntry {
	if (input.Event() == EventClock || input.Event() == EventFlush) && len(outputs) == 0 {
		return nil
	}
	return append([]InterruptEntry{input}, outputs...)
}

func (g *InterruptGate) absorb(o *GateOutcome, out InterruptEntry, judged *Inbound) {
	switch out.Event() {
	case EventVerdict:
		g.verdict(o, out, judged)
	case EventResolve:
		g.resolved(o, out)
	case EventStop:
		o.Stops = append(o.Stops, StopOrder{Device: out.Text("device"), Command: out.Text("command")})
	case EventReply:
		o.Reply = out
	}
}

func (g *InterruptGate) verdict(o *GateOutcome, out InterruptEntry, judged *Inbound) {
	if judged == nil {
		return
	}
	id := out.Text("command_id")
	switch out.Text("verdict") {
	case VerdictPass:
		o.Proceed = true
	case VerdictQueue:
		g.queue(o, id, *judged)
	default:
		o.Refused = append(o.Refused, GateAnswer{CommandID: id, Device: judged.Device, Reason: out.Text("reason")})
	}
}

func (g *InterruptGate) queue(o *GateOutcome, id string, in Inbound) {
	if _, held := g.payloads[id]; !held {
		g.payloads[id] = in
	}
	o.Queued = append(o.Queued, GateAnswer{CommandID: id, Device: in.Device})
}

func (g *InterruptGate) resolved(o *GateOutcome, out InterruptEntry) {
	ids := append([]string{out.Text("command_id")}, stringsOf(out["waiting"])...)
	for _, id := range ids {
		payload, held := g.payloads[id]
		delete(g.payloads, id)
		if !held {
			continue
		}
		if out.Text("resolution") == ResolutionForward {
			o.Forward = append(o.Forward, payload)
			continue
		}
		o.Refused = append(o.Refused, GateAnswer{CommandID: id, Device: payload.Device, Reason: out.Text("reason")})
	}
}
