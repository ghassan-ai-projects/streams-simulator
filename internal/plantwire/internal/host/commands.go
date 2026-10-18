package host

import (
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

const commandLifetimeNS = int64(120e9)

func (h *Host) sync(m domain.Message) []map[string]any {
	now := m.WorldTimeNS
	h.verify(now)
	var records []domain.Record
	for len(h.script) > 0 && h.script[0].AtNS <= now {
		if command, sent := h.issue(h.script[0], now); sent {
			records = append(records, command)
		}
		h.script = h.script[1:]
	}
	records = append(records, h.barrier(m.Seq))
	return h.reply(records...)
}

func (h *Host) issue(step Scripted, now int64) (domain.Record, bool) {
	id := step.ID
	h.report.Commands[id] = &Status{Receipts: []string{}}
	d := h.devices[step.Device]
	if d == nil {
		h.report.Commands[id].Denied = "not_commissioned"
		return domain.Record{}, false
	}
	h.issued[id], h.issuedAt[id], h.order = step, now, append(h.order, id)
	h.issuedBoot[id] = d.boot
	key := domain.Command{Device: step.Device, Name: step.Command, Args: step.Args, Boot: d.boot}.Content()
	return control("DCMD", map[string]any{"command_id": id, "idempotency_key": key, "device": step.Device, "command": step.Command,
		"args": step.Args, "expected_boot_id": d.boot, "deadline_ns": now + commandLifetimeNS, "not_before_ns": 0,
		"authority_epoch": h.options.Epoch, "policy_digest": h.options.Policy, "fallback_on_lease_loss": "safe_off"}), true
}

func (h *Host) answer(m domain.Message) {
	status := h.report.Commands[m.CommandID]
	if status == nil || m.BootID != h.issuedBoot[m.CommandID] {
		return
	}
	if m.Type == "DRECEIPT" {
		status.Receipts = append(status.Receipts, receipt(m))
		return
	}
	if status.Result == "" && slices.Contains(status.Receipts, "accepted") {
		status.Result = m.ReportedOutcome
		h.resultAt[m.CommandID] = h.samples
	}
}

func receipt(m domain.Message) string {
	if m.Reason != "" {
		return m.Status + ":" + m.Reason
	}
	return m.Status
}

func (h *Host) verify(now int64) {
	for _, id := range h.order {
		step, status := h.issued[id], h.report.Commands[id]
		if step.Verify == nil || status.Verified != nil || status.Result != "succeeded" {
			continue
		}
		observed := h.observedAfterResult(id, step)
		if observed || now-h.issuedAt[id] >= step.Verify.WithinNS {
			status.Verified = &observed
			h.recordEvidence(id, step, observed)
		}
	}
}

func (h *Host) observedAfterResult(id string, step Scripted) bool {
	d := h.devices[step.Device]
	return d != nil && d.boot == h.issuedBoot[id] && d.valid[step.Verify.Metric] && d.updated[step.Verify.Metric] > h.resultAt[id] &&
		math.Abs(d.values[step.Verify.Metric]-step.Verify.Equals) <= step.Verify.Tolerance
}
