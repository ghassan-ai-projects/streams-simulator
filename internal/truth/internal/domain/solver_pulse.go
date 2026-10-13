package domain

import (
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (scan *observabilityScan) judgedAtRecordingsWhen(pulse bool, onsetNS int64) *observabilityScan {
	if pulse {
		scan.judgeAtRecordings(onsetNS)
	}
	return scan
}

func (scan *observabilityScan) judgeAtRecordings(onsetNS int64) {
	scan.recorded = true
	channels := detectorChannels(scan.detector)
	scan.faulted.SetEmitter(func(event model.SimEvent) {
		at, err := model.ParseTime(event.EventTime)
		if err != nil || event.EntityID != scan.entity || !slices.Contains(channels, event.Channel) || at < onsetNS || scan.res.unavoidableFound {
			return
		}
		scan.observeThresholds(at)
	})
}
