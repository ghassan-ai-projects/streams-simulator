package host

import "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"

func (h *Host) recordEvidence(id string, step Scripted, observed bool) {
	outcome := "not_observed"
	if observed {
		outcome = "observed"
	}
	entry := map[string]any{"command_id": id, "device": step.Device, "boot_id": h.issuedBoot[id],
		"metric": step.Verify.Metric, "expected": step.Verify.Equals, "tolerance": step.Verify.Tolerance, "outcome": outcome}
	if observed {
		entry["observed_value"] = h.devices[step.Device].values[step.Verify.Metric]
	}
	h.evidence = append(h.evidence, entry)
}

func (h *Host) barrier(seq uint64) domain.Record {
	fields := map[string]any{"through_seq": seq}
	if len(h.evidence) > 0 {
		count := min(128, len(h.evidence))
		fields["command_evidence"] = h.evidence[:count]
		h.evidence = h.evidence[count:]
	}
	return control("HSYNC", fields)
}
