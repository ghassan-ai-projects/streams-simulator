package domain

import (
	"fmt"
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonvalue"
)

func (l *Live) command(id string) *LiveCommand {
	index := slices.IndexFunc(l.commands, func(c *LiveCommand) bool { return c.CommandID == id })
	if index < 0 {
		return nil
	}
	return l.commands[index]
}

func (l *Live) sent(r map[string]any, at int64) bool {
	if l.command(text(r, "command_id")) != nil {
		return false
	}
	c := &LiveCommand{CommandID: text(r, "command_id"), Device: text(r, "device"), Command: text(r, "command"),
		BootID: text(r, "expected_boot_id"), Sent: true, Receipt: "pending", Report: "pending", Effect: "unknown",
		SentAtNS: at, UpdatedAtNS: at}
	l.commands = append(l.commands, c)
	if len(l.commands) > maxLiveCommands {
		l.commands = slices.Clone(l.commands[1:])
		l.truncated = true
	}
	l.mark(at, "command", "Command sent", c.Device)
	return true
}

func (l *Live) answer(r map[string]any, at int64) bool {
	c := l.command(text(r, "command_id"))
	if c == nil || c.BootID != text(r, "boot_id") {
		return false
	}
	if text(r, "type") == "DRESULT" {
		return l.result(c, r, at)
	}
	return l.receipt(c, r, at)
}

func (l *Live) receipt(c *LiveCommand, r map[string]any, at int64) bool {
	status := text(r, "status")
	if c.Receipt != "pending" || (status != "accepted" && status != "refused") {
		return false
	}
	c.Receipt, c.Reason, c.UpdatedAtNS = status, text(r, "reason"), at
	l.mark(at, "command", "Command "+status, c.Device)
	return true
}

func (l *Live) result(c *LiveCommand, r map[string]any, at int64) bool {
	if c.Receipt != "accepted" || c.Report != "pending" {
		return false
	}
	c.Report, c.UpdatedAtNS = text(r, "reported_outcome"), at
	l.mark(at, "command", "Board reported "+c.Report, c.Device)
	return true
}

func (l *Live) evidence(r map[string]any, at int64) bool {
	changed := false
	for _, entry := range objects(r["command_evidence"]) {
		changed = l.confirm(entry, at) || changed
	}
	return changed
}

func (l *Live) confirm(entry map[string]any, at int64) bool {
	c := l.command(text(entry, "command_id"))
	if !l.evidenceMatches(c, entry) {
		return false
	}
	c.EffectEvidence = fmt.Sprintf("Host checked %s: %s", text(entry, "metric"), text(entry, "outcome"))
	if text(entry, "outcome") == "observed" {
		c.Effect = "observed"
	}
	c.UpdatedAtNS = at
	l.mark(at, "command", c.EffectEvidence, c.Device)
	return true
}

func (l *Live) evidenceMatches(c *LiveCommand, e map[string]any) bool {
	if c == nil || c.Receipt != "accepted" || c.Report != "succeeded" || c.EffectEvidence != "" {
		return false
	}
	d := l.devices[c.Device]
	return d != nil && d.Health == "online" && d.BootID == c.BootID && c.Device == text(e, "device") &&
		c.BootID == text(e, "boot_id") && text(e, "metric") != "" &&
		validEffectEvidence(e)
}

func validEffectEvidence(e map[string]any) bool {
	expected, expectedOK := jsonvalue.Number(e["expected"])
	tolerance, toleranceOK := jsonvalue.Number(e["tolerance"])
	if !expectedOK || !toleranceOK || tolerance < 0 || math.IsInf(expected,0) || math.IsNaN(expected) {
		return false
	}
	if text(e,"outcome") == "not_observed" {
		return true
	}
	value, valueOK := jsonvalue.Number(e["observed_value"])
	return text(e,"outcome") == "observed" && valueOK && math.Abs(value-expected) <= tolerance
}
