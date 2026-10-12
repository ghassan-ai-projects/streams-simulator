package domain

import (
	"errors"
	"fmt"
)

type Message struct {
	Type            string         `json:"type"`
	ProtocolVersion int64          `json:"protocol_version"`
	SessionID       string         `json:"session_id"`
	Seq             uint64         `json:"seq"`
	ProtocolMin     int64          `json:"protocol_min"`
	ProtocolMax     int64          `json:"protocol_max"`
	Environment     string         `json:"environment"`
	ContractVersion string         `json:"contract_version"`
	ContractDigest  string         `json:"contract_digest"`
	BDSeq           uint64         `json:"bd_seq"`
	Boards          []Board        `json:"boards"`
	Online          bool           `json:"online"`
	Refusal         string         `json:"refusal"`
	AuthorityEpoch  int64          `json:"authority_epoch"`
	Command         string         `json:"command"`
	CommandID       string         `json:"command_id"`
	IdempotencyKey  string         `json:"idempotency_key"`
	Device          string         `json:"device"`
	Kind            string         `json:"kind"`
	BootID          string         `json:"boot_id"`
	ExpectedBootID  string         `json:"expected_boot_id"`
	PolicyDigest    string         `json:"policy_digest"`
	Args            map[string]any `json:"args"`
	DeadlineNS      int64          `json:"deadline_ns"`
	NotBeforeNS     int64          `json:"not_before_ns"`
	Historical      bool           `json:"historical"`
	Metrics         []Metric       `json:"metrics"`
	Status          string         `json:"status"`
	Reason          string         `json:"reason"`
	ReportedOutcome string         `json:"reported_outcome"`
	ThroughSeq      uint64         `json:"through_seq"`
	WorldTimeNS     int64          `json:"world_time_ns"`
}

type Board struct {
	Board   string   `json:"board"`
	Members []string `json:"members"`
}

type Metric struct {
	Name    string `json:"name"`
	Alias   int64  `json:"alias"`
	Value   any    `json:"value"`
	Quality string `json:"quality"`
}

type Inbound struct {
	Message
	Record      map[string]any
	Invalid     error
	Unsequenced bool
}

func Decode(line []byte) (Inbound, bool) {
	record, invalid := Parse(line)
	if record == nil {
		return Inbound{}, false
	}
	seq, sequenced := SeqOf(record["seq"])
	in := Inbound{Record: record, Invalid: invalid, Unsequenced: !sequenced,
		Message: Message{Type: text(record, "type"), Seq: seq, CommandID: text(record, "command_id")}}
	if invalid == nil {
		in.Invalid = errors.Join(decode(line, &in.Message), sized(in.Type, len(line)))
	}
	return in, true
}

const (
	MaxRecordBytes = 64 << 10
	MaxBirthBytes  = 1 << 20
)

var (
	ErrTooLarge   = errors.New("plantwire: record larger than its limit")
	ErrNotSession = errors.New("plantwire: record from another session or version")
)

func sized(kind string, size int) error {
	if size > MaxBirthBytes || (size > MaxRecordBytes && kind != "NBIRTH" && kind != "DBIRTH") {
		return fmt.Errorf("%w: %s of %d bytes", ErrTooLarge, kind, size)
	}
	return nil
}

func (in *Inbound) Within(session string, version int) {
	if in.Invalid == nil && (in.SessionID != session || in.ProtocolVersion != int64(version)) {
		in.Invalid = fmt.Errorf("%w: %s %s v%d", ErrNotSession, in.Type, in.SessionID, in.ProtocolVersion)
	}
}

func text(record map[string]any, key string) string {
	value, _ := record[key].(string)
	return value
}
