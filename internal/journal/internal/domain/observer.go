package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

// Opened is the identity of a run, written first to the observer stream.
type Opened struct {
	RunID       string
	SimVersion  string
	Domain      model.ArtifactRef
	Adapter     model.ArtifactRef
	Seed        uint64
	StartTimeNS int64
	EntityIDs   []string
	Mode        string
}

// Call names one effector call.
type Call struct {
	Effector  string
	EntityID  string
	CommandID string
}

// Opened records the run's identity. The observer stream gets what a consumer
// could know; the seed is not among it (MCP_SURFACE section 4.5: with the seed
// and the domain, a consumer could regenerate the fault-free world), so it goes
// to the director stream only.
func (j *Journal) Opened(o Opened) {
	j.observer.Append(TypeRunOpened, o.StartTimeNS, map[string]any{
		"run_id": o.RunID, "sim_version": o.SimVersion, "domain": o.Domain, "adapter": o.Adapter,
		"start_time_ns": o.StartTimeNS, "entity_ids": o.EntityIDs, "mode": o.Mode,
	})
	j.director.Append(TypeRunOpened, o.StartTimeNS, map[string]any{"run_id": o.RunID, "seed": o.Seed})
}

// Observation records a delivered record exactly as the sink received it: the
// adapter's rendered line, with nothing of the native event the consumer was
// not given (its emission number, for one, would show what was dropped).
func (j *Journal) Observation(atNS int64, line string) {
	j.observer.Append(TypeObservation, atNS, map[string]any{"line": line})
}

// Malformed records a delivery the consumer received as a broken line, as
// received.
func (j *Journal) Malformed(atNS int64, line string) {
	j.observer.Append(TypeObservation, atNS, map[string]any{"line": line, "malformed": true})
}

// ClockAdvanced records that the world clock moved to toNS.
func (j *Journal) ClockAdvanced(atNS, toNS int64) {
	j.observer.Append(TypeClockAdvanced, atNS, map[string]any{"to_ns": toNS})
}

// Acknowledged records the acknowledgement an effector call returned to the
// operator, exactly as the operator received it.
func (j *Journal) Acknowledged(atNS int64, call Call, ack *model.EffectorAcknowledgement) {
	j.observer.Append(TypeEffectorAck, atNS, map[string]any{
		"command_id": call.CommandID, "effector": call.Effector, "entity_id": call.EntityID, "response": ack,
	})
}

// ConsumerReported records a consumer report. It is a session record: not in
// the command log, so outside the stream digest.
func (j *Journal) ConsumerReported(quiescedThroughNS int64, verdictSubmitted bool) {
	j.observer.Append(TypeConsumerReported, 0, map[string]any{
		"quiesced_through_ns": quiescedThroughNS, "verdict_submitted": verdictSubmitted,
	})
}
