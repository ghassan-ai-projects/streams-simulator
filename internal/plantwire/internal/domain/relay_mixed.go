package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
)

var (
	ErrSimulatedBirth = errors.New("plantwire: the simulated node's birth cannot join a mixed relay")
	ErrNotRecord      = errors.New("plantwire: not a plant wire record")
)

type simSide struct {
	message Message
	devices map[string]bool
	birth   map[string]any
	out     *Transmitter
	in      Receiver
	started bool
	lost    bool
	born    map[string]string
}

func NewMixedRelay(tickMS int64, simulated Inbound) (*Relay, error) {
	admission := Admission{Environment: EnvironmentSimulated, ProtocolMax: relayProtocol}
	if simulated.Invalid != nil || simulated.Type != "NBIRTH" || admission.Refusal(simulated.Message, false) != "" {
		return nil, simulatedBirthRefusal(simulated.Invalid)
	}
	relay := NewRelay(tickMS)
	relay.sim = &simSide{message: simulated.Message, devices: devicesOf(simulated.Boards), birth: content(simulated.Record),
		out: NewTransmitter(simulated.SessionID), born: map[string]string{}}
	return relay, nil
}

func simulatedBirthRefusal(invalid error) error {
	if invalid == nil {
		return fmt.Errorf("%w: it must be a valid NBIRTH of a simulated node", ErrSimulatedBirth)
	}
	return fmt.Errorf("%w: it must be a valid NBIRTH of a simulated node (%w)", ErrSimulatedBirth, invalid)
}

func BirthInbound(record Record, session string) (Inbound, error) {
	out := NewTransmitter(session)
	out.Queue(record)
	return InboundOf(out.Drain()[0])
}

func InboundOf(record map[string]any) (Inbound, error) {
	raw, err := json.Marshal(record)
	if err != nil {
		return Inbound{}, fmt.Errorf("%w: %w", ErrNotRecord, err)
	}
	in, ok := Decode(raw)
	if !ok {
		return Inbound{}, fmt.Errorf("%w: no JSON object", ErrNotRecord)
	}
	return in, nil
}

func (r *Relay) gatewayRefusal(in Inbound) (refusal, reason string) {
	if refusal = r.admission.Refusal(in.Message, false); refusal != "" {
		return refusal, r.admission.Reason(refusal, in.Message)
	}
	if r.sim == nil {
		return "", ""
	}
	if reason = Collision(in.Message, r.sim.message); reason != "" {
		return "inventory_mismatch", reason
	}
	if reason = r.bindingRefusal(in.Message); reason != "" {
		return "inventory_mismatch", reason
	}
	return "", ""
}

func Collision(gateway, simulated Message) string {
	if board := firstShared(boardNames(gateway.Boards), boardNames(simulated.Boards)); board != "" {
		return fmt.Sprintf("board %q is announced by the gateway and by the simulated world", clip(board))
	}
	if device := firstShared(devicesOf(gateway.Boards), devicesOf(simulated.Boards)); device != "" {
		return fmt.Sprintf("device %q is on a gateway board and on a simulated board", clip(device))
	}
	if kind := differingKind(gateway.Kinds, simulated.Kinds); kind != "" {
		return fmt.Sprintf("kind %q has a different kind_digest on the gateway and in the simulated world", clip(kind))
	}
	return ""
}

func devicesOf(boards []Board) map[string]bool {
	devices := map[string]bool{}
	for _, board := range boards {
		for _, member := range board.Members {
			devices[member] = true
		}
	}
	return devices
}

func boardNames(boards []Board) map[string]bool {
	names := map[string]bool{}
	for _, board := range boards {
		names[board.Board] = true
	}
	return names
}

func firstShared(a, b map[string]bool) string {
	for _, name := range slices.Sorted(maps.Keys(a)) {
		if b[name] {
			return name
		}
	}
	return ""
}

func differingKind(gateway, simulated []any) string {
	digests := map[string]string{}
	for _, entry := range gateway {
		digests[kindNamed(entry)] = digestOfKind(entry)
	}
	for _, entry := range simulated {
		if digest, shared := digests[kindNamed(entry)]; shared && digest != digestOfKind(entry) {
			return kindNamed(entry)
		}
	}
	return ""
}

func kindNamed(entry any) string {
	fields, _ := entry.(map[string]any)
	name, _ := fields["kind"].(string)
	return name
}

func digestOfKind(entry any) string {
	fields, _ := entry.(map[string]any)
	digest, _ := fields["kind_digest"].(string)
	return digest
}

func listOf(value any) []any {
	list, _ := value.([]any)
	return list
}

func (r *Relay) birthContent() map[string]any {
	if r.sim == nil {
		return content(r.birth.Record)
	}
	return mergedBirth(r.birth.Record, r.sim.birth)
}

func mergedBirth(gateway, simulated map[string]any) map[string]any {
	merged := content(gateway)
	kinds := mergedKinds(listOf(gateway["kinds"]), listOf(simulated["kinds"]))
	boards := append(slices.Clone(listOf(gateway["boards"])), listOf(simulated["boards"])...)
	merged["environment"] = AnnouncedEnvironment(boardEnvironments(boards))
	merged["kinds"] = kinds
	merged["boards"] = boards
	merged["capability_digest"] = CapabilityDigest(kinds)
	if pacedByWall(simulated) {
		merged["node_metrics"] = wallClockMetrics(gateway["node_metrics"])
	}
	return merged
}

func pacedByWall(birth map[string]any) bool {
	metrics, _ := birth["node_metrics"].(map[string]any)
	return metrics["clock_mode"] == ClockModeWall
}

func wallClockMetrics(announced any) map[string]any {
	metrics, _ := announced.(map[string]any)
	metrics = maps.Clone(metrics)
	if metrics == nil {
		metrics = map[string]any{}
	}
	metrics["clock_mode"] = ClockModeWall
	return metrics
}

func boardEnvironments(boards []any) map[string]bool {
	environments := map[string]bool{}
	for _, board := range boards {
		fields, _ := board.(map[string]any)
		environment, _ := fields["environment"].(string)
		environments[environment] = true
	}
	return environments
}

func mergedKinds(gateway, simulated []any) []any {
	named := map[string]bool{}
	merged := slices.Clone(gateway)
	for _, entry := range gateway {
		named[kindNamed(entry)] = true
	}
	for _, entry := range simulated {
		if !named[kindNamed(entry)] {
			merged = append(merged, entry)
		}
	}
	return merged
}

func (r *Relay) strayDevice(in Inbound) bool {
	return r.sim != nil && in.Device != "" && !r.gatewayDevices[in.Device]
}

func (r *Relay) markBirthed(seq uint64) {
	r.host.birthed, r.host.birthSeq = true, seq
	clear(r.gateway.born)
	if r.sim != nil {
		clear(r.sim.born)
	}
}
