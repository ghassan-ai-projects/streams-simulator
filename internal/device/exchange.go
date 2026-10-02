package device

import (
	"fmt"
	"log/slog"
)

func serveDeviceFrame(d *Device, gate *wireGate, frame []byte) error {
	// query_state is a host→device gateway-link control (NOT one of the four
	// device wire records): the upstream QueryState asks for a fresh state.
	if isQueryState(frame) {
		stateFrame, encErr := EncodeRecord(d.State())
		if encErr != nil {
			return fmt.Errorf("device: encode state: %w", encErr)
		}
		if err := gate.send(stateFrame); err != nil {
			return err
		}
		return nil
	}

	// The command stream mirrors the effector's Exchange: one command yields
	// exactly one ordered receipt/result pair. Execution truth (current_output)
	// is still read back via query_state; the result is device-reported
	// terminal status, not independent physical confirmation.
	outcome, handleErr := d.handleCommand(frame)
	if handleErr != nil {
		return sendMalformedExchange(gate)
	}
	return sendCommandOutcome(d, gate, frame, outcome)
}

func sendMalformedExchange(gate *wireGate) error {
	// A malformed command is a wire error, not a silent drop: report it
	// as a rejected receipt/result pair the upstream can act on.
	reject, encErr := EncodeRecord(malformedReceipt())
	if encErr != nil {
		return fmt.Errorf("device: encode malformed receipt: %w", encErr)
	}
	if err := gate.send(reject); err != nil {
		return err
	}
	malformed, encErr := EncodeRecord(malformedResult())
	if encErr != nil {
		return fmt.Errorf("device: encode malformed result: %w", encErr)
	}
	if err := gate.send(malformed); err != nil {
		return err
	}
	return nil
}

func sendCommandOutcome(d *Device, gate *wireGate, frame []byte, outcome Outcome) error {
	if outcome.Fault != "" {
		slog.Info("device fault injected", "fault", outcome.Fault, "accepted_command", outcome.AcceptedCommand)
	}
	accepted, _ := outcome.Receipt["accepted"].(bool)
	if !accepted {
		slog.Info("device command rejected", "command_id", outcome.Receipt["command_id"], "reject_code", outcome.Receipt["reject_code"])
	}
	receipt, err := EncodeRecord(outcome.Receipt)
	if err != nil {
		return fmt.Errorf("device: encode receipt: %w", err)
	}
	result, err := EncodeRecord(outcome.Result)
	if err != nil {
		return fmt.Errorf("device: encode result: %w", err)
	}
	if outcome.Duplicate {
		// Re-run the exact frame through admission/idempotency so the test
		// exercises the same path as a gateway duplicate. The replay is an
		// internal delivery, not a second protocol exchange: one inbound
		// command line must produce at most one wire receipt.
		if _, duplicateErr := d.handleCommand(frame); duplicateErr != nil {
			return fmt.Errorf("device: duplicate command: %w", duplicateErr)
		}
	}
	if outcome.AckLost {
		return nil // ack_lost: withhold receipt and result; effect stands, upstream reconciles via query_state
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
