package domain

import (
	"fmt"
	"strings"
)

const scopeDomain = "domain"

func notModeled(assumptions *Assumptions) []NotModeled {
	entries := domainLimits(assumptions)
	for _, kind := range assumptions.Kinds {
		entries = append(entries, kindLimits(kind)...)
	}
	return entries
}

func domainLimits(assumptions *Assumptions) []NotModeled {
	entries := []NotModeled{}
	if assumptions.Origin.Kind == OriginTwin {
		entries = append(entries, twinLimits()...)
	}
	if usesLibraryInstrument(assumptions) {
		entries = append(entries, NotModeled{"instrument_semantics_unchecked", scopeDomain,
			"A library instrument was checked for unit and range only; that it measures what the reading means is not checked."})
	}
	return append(entries, NotModeled{"no_calibration_evidence", scopeDomain,
		"No calibration or hardware validation is attached; a passing check shows a valid interface and plausible output, not that the physics match."})
}

func twinLimits() []NotModeled {
	return []NotModeled{
		{"kind_file_members_not_reproduced", scopeDomain, "A twin reproduces only the kind, its metrics (name, data_type, unit, eu_range, resolution, writable) and its commands (name, args_schema, risk_class, safe_stop); any other member of a device kind file is not reproduced."},
		{"first_order_response_only", scopeDomain, "A command moves a reading directly or through a first-order lag; there is no dead time, second-order response, hysteresis or non-linearity, and a reading saturates only at its eu_range."},
		{"device_behavior_not_modeled", scopeDomain, "Device firmware, scheduling and wire timing are not modeled; the twin answers at its own acknowledgement latency, not the device's."},
	}
}

func usesLibraryInstrument(assumptions *Assumptions) bool {
	for _, kind := range assumptions.Kinds {
		for _, reading := range kind.Readings {
			if reading.Basis == BasisLibraryInstrument {
				return true
			}
		}
	}
	return false
}

func kindLimits(kind KindAssumptions) []NotModeled {
	entries := []NotModeled{}
	for _, reading := range kind.Readings {
		entries = append(entries, readingLimits(kind.Kind, reading)...)
	}
	for _, dynamics := range kind.Dynamics {
		entries = append(entries, dynamicsLimits(kind.Kind, dynamics)...)
	}
	for _, command := range kind.Commands {
		entries = append(entries, commandLimits(kind.Kind, command)...)
	}
	return entries
}

func readingLimits(kind string, reading ReadingAssumption) []NotModeled {
	scope := scopeOf(kind, reading.Channel)
	if reading.Noise.Sigma != 0 {
		return nil
	}
	return []NotModeled{{"no_measurement_noise", scope, fmt.Sprintf("%s reaches the consumer exactly as the true value: its noise sigma is 0.", scope)}}
}

func dynamicsLimits(kind string, dynamics DynamicsAssumption) []NotModeled {
	scope := scopeOf(kind, dynamics.Target)
	if dynamics.Form != formFirstOrderLag || dynamics.Clamp != nil {
		return nil
	}
	return []NotModeled{{"saturation_not_declared", scope, fmt.Sprintf("%s follows its inputs without a clamp: no saturation or physical limit is declared.", scope)}}
}

func commandLimits(kind string, command CommandAssumption) []NotModeled {
	scope, entries := scopeOf(kind, command.Name), []NotModeled{}
	if command.Response == ResponseNone {
		entries = append(entries, NotModeled{"command_changes_nothing", scope, scope + " is acknowledged and changes nothing in the twin."})
	}
	if command.SafeStop && command.Effect.TimeConstantS == 0 && command.Effect.DeadTimeS == 0 {
		entries = append(entries, NotModeled{"safe_stop_is_a_step", scope, scope + " takes effect as a step, with no run-down time."})
	}
	if neverFails(command.Acknowledgement) {
		entries = append(entries, NotModeled{"acknowledgement_never_fails", scope, scope + " is always acknowledged ok; no refusal or timeout is modeled."})
	}
	return entries
}

func neverFails(ack AcknowledgementValue) bool {
	for _, mode := range ack.FailureModes {
		if mode.Mode != "ok" && mode.Probability > 0 {
			return false
		}
	}
	return true
}

func scopeOf(kind, name string) string {
	if strings.HasPrefix(name, kind+".") {
		return name
	}
	return kind + "." + name
}
