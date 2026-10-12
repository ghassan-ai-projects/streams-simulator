package host

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func (h *Host) Receive(record map[string]any) []map[string]any {
	arrival := h.in.Accept(domain.Seq(record["seq"]))
	if arrival == domain.Duplicate || h.report.Refused != "" {
		return nil
	}
	asked := h.askRebirth(arrival)
	if domain.Check(record) != nil || (!h.accepted && record["type"] != "NBIRTH") {
		return asked
	}
	return append(asked, h.dispatch(record)...)
}

func (h *Host) askRebirth(arrival domain.Arrival) []map[string]any {
	if arrival != domain.Gap || h.asked {
		return nil
	}
	h.asked = true
	h.report.Rebirths++
	return h.reply(control("NCMD", map[string]any{"command": "rebirth"}))
}

func (h *Host) dispatch(record map[string]any) []map[string]any {
	switch record["type"] {
	case "NBIRTH":
		return h.nodeBirth(record)
	case "DBIRTH":
		h.deviceBirth(record)
	case "DDATA":
		h.data(record)
	case "DRECEIPT", "DRESULT":
		h.answer(record)
	case "NSYNC":
		return h.sync(record)
	}
	return nil
}

func (h *Host) nodeBirth(record map[string]any) []map[string]any {
	h.asked = false
	if h.accepted {
		return nil
	}
	h.out = domain.NewTransmitter(record["session_id"].(string))
	if refusal := h.refusal(record); refusal != "" {
		h.report.Refused = refusal
		return h.reply(control("STATE", h.state(false, refusal)))
	}
	h.accepted = true
	h.out.UseVersion(int(min(domain.Int(record["protocol_max"]), int64(h.options.ProtocolMax))))
	return h.reply(control("STATE", h.state(true, "")))
}

func (h *Host) refusal(record map[string]any) string {
	low, high := domain.Int(record["protocol_min"]), domain.Int(record["protocol_max"])
	switch {
	case low > high:
		return "protocol_invalid"
	case record["environment"] != h.options.Environment:
		return "environment_mismatch"
	case record["contract_digest"] != domain.ContractDigest():
		return "contract_unsupported"
	case low > int64(h.options.ProtocolMax) || high < domain.Version:
		return "version_unsupported"
	}
	return ""
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
