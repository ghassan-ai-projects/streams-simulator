package domain

import (
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	alarmUnacknowledged = 1
	alarmAcknowledged   = 2
)

type kpiTrack struct {
	state    string
	alarm    bool
	priority string

	samples   int
	sum       float64
	low, high float64

	activations []int64
	activeSince int64
	activeNS    int64
	longestNS   int64
}

func (k *kpiTrack) sample(at int64, value float64) {
	if k.alarm {
		k.sampleAlarm(at, value == alarmUnacknowledged || value == alarmAcknowledged)
		return
	}
	if k.samples == 0 {
		k.low, k.high = value, value
	}
	k.samples++
	k.sum += value
	k.low, k.high = math.Min(k.low, value), math.Max(k.high, value)
}

func (k *kpiTrack) sampleAlarm(at int64, active bool) {
	switch was := k.activeSince >= 0; {
	case active && !was:
		if k.samples > 0 {
			k.activations = append(k.activations, at)
		}
		k.activeSince = at
	case !active && was:
		k.activeNS, k.longestNS = k.activeNS+at-k.activeSince, max(k.longestNS, at-k.activeSince)
		k.activeSince = -1
	}
	k.samples++
}

func (k *kpiTrack) value() model.KPIValue {
	value := model.KPIValue{State: k.state, Samples: k.samples, Min: k.low, Max: k.high}
	if k.samples > 0 {
		value.Mean = k.sum / float64(k.samples)
	}
	return value
}

func (k *kpiTrack) alarmKPI(endNS int64, retired bool) model.AlarmKPI {
	active, longest := k.activeNS, k.longestNS
	if k.activeSince >= 0 {
		active, longest = active+endNS-k.activeSince, max(longest, endNS-k.activeSince)
	}
	return model.AlarmKPI{State: k.state, Priority: k.priority, Activations: len(k.activations), Chattering: chattering(k.activations),
		ActiveS: float64(active) / secondsPerNS, LongestActiveS: float64(longest) / secondsPerNS, ActiveAtEnd: k.activeSince >= 0 && !retired}
}

func chattering(activations []int64) bool {
	for i := 2; i < len(activations); i++ {
		if activations[i]-activations[i-2] < chatterWindowNS {
			return true
		}
	}
	return false
}
