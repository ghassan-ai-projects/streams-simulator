package domain

import (
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var differences = []difference{
	{"different_mode", func(a, b RunInputs) bool { return a.Mode != b.Mode }, func(a, b RunInputs) string {
		return "The runs were made in different modes (" + a.Mode + " and " + b.Mode + ")."
	}},
	{"different_simulator_version", func(a, b RunInputs) bool { return a.SimVersion != b.SimVersion }, func(a, b RunInputs) string {
		return "The runs were made by different simulator versions (" + a.SimVersion + " and " + b.SimVersion + ")."
	}},
	{"different_domain", func(a, b RunInputs) bool { return idOf(a.Domain) != idOf(b.Domain) }, func(a, b RunInputs) string {
		return "The runs use different domains (" + idOf(a.Domain) + " and " + idOf(b.Domain) + ")."
	}},
	{"different_domain_digest", func(a, b RunInputs) bool {
		return idOf(a.Domain) == idOf(b.Domain) && digestOf(a.Domain) != digestOf(b.Domain)
	}, func(a, b RunInputs) string {
		return "The runs use the domain " + idOf(a.Domain) + " in different versions: the model changed between them."
	}},
	{"different_adapter", func(a, b RunInputs) bool {
		return idOf(a.Adapter) != idOf(b.Adapter) || digestOf(a.Adapter) != digestOf(b.Adapter)
	},
		func(a, b RunInputs) string {
			return "The runs use different adapters (" + idOf(a.Adapter) + " and " + idOf(b.Adapter) + ")."
		}},
	{"different_seed", func(a, b RunInputs) bool { return seedText(a.Seed) != seedText(b.Seed) }, func(a, b RunInputs) string {
		return "The runs use different seeds (" + seedText(a.Seed) + " and " + seedText(b.Seed) + "), so the randomness differs."
	}},
	{"different_time_mode", func(a, b RunInputs) bool { return a.TimeMode != b.TimeMode }, func(a, b RunInputs) string {
		return "The runs use different time modes (" + a.TimeMode + " and " + b.TimeMode + ")."
	}},
	{"different_start_time", func(a, b RunInputs) bool { return a.StartTimeNS != b.StartTimeNS }, func(a, b RunInputs) string {
		return "The runs start at different simulated instants, so absolute times in the plan differ."
	}},
	{"different_requested_duration", func(a, b RunInputs) bool { return requestedLength(a) != requestedLength(b) }, func(a, b RunInputs) string {
		return "The runs were asked for different durations (" + requestedSeconds(a) + " and " + requestedSeconds(b) + ")."
	}},
	{"different_simulated_length", func(a, b RunInputs) bool { return a.SimulatedLengthNS != b.SimulatedLengthNS }, func(a, b RunInputs) string {
		return "The runs cover different simulated lengths (" + seconds(a.SimulatedLengthNS) + " and " + seconds(b.SimulatedLengthNS) + "); whole-run results are not over the same span."
	}},
}

func idOf(ref *model.ArtifactRef) string {
	if ref == nil {
		return "unknown"
	}
	return ref.ID
}

func digestOf(ref *model.ArtifactRef) string {
	if ref == nil {
		return ""
	}
	return ref.Digest
}

func requestedLength(in RunInputs) int64 {
	if in.RequestedEndNS == 0 {
		return 0
	}
	return in.RequestedEndNS - in.StartTimeNS
}

func requestedSeconds(in RunInputs) string {
	if in.RequestedEndNS == 0 {
		return "not recorded"
	}
	return seconds(requestedLength(in))
}

func seconds(ns int64) string {
	return strconv.FormatFloat(float64(ns)/1e9, 'f', -1, 64) + "s"
}
