package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

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

type Call struct {
	Effector  string
	EntityID  string
	CommandID string
}

func (j *Journal) Opened(o Opened) {
	j.observer.Append(TypeRunOpened, o.StartTimeNS, map[string]any{
		"run_id": o.RunID, "sim_version": o.SimVersion, "domain": o.Domain, "adapter": o.Adapter,
		"start_time_ns": o.StartTimeNS, "entity_ids": o.EntityIDs, "mode": o.Mode,
	})
	j.director.Append(TypeRunOpened, o.StartTimeNS, map[string]any{"run_id": o.RunID, "seed": o.Seed})
}

func (j *Journal) Observation(atNS int64, line string) {
	j.observer.Append(TypeObservation, atNS, map[string]any{"line": line})
}

func (j *Journal) Framing(atNS int64, kind, line string) {
	j.observer.Append(TypeObservation, atNS, map[string]any{"line": line, "framing": kind})
}

func (j *Journal) Malformed(atNS int64, line string) {
	j.observer.Append(TypeObservation, atNS, map[string]any{"line": line, "malformed": true})
}

func (j *Journal) ClockAdvanced(atNS, toNS int64) {
	j.observer.Append(TypeClockAdvanced, atNS, map[string]any{"to_ns": toNS})
}

func (j *Journal) Acknowledged(atNS int64, call Call, ack *model.EffectorAcknowledgement) {
	j.observer.Append(TypeEffectorAck, atNS, map[string]any{
		"command_id": call.CommandID, "effector": call.Effector, "entity_id": call.EntityID, "response": ack,
	})
}

func (j *Journal) ConsumerReported(quiescedThroughNS int64, verdictSubmitted bool) {
	j.observer.Append(TypeConsumerReported, 0, map[string]any{
		"quiesced_through_ns": quiescedThroughNS, "verdict_submitted": verdictSubmitted,
	})
}

func (j *Journal) Wire(atNS int64, direction string, record map[string]any) {
	j.observer.Append(TypeWire, atNS, map[string]any{"direction": direction, "record": record})
}

func (j *Journal) Origin(atNS int64, fields map[string]any) {
	j.observer.Append(TypeOrigin, atNS, fields)
}

func (j *Journal) Binding(atNS int64, fields map[string]any) {
	j.observer.Append(TypeBinding, atNS, fields)
}

func (j *Journal) Pacing(atNS int64, throughSeq uint64, waitedMS int64, missed bool) {
	j.observer.Append(TypePacing, atNS, map[string]any{"through_seq": throughSeq, "waited_ms": waitedMS, "missed": missed})
}
