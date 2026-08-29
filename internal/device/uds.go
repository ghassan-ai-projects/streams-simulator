package device

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

// ServeConn runs one device over a byte-stream connection using newline-delimited
// framing. On connect it emits the current device.state; then for each inbound
// command line it applies the command and writes the receipt followed by the
// result. Under an ack_lost fault it writes NOTHING for that command — the effect
// is still applied, so the upstream must reconcile via a later state query, which
// is exactly the unknown-outcome path.
//
// Raw serial framing (COBS, checksums, reconnect, device identity) is a gateway
// concern; this NDJSON line framing is the debug/emulator transport.
func ServeConn(conn io.ReadWriter, d *Device) error {
	return ServeConnWithFaults(conn, d, WireFaults{})
}

// ServeConnWithFaults is ServeConn with a deterministic transport-fault plan
// applied to the outbound frames (state/receipt/result), indexed by emission
// order across the connection. See WireFaults.
func ServeConnWithFaults(conn io.ReadWriter, d *Device, faults WireFaults) error {
	gate := newWireGate(conn, faults)
	stateFrame, err := EncodeRecord(d.State())
	if err != nil {
		return fmt.Errorf("device: encode initial state: %w", err)
	}
	if err := gate.send(stateFrame); err != nil {
		return fmt.Errorf("device: write initial state: %w", err)
	}

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 4096), maxFrameBytes)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		frame := append(append([]byte{}, line...), '\n')
		receipt, result, ackLost, handleErr := d.HandleCommand(frame)
		if handleErr != nil {
			// A malformed command is a wire error, not a silent drop: report it
			// as a rejected receipt the upstream can act on.
			reject, encErr := EncodeRecord(malformedReceipt())
			if encErr != nil {
				return fmt.Errorf("device: encode malformed receipt: %w", encErr)
			}
			if err := gate.send(reject); err != nil {
				return err
			}
			continue
		}
		if ackLost {
			continue // ack_lost: withhold receipt and result; effect stands
		}
		if err := gate.send(receipt); err != nil {
			return err
		}
		if err := gate.send(result); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("device: read connection: %w", err)
	}
	return gate.flush()
}

func malformedReceipt() map[string]any {
	return map[string]any{
		"message_type":     "receipt",
		"protocol_version": float64(ProtocolVersion),
		"command_id":       "unknown",
		"boot_id":          "unknown",
		"accepted":         false,
		"reject_code":      "malformed",
	}
}

// Listen serves a single device on a Unix domain socket at path until the
// listener is closed. Each accepted connection is handled sequentially — one
// device speaks to one gateway link at a time. It removes a stale socket file
// at path before binding.
func Listen(path string, d *Device) (*net.UnixListener, error) {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("device: clear stale socket %s: %w", path, err)
	}
	addr, err := net.ResolveUnixAddr("unix", path)
	if err != nil {
		return nil, fmt.Errorf("device: resolve socket %s: %w", path, err)
	}
	listener, err := net.ListenUnix("unix", addr)
	if err != nil {
		return nil, fmt.Errorf("device: listen on %s: %w", path, err)
	}
	go func() {
		for {
			conn, err := listener.AcceptUnix()
			if err != nil {
				return // listener closed
			}
			_ = ServeConn(conn, d)
			_ = conn.Close()
		}
	}()
	return listener, nil
}
