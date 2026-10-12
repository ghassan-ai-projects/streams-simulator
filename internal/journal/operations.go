package journal

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/journal/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Read decodes every complete line of a journal and ignores an unterminated
// last line. It stops at the first line that is not a record and returns the
// records before it with the error.
func Read(raw []byte) ([]Record, error) {
	return layer.Read(raw)
}

// Digest is the SHA-256 of the lines of a journal that replay can reproduce:
// session records and a torn last line do not change it.
func Digest(raw []byte) string {
	return layer.Digest(raw)
}

// FirstDifference is the index, among the replayable records, of the first one
// where two journals differ or one ends early; -1 when they agree.
func FirstDifference(a, b []byte) int {
	return layer.FirstDifference(a, b)
}

// Opened records the run's identity.
func (j *Journal) Opened(o Opened) {
	j.impl.Opened(o)
}

// Observation records a delivered record exactly as the sink received it.
func (j *Journal) Observation(atNS int64, line string) {
	j.impl.Observation(atNS, line)
}

// Malformed records a delivery the consumer received as a broken line.
func (j *Journal) Malformed(atNS int64, line string) {
	j.impl.Malformed(atNS, line)
}

// ClockAdvanced records that the world clock moved to toNS.
func (j *Journal) ClockAdvanced(atNS, toNS int64) {
	j.impl.ClockAdvanced(atNS, toNS)
}

// Acknowledged records the acknowledgement an effector call returned to the
// operator.
func (j *Journal) Acknowledged(atNS int64, call Call, ack *model.EffectorAcknowledgement) {
	j.impl.Acknowledged(atNS, call, ack)
}

// ConsumerReported records a consumer report (a session record, outside the
// stream digest).
func (j *Journal) ConsumerReported(quiescedThroughNS int64, verdictSubmitted bool) {
	j.impl.ConsumerReported(quiescedThroughNS, verdictSubmitted)
}

// Command records a command-log entry.
func (j *Journal) Command(atNS int64, command model.Command) {
	j.impl.Command(atNS, command)
}

// Delivery records a delivery-ledger row.
func (j *Journal) Delivery(atNS int64, row model.LedgerRecord) {
	j.impl.Delivery(atNS, row)
}

// State records the world-state snapshot taken at an emission.
func (j *Journal) State(atNS int64, snapshot model.StateSnapshot) {
	j.impl.State(atNS, snapshot)
}

// Outcome records the world's record of an effector call: its result, or the
// refusal when the call was declined.
func (j *Journal) Outcome(atNS int64, call Call, result any, refusal string) {
	j.impl.Outcome(atNS, call, result, refusal)
}

// Unblinded records that the run was unblinded and why.
func (j *Journal) Unblinded(atNS int64, reason string) {
	j.impl.Unblinded(atNS, reason)
}

// Flush pushes both streams to their files at a command boundary and returns
// the first failure either has met.
func (j *Journal) Flush() error {
	return j.impl.Flush()
}

// Close writes run.closed to both streams, syncs and closes them, and returns
// their digests with the first failure.
func (j *Journal) Close(atNS int64, incomplete bool, runError string) (Digests, error) {
	return j.impl.Close(atNS, incomplete, runError)
}

// Digests returns the digests of the replayable lines written so far.
func (j *Journal) Digests() Digests {
	return j.impl.Digests()
}

// Bytes returns the retained lines of the named stream (nil unless retained).
func (j *Journal) Bytes(stream string) ([]byte, error) {
	return j.impl.Bytes(stream)
}
