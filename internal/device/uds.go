package device

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
)

// ServeConn runs one device over a byte-stream connection using newline-delimited
// framing. On connect it emits the current device.state; then for each inbound
// command line it applies the command and writes one receipt. The result is
// available from HandleCommand for direct callers; the gateway link uses state
// queries for observed execution state. Under an ack_lost fault it writes
// NOTHING for that command — the effect is still applied, so the upstream must
// reconcile via a later state query.
//
// Raw serial framing (COBS, checksums, reconnect, device identity) is a gateway
// concern; this NDJSON line framing is the debug/emulator transport.
func ServeConn(conn io.ReadWriter, d *Device) error {
	return ServeConnWithFaults(conn, d, WireFaults{})
}

// ServeConnWithFaults is ServeConn with a deterministic transport-fault plan
// applied to the outbound frames (state/receipt), indexed by emission
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
			continue
		}

		// The command stream mirrors the effector's Exchange: one command yields
		// exactly one receipt. Execution truth (current_output) is read back via
		// query_state, so the result record is not pushed unsolicited here — that
		// would desync the receipt the upstream reads next.
		receipt, _, ackLost, handleErr := d.HandleCommand(frame)
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
			continue // ack_lost: withhold the receipt; effect stands, upstream reconciles via query_state
		}
		if err := gate.send(receipt); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("device: read connection: %w", err)
	}
	return gate.flush()
}

// QueryStateControl is the host→device gateway-link control line that asks the
// device for a fresh state record. It is transport control, not one of the four
// device wire records, so it carries only message_type.
const QueryStateControl = `{"message_type":"query_state"}`

func isQueryState(frame []byte) bool {
	var probe struct {
		MessageType string `json:"message_type"`
	}
	if err := json.Unmarshal(frame, &probe); err != nil {
		return false
	}
	return probe.MessageType == "query_state"
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
// at path before binding, but never removes a regular file or directory.
func Listen(path string, d *Device) (*net.UnixListener, error) {
	if err := removeStaleSocket(path); err != nil {
		return nil, err
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
			if err := ServeConn(conn, d); err != nil {
				slog.Error("device connection failed", "error", err)
			}
			if err := conn.Close(); err != nil {
				slog.Error("device connection close failed", "error", err)
			}
		}
	}()
	return listener, nil
}

func removeStaleSocket(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("device: inspect socket path %s: %w", path, err)
	}
	if info.Mode()&os.ModeSocket == 0 {
		return fmt.Errorf("device: refusing to remove non-socket path %s", path)
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("device: clear stale socket %s: %w", path, err)
	}
	return nil
}
