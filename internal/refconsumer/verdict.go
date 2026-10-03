package refconsumer

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (r *Runner) reportProcessedWatermark(now int64) error {
	// Assert processing through the later observed or declared batch end.
	if r.quiescence != nil {
		if err := r.quiescence.ReportQuiesced(now); err != nil {
			return fmt.Errorf("Process: report quiescence: %w", err)
		}
	}
	return nil
}

func (r *Runner) consumerVerdict() *model.Verdict {
	return &model.Verdict{SchemaVersion: "0.1", RunID: r.runID,
		Consumer:   model.ConsumerInfo{Name: "streamsim-refconsumer", Version: "0.1.0", ConfigDigest: r.configDigest()},
		Detections: r.detections, Actions: r.actions, Counters: r.verdictCounters()}
}

func (r *Runner) verdictCounters() map[string]int64 {
	return map[string]int64{"records_seen": r.recordsSeen,
		"detections": int64(len(r.detections)), "actions": int64(len(r.actions))}
}
