package host

import (
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func (h *Host) Receive(in domain.Inbound) []map[string]any {
	if in.Unsequenced || h.Done() {
		return nil
	}
	arrival := h.in.Accept(in.Seq)
	if arrival == domain.Duplicate {
		return nil
	}
	asked := h.askRebirth(arrival)
	if !h.admits(&in) {
		return asked
	}
	return append(asked, h.dispatch(in.Message)...)
}

func (h *Host) admits(in *domain.Inbound) bool {
	if h.accepted {
		in.Within(h.session, h.version)
	}
	return in.Invalid == nil && (h.accepted || in.Type == "NBIRTH")
}

func (h *Host) askRebirth(arrival domain.Arrival) []map[string]any {
	if arrival != domain.Gap || h.asked {
		return nil
	}
	h.asked = true
	h.report.Rebirths++
	return h.reply(control("NCMD", map[string]any{"command": "rebirth"}))
}

func (h *Host) dispatch(m domain.Message) []map[string]any {
	switch m.Type {
	case "NBIRTH":
		return h.nodeBirth(m)
	case "NDEATH":
		h.nodeDeath(m)
	case "NSYNC":
		return h.sync(m)
	default:
		h.deviceRecord(m)
	}
	return nil
}

func (h *Host) deviceRecord(m domain.Message) {
	switch m.Type {
	case "DBIRTH":
		h.deviceBirth(m)
	case "DDATA":
		h.data(m)
	case "DDEATH":
		delete(h.devices, m.Device)
	case "DRECEIPT", "DRESULT":
		h.answer(m)
	}
}

func (h *Host) nodeBirth(m domain.Message) []map[string]any {
	h.asked, h.bdSeq = false, m.BDSeq
	if h.accepted {
		return nil
	}
	h.out = domain.NewTransmitter(m.SessionID)
	if refusal := h.refusal(m); refusal != "" {
		h.report.Refused, h.report.Reason = refusal, h.reason(refusal, m)
		return h.reply(control("STATE", h.state(false, refusal)))
	}
	h.accepted, h.session = true, m.SessionID
	h.version = int(min(m.ProtocolMax, int64(h.options.ProtocolMax)))
	h.out.UseVersion(h.version)
	return h.reply(control("STATE", h.state(true, "")))
}

func (h *Host) nodeDeath(m domain.Message) {
	if m.BDSeq >= h.bdSeq {
		clear(h.devices)
	}
}

func (h *Host) refusal(m domain.Message) string {
	switch {
	case domain.ProtocolInvalid(m):
		return "protocol_invalid"
	case domain.EnvironmentFaultFor(h.options.Environment, m) != nil:
		return "environment_mismatch"
	case domain.ContractUnsupported(m):
		return "contract_unsupported"
	case m.ProtocolMin > int64(h.options.ProtocolMax) || m.ProtocolMax < domain.Version:
		return "version_unsupported"
	case h.missingFromBirth(m) || domain.KindsMisdigested(m.Kinds) || domain.KindsDuplicated(m.Kinds):
		return "inventory_mismatch"
	}
	return ""
}

func (h *Host) reason(refusal string, m domain.Message) string {
	if fault := domain.EnvironmentFaultFor(h.options.Environment, m); refusal == "environment_mismatch" && fault != nil {
		return fault.Reason
	}
	return ""
}

func (h *Host) missingFromBirth(m domain.Message) bool {
	present := map[string]bool{}
	for _, board := range m.Boards {
		for _, member := range board.Members {
			present[member] = true
		}
	}
	for _, device := range slices.Sorted(maps.Keys(h.options.Inventory)) {
		if !present[device] {
			return true
		}
	}
	return false
}

func (h *Host) state(online bool, refusal string) map[string]any {
	state := map[string]any{"host_id": "reference-host", "online": online, "heartbeat": 0, "heartbeat_interval_ms": 1000,
		"authority_epoch": h.options.Epoch, "protocol_min": domain.Version, "protocol_max": h.options.ProtocolMax}
	if refusal != "" {
		state["refusal"] = refusal
	}
	return state
}

func (h *Host) reply(records ...domain.Record) []map[string]any {
	for _, record := range records {
		h.out.Queue(record)
	}
	return h.out.Drain()
}

func control(kind string, fields map[string]any) domain.Record {
	return domain.Record{Type: kind, Lane: domain.LaneControl, Fields: fields}
}
