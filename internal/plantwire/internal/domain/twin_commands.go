package domain

import (
	"bytes"
	"fmt"
)

const (
	CommandApplied     = "applied"
	CommandNotApplied  = "not_applied"
	CommandTwinRefused = "twin_refused"
)

const (
	NotAppliedInterruptRefused    = "interrupt_refused"
	NotAppliedInterruptUnresolved = "interrupt_unresolved"
	NotAppliedNeverForwarded      = "never_forwarded"
	NotAppliedRefusedByBoard      = "refused_by_board"
	NotAppliedNotAccepted         = "not_accepted"
	NotAppliedBoardFailed         = "board_reported_failure"
)

const OriginNode = "node"

type HostCommand struct {
	ID, Device, Name, Origin string
	Args                     map[string]any
}

type TwinCommand struct {
	ID     string         `json:"command_id"`
	Device string         `json:"device"`
	Name   string         `json:"command"`
	AtNS   int64          `json:"at_ns,omitempty"`
	Origin string         `json:"origin,omitempty"`
	Status string         `json:"status"`
	Reason string         `json:"reason,omitempty"`
	Detail string         `json:"detail,omitempty"`
	Args   map[string]any `json:"-"`
}

type InterruptBlock struct {
	Reason     string
	Unresolved bool
}

func ReadHostCommands(journal []byte) ([]HostCommand, error) {
	var found []HostCommand
	seen := map[string]bool{}
	for number, line := range completeJournalLines(journal) {
		command, ok, err := hostCommandOf(line)
		if err != nil {
			return nil, fmt.Errorf("%w: line %d: %w", ErrTwinJournal, number+1, err)
		}
		if ok && !seen[command.ID] {
			seen[command.ID] = true
			found = append(found, command)
		}
	}
	return found, nil
}

func completeJournalLines(raw []byte) [][]byte {
	end := bytes.LastIndexByte(raw, '\n')
	if end < 0 {
		return nil
	}
	return bytes.Split(raw[:end], []byte("\n"))
}

func hostCommandOf(line []byte) (HostCommand, bool, error) {
	if !bytes.Contains(line, []byte(`"DCMD"`)) {
		return HostCommand{}, false, nil
	}
	record, err := hostWireRecord(line)
	if err != nil || record == nil {
		return HostCommand{}, false, err
	}
	in, err := InboundOf(record)
	if err != nil || in.Invalid != nil || in.Type != "DCMD" {
		return HostCommand{}, false, nil
	}
	return HostCommand{ID: in.CommandID, Device: in.Device, Name: in.Command, Args: in.Args}, true, nil
}

func hostWireRecord(line []byte) (map[string]any, error) {
	var entry map[string]any
	if err := decode(line, &entry); err != nil {
		return nil, err
	}
	record, _ := entry["record"].(map[string]any)
	if text(entry, "type") != "wire" || text(entry, "direction") != TraceHost {
		record = nil
	}
	return record, nil
}

func NodeStops(log []map[string]any) []HostCommand {
	var found []HostCommand
	declared, seen := map[string]string{}, map[string]bool{}
	for _, record := range log {
		found = appendNodeStop(found, InterruptEntry(record), declared, seen)
	}
	return found
}

func appendNodeStop(found []HostCommand, entry InterruptEntry, declared map[string]string, seen map[string]bool) []HostCommand {
	device, id := entry.Text("device"), entry.Text("command_id")
	switch {
	case entry.Event() == EventStop:
		declared[device] = entry.Text("command")
	case entry.Event() == EventEffect && id != "none" && !seen[id]:
		seen[id] = true
		return append(found, HostCommand{ID: id, Device: device, Name: declared[device], Origin: OriginNode, Args: map[string]any{}})
	}
	return found
}

func InterruptBlocks(log []map[string]any) map[string]InterruptBlock {
	blocks := map[string]InterruptBlock{}
	for _, record := range log {
		entry := InterruptEntry(record)
		switch entry.Event() {
		case EventVerdict:
			blockVerdict(blocks, entry)
		case EventResolve:
			blockResolution(blocks, entry)
		case EventPurge:
			blocks[entry.Text("command_id")] = InterruptBlock{Reason: entry.Text("reason")}
		}
	}
	return blocks
}

func blockVerdict(blocks map[string]InterruptBlock, entry InterruptEntry) {
	id := entry.Text("command_id")
	switch entry.Text("verdict") {
	case VerdictRefuse:
		blocks[id] = InterruptBlock{Reason: entry.Text("reason")}
	case VerdictQueue:
		blocks[id] = InterruptBlock{Reason: "queued", Unresolved: true}
	default:
		delete(blocks, id)
	}
}

func blockResolution(blocks map[string]InterruptBlock, entry InterruptEntry) {
	ids := append([]string{entry.Text("command_id")}, stringsOf(entry["waiting"])...)
	for _, id := range ids {
		if entry.Text("resolution") == ResolutionRefuse {
			blocks[id] = InterruptBlock{Reason: entry.Text("reason")}
		} else {
			delete(blocks, id)
		}
	}
}

func ClassifyCommands(commands []HostCommand, board BoardTrace, blocks map[string]InterruptBlock) []TwinCommand {
	classified := []TwinCommand{}
	for _, command := range commands {
		if board.Has(command.Device) {
			classified = append(classified, classifyCommand(command, board, blocks))
		}
	}
	return classified
}

func classifyCommand(command HostCommand, board BoardTrace, blocks map[string]InterruptBlock) TwinCommand {
	receipt := board.Receipts[command.ID]
	block, blocked := blocks[command.ID]
	view := TwinCommand{ID: command.ID, Device: command.Device, Name: command.Name, Args: command.Args, Origin: command.Origin, Status: CommandNotApplied}
	view.AtNS = max(receipt.AtNS, board.StartNS)
	view.Reason, view.Detail = notAppliedReason(block, blocked, receipt)
	if view.Reason == "" {
		view.Status = CommandApplied
	}
	return view
}

func notAppliedReason(block InterruptBlock, blocked bool, receipt BoardReceipt) (reason, detail string) {
	if blocked {
		return interruptReason(block)
	}
	return receiptReason(receipt)
}

func interruptReason(block InterruptBlock) (reason, detail string) {
	if block.Unresolved {
		return NotAppliedInterruptUnresolved, block.Reason
	}
	return NotAppliedInterruptRefused, block.Reason
}

func receiptReason(receipt BoardReceipt) (reason, detail string) {
	switch {
	case receipt.Status == "":
		return NotAppliedNeverForwarded, ""
	case receipt.Status == "refused":
		return NotAppliedRefusedByBoard, receipt.Reason
	case receipt.Status != "accepted":
		return NotAppliedNotAccepted, receipt.Status
	case receipt.Outcome == "failed":
		return NotAppliedBoardFailed, ""
	}
	return "", ""
}
