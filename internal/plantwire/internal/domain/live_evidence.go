package domain

import (
	"fmt"
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonvalue"
)

func finiteNumber(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func WithinTolerance(observed, expected, tolerance float64) bool {
	return finiteNumber(observed) && finiteNumber(expected) && finiteNumber(tolerance) && tolerance >= 0 && math.Abs(observed-expected) <= tolerance
}

func (l *Live) evidence(r map[string]any, at int64) bool {
	entries := objects(r["command_evidence"])
	changed := false
	for _, entry := range entries[:min(maxEvidence, len(entries))] {
		changed = l.confirm(entry, at) || changed
	}
	return changed
}

func (l *Live) confirm(entry map[string]any, at int64) bool {
	c := l.command(text(entry, "command_id"))
	if !awaitsEvidence(c, entry) {
		return false
	}
	d := l.devices[c.Device]
	if d == nil || !d.up() || d.BootID != c.BootID {
		return l.unapplied(c, at)
	}
	if !l.plausibleEvidence(d, entry) {
		return false
	}
	l.recordVerdict(c, d, entry, at)
	return true
}

func (l *Live) unapplied(c *LiveCommand, at int64) bool {
	l.mark(at, "command", "command", c.Device+" command "+c.Command+" evidence not applied: device offline or boot changed", c.Device)
	return true
}

func (l *Live) recordVerdict(c *LiveCommand, d *LiveDevice, e map[string]any, at int64) {
	outcome := text(e, "outcome")
	if outcome == "observed" && !heldReadingAgrees(d, text(e, "metric"), e) {
		outcome = "unverified"
	}
	c.EffectEvidence, c.UpdatedAtNS = fmt.Sprintf("Host checked %s: %s", text(e, "metric"), outcome), at
	if outcome == "observed" {
		c.Effect = "observed"
	}
	l.mark(at, "command", "command", c.Device+" command "+c.Command+": "+c.EffectEvidence, c.Device)
}

func awaitsEvidence(c *LiveCommand, e map[string]any) bool {
	return c != nil && c.Receipt == "accepted" && c.Report == "succeeded" && c.EffectEvidence == "" &&
		c.Device == text(e, "device") && c.BootID == text(e, "boot_id") && text(e, "metric") != ""
}

func (l *Live) plausibleEvidence(d *LiveDevice, e map[string]any) bool {
	template := l.kinds[d.kind][text(e, "metric")]
	expected, expectedOK := jsonvalue.Number(e["expected"])
	tolerance, toleranceOK := jsonvalue.Number(e["tolerance"])
	if template == nil || !expectedOK || !toleranceOK || !finiteNumber(expected) || !finiteNumber(tolerance) || tolerance < 0 || tolerance > span(template) {
		return false
	}
	return outcomePlausible(e, expected, tolerance)
}

func outcomePlausible(e map[string]any, expected, tolerance float64) bool {
	value, valueOK := jsonvalue.Number(e["observed_value"])
	switch text(e, "outcome") {
	case "not_observed":
		return true
	case "observed":
		return valueOK && WithinTolerance(value, expected, tolerance)
	}
	return false
}

func span(template map[string]any) float64 {
	bounds, ok := template["eu_range"].(map[string]any)
	low, lowOK := jsonvalue.Number(bounds["min"])
	high, highOK := jsonvalue.Number(bounds["max"])
	if !ok || !lowOK || !highOK {
		return math.MaxFloat64
	}
	return high - low
}

func heldReadingAgrees(d *LiveDevice, metric string, e map[string]any) bool {
	observed, _ := jsonvalue.Number(e["observed_value"])
	tolerance, _ := jsonvalue.Number(e["tolerance"])
	for _, reading := range d.Readings {
		held, numeric := jsonvalue.Number(reading.Value)
		if reading.Name == metric && reading.State == "good" && numeric {
			return WithinTolerance(held, observed, tolerance)
		}
	}
	return false
}
