package domain

import (
	"fmt"
	"slices"
)

const (
	stopDeadlineNS    = int64(30e9)
	maxIssuedStops    = 4096
	stopPolicy        = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	stopIDPrefix      = "node-stop-"
	stopNotBorn       = "device_not_born"
	stopSimulatorGone = "simulated_node_lost"
)

type relayInterrupts struct {
	gate   *InterruptGate
	wallNS int64
	simNS  int64
	kinds  map[string]string
	orders int
	issued map[string]string
}

func (r *Relay) EnableInterrupts() {
	r.ints = &relayInterrupts{gate: NewInterruptGate(), kinds: map[string]string{}, issued: map[string]string{}}
}

func (r *Relay) InterruptView() map[string]any {
	if r.ints == nil {
		return EmptyInterruptView()
	}
	return r.ints.gate.View()
}

func (r *Relay) SetWall(wallNS int64) {
	if r.ints != nil {
		r.ints.wallNS = wallNS
	}
}

func (r *Relay) noteKind(in Inbound) {
	if r.ints != nil {
		r.ints.kinds[in.Device] = in.Kind
	}
}

func (r *Relay) noteWorldTime(in Inbound) {
	if r.ints != nil {
		r.ints.simNS = max(r.ints.simNS, in.WorldTimeNS)
	}
}

func (i *relayInterrupts) clockOf(origin Origin) (string, int64) {
	if origin == OriginSimulated {
		return ClockWorld, i.simNS
	}
	return ClockWall, i.wallNS
}

func (i *relayInterrupts) now(clock string) int64 {
	if clock == ClockWorld {
		return i.simNS
	}
	return i.wallNS
}

func (r *Relay) Control(request InterruptEntry) (map[string]any, Moves) {
	moves := Moves{}
	r.syncInterruptInventory(&moves)
	outcome := r.ints.gate.Control(request, ClockWall, r.ints.wallNS)
	r.applyGate(&moves, outcome)
	return outcome.Reply, moves
}

func (r *Relay) gatedCommand(m *Moves, in Inbound) {
	r.syncInterruptInventory(m)
	clock, now := r.ints.clockOf(r.OriginOf(in.Device))
	outcome := r.ints.gate.Judge(in, r.declaresSafeStop(in), clock, now)
	r.applyGate(m, outcome)
	if outcome.Proceed {
		r.onward(m, in)
	}
}

func (r *Relay) onward(m *Moves, in Inbound) {
	if r.sim != nil {
		r.routeDown(m, in)
		return
	}
	r.forwardDown(m, in)
}

func (r *Relay) syncInterruptInventory(m *Moves) {
	outcome := r.ints.gate.SetInventory(r.interruptInventory(), ClockWall, r.ints.wallNS)
	m.Interrupt = append(m.Interrupt, outcome.Entries...)
}

func (r *Relay) applyGate(m *Moves, o GateOutcome) {
	m.Interrupt = append(m.Interrupt, o.Entries...)
	if r.host.shaken && r.host.out != nil {
		r.answerGate(m, o)
	}
	r.issueStops(m, o.Stops)
}

func (r *Relay) answerGate(m *Moves, o GateOutcome) {
	for _, queued := range o.Queued {
		r.transmitUp(m, Receipt(queued.CommandID, "queued", "", r.queuedBoot(queued.Device)))
	}
	for _, refused := range o.Refused {
		r.transmitUp(m, Receipt(refused.CommandID, "refused", refused.Reason, noBoot))
	}
	for _, forward := range o.Forward {
		r.onward(m, forward)
	}
}

func (r *Relay) queuedBoot(device string) string {
	if boot := r.bootOf(device); boot != "" {
		return boot
	}
	return noBoot
}

func (r *Relay) publishInventory(m *Moves) {
	if r.ints != nil {
		r.syncInterruptInventory(m)
	}
}

func (r *Relay) expireApprovals(m *Moves) {
	if r.ints == nil {
		return
	}
	for _, clock := range []string{ClockWall, ClockWorld} {
		r.applyGate(m, r.ints.gate.Advance(clock, r.ints.now(clock)))
	}
}

func (r *Relay) flushApprovals(m *Moves) {
	if r.ints == nil {
		return
	}
	for _, clock := range []string{ClockWall, ClockWorld} {
		m.Interrupt = append(m.Interrupt, r.ints.gate.Flush(clock, r.ints.now(clock)).Entries...)
	}
}

func (r *Relay) ownAnswer(m *Moves, in Inbound) bool {
	device, issued := "", false
	if r.ints != nil && answersCommand(in.Type) {
		device, issued = r.ints.issuedTo(in.CommandID)
	}
	if issued {
		m.Interrupt = append(m.Interrupt, EffectEntry(in.CommandID, device, answerText(in), ClockWall, r.ints.wallNS))
	}
	return issued
}

func (i *relayInterrupts) issuedTo(commandID string) (string, bool) {
	device, issued := i.issued[commandID]
	return device, issued
}

func answerText(in Inbound) string {
	if in.Type == "DRESULT" {
		return in.ReportedOutcome
	}
	if in.Reason != "" {
		return in.Status + ":" + in.Reason
	}
	return in.Status
}

func (r *Relay) declaresSafeStop(in Inbound) bool {
	kind := r.ints.kinds[in.Device]
	return kind != "" && slices.ContainsFunc(r.kindLists(), func(list []any) bool { return safeStopOf(list, kind) == in.Command })
}

func (r *Relay) kindLists() [][]any {
	if r.sim == nil {
		return [][]any{r.birth.Kinds}
	}
	return [][]any{r.birth.Kinds, r.sim.message.Kinds}
}

func safeStopOf(kinds []any, kind string) string {
	for _, entry := range kinds {
		fields, _ := entry.(map[string]any)
		if kindNamed(entry) != kind {
			continue
		}
		for _, command := range objects(fields["commands"]) {
			if command["safe_stop"] == true {
				return text(command, "name")
			}
		}
	}
	return ""
}

func (r *Relay) interruptInventory() []InterruptDevice {
	var devices []InterruptDevice
	if r.gateway.admitted {
		devices = r.boardDevices(r.birth.Message, ClockWall)
	}
	if r.sim != nil {
		devices = append(devices, r.boardDevices(r.sim.message, ClockWorld)...)
	}
	return devices
}

func (r *Relay) boardDevices(birth Message, clock string) []InterruptDevice {
	var devices []InterruptDevice
	for _, board := range birth.Boards {
		for _, member := range board.Members {
			devices = append(devices, InterruptDevice{Device: member, Board: board.Board, Clock: clock,
				SafeStop: safeStopOf(birth.Kinds, r.ints.kinds[member])})
		}
	}
	return devices
}

func (r *Relay) issueStops(m *Moves, orders []StopOrder) {
	for _, order := range orders {
		r.issueStop(m, order)
	}
}

func (r *Relay) issueStop(m *Moves, order StopOrder) {
	origin, boot := r.OriginOf(order.Device), r.bootOf(order.Device)
	switch {
	case boot == "":
		m.Interrupt = append(m.Interrupt, r.stopFailure(order, stopNotBorn))
	case origin == OriginSimulated && r.sim.lost:
		m.Interrupt = append(m.Interrupt, r.stopFailure(order, stopSimulatorGone))
	case origin == OriginGateway:
		r.transmitDown(m, r.stopRecord(order, origin, boot))
	default:
		r.sendSim(m, r.stopRecord(order, origin, boot))
	}
}

func (r *Relay) bootOf(device string) string {
	if boot, born := r.gateway.born[device]; born && r.gatewayDevices[device] {
		return boot
	}
	if r.sim != nil && r.sim.devices[device] {
		return r.sim.born[device]
	}
	return ""
}

func (r *Relay) stopFailure(order StopOrder, why string) InterruptEntry {
	return EffectEntry("none", order.Device, why, ClockWall, r.ints.wallNS)
}

func (r *Relay) stopRecord(order StopOrder, origin Origin, boot string) Record {
	_, now := r.ints.clockOf(origin)
	if len(r.ints.issued) >= maxIssuedStops {
		clear(r.ints.issued)
	}
	r.ints.orders++
	id := fmt.Sprintf("%s%d", stopIDPrefix, r.ints.orders)
	r.ints.issued[id] = order.Device
	command := Command{Device: order.Device, Name: order.Command, Args: map[string]any{}, Boot: boot}
	return Record{Type: "DCMD", Lane: LaneControl, Fields: map[string]any{"command_id": id, "idempotency_key": command.Content(),
		"device": order.Device, "command": order.Command, "args": map[string]any{}, "expected_boot_id": boot, "deadline_ns": now + stopDeadlineNS,
		"not_before_ns": int64(0), "authority_epoch": exactInt(r.host.state["authority_epoch"]), "policy_digest": stopPolicy,
		"fallback_on_lease_loss": "safe_off"}}
}
