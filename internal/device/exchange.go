package device

import (
	"fmt"
	"log/slog"
)

func serveDeviceFrame(d *Device, gate *wireGate, frame []byte) error {
	// query_state is gateway-link control, separate from device wire records.
	if isQueryState(frame) {
		return sendEncodedRecord(gate, d.State(), "state")
	}
	// Execution truth is read via state; results report terminal status only.
	outcome, err := d.handleCommand(frame)
	if err != nil {
		return sendMalformedExchange(gate)
	}
	return sendCommandOutcome(d, gate, frame, outcome)
}

func sendMalformedExchange(gate *wireGate) error {
	// Malformed commands receive an ordered rejection receipt/result pair.
	if err := sendEncodedRecord(gate, malformedReceipt(), "malformed receipt"); err != nil {
		return err
	}
	return sendEncodedRecord(gate, malformedResult(), "malformed result")
}

func sendCommandOutcome(d *Device, gate *wireGate, frame []byte, outcome Outcome) error {
	logCommandOutcome(outcome)
	receipt, result, err := encodeCommandOutcome(outcome)
	if err != nil {
		return err
	}
	if err := replayDuplicateCommand(d, frame, outcome); err != nil {
		return err
	}
	return deliverCommandOutcome(gate, receipt, result, outcome)
}

func sendEncodedRecord(gate *wireGate, record map[string]any, kind string) error {
	frame, err := EncodeRecord(record)
	if err != nil {
		return fmt.Errorf("device: encode %s: %w", kind, err)
	}
	return gate.send(frame)
}

func logCommandOutcome(outcome Outcome) {
	if outcome.Fault != "" {
		slog.Info("device fault injected", "fault", outcome.Fault, "accepted_command", outcome.AcceptedCommand)
	}
	accepted, _ := outcome.Receipt["accepted"].(bool)
	if !accepted {
		slog.Info("device command rejected", "command_id", outcome.Receipt["command_id"], "reject_code", outcome.Receipt["reject_code"])
	}
}

func encodeCommandOutcome(outcome Outcome) ([]byte, []byte, error) {
	receipt, err := EncodeRecord(outcome.Receipt)
	if err != nil {
		return nil, nil, fmt.Errorf("device: encode receipt: %w", err)
	}
	result, err := EncodeRecord(outcome.Result)
	if err != nil {
		return nil, nil, fmt.Errorf("device: encode result: %w", err)
	}
	return receipt, result, nil
}

func replayDuplicateCommand(d *Device, frame []byte, outcome Outcome) error {
	if outcome.Duplicate {
		// Exercise idempotency internally; retain one protocol exchange.
		if _, err := d.handleCommand(frame); err != nil {
			return fmt.Errorf("device: duplicate command: %w", err)
		}
	}
	return nil
}

func deliverCommandOutcome(gate *wireGate, receipt, result []byte, outcome Outcome) error {
	if outcome.AckLost {
		return nil
	}
	if err := gate.send(receipt); err != nil {
		return err
	}
	if err := gate.send(result); err != nil {
		return err
	}
	if outcome.Disconnect {
		return ErrInjectedDisconnect
	}
	return nil
}
