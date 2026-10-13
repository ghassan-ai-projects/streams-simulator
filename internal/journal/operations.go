package journal

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/journal/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func Read(raw []byte) ([]Record, error) {
	return layer.Read(raw)
}

func Digest(raw []byte) string {
	return layer.Digest(raw)
}

func FirstDifference(a, b []byte) int {
	return layer.FirstDifference(a, b)
}

func (j *Journal) Opened(o Opened) {
	j.impl.Opened(o)
}

func (j *Journal) Observation(atNS int64, line string) {
	j.impl.Observation(atNS, line)
}

func (j *Journal) Framing(atNS int64, kind, line string) {
	j.impl.Framing(atNS, kind, line)
}

func (j *Journal) Malformed(atNS int64, line string) {
	j.impl.Malformed(atNS, line)
}

func (j *Journal) ClockAdvanced(atNS, toNS int64) {
	j.impl.ClockAdvanced(atNS, toNS)
}

func (j *Journal) Acknowledged(atNS int64, call Call, ack *model.EffectorAcknowledgement) {
	j.impl.Acknowledged(atNS, call, ack)
}

func (j *Journal) ConsumerReported(quiescedThroughNS int64, verdictSubmitted bool) {
	j.impl.ConsumerReported(quiescedThroughNS, verdictSubmitted)
}

func (j *Journal) Wire(atNS int64, direction string, record map[string]any) {
	j.impl.Wire(atNS, direction, record)
}

func (j *Journal) Pacing(atNS int64, throughSeq uint64, waitedMS int64, missed bool) {
	j.impl.Pacing(atNS, throughSeq, waitedMS, missed)
}

func (j *Journal) Command(atNS int64, command model.Command) {
	j.impl.Command(atNS, command)
}

func (j *Journal) Delivery(atNS int64, row model.LedgerRecord) {
	j.impl.Delivery(atNS, row)
}

func (j *Journal) State(atNS int64, snapshot model.StateSnapshot) {
	j.impl.State(atNS, snapshot)
}

func (j *Journal) KPITruth(atNS int64, truth *model.KPITruth) {
	j.impl.KPITruth(atNS, truth)
}

func (j *Journal) Outcome(atNS int64, call Call, result any, refusal string) {
	j.impl.Outcome(atNS, call, result, refusal)
}

func (j *Journal) Unblinded(atNS int64, reason string) {
	j.impl.Unblinded(atNS, reason)
}

func (j *Journal) Flush() error {
	return j.impl.Flush()
}

func (j *Journal) Close(atNS int64, incomplete bool, runError string) (Digests, error) {
	return j.impl.Close(atNS, incomplete, runError)
}

func (j *Journal) Digests() Digests {
	return j.impl.Digests()
}

func (j *Journal) Bytes(stream string) ([]byte, error) {
	return j.impl.Bytes(stream)
}
