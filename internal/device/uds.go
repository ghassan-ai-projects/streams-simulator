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

// ErrInjectedDisconnect identifies a deterministic disconnect fault. The
// listener closes the current connection and remains available for the next
// gateway connection.
var ErrInjectedDisconnect = errors.New("device: injected disconnect")

// ServeConn runs one device over a byte-stream connection using newline-delimited
// framing. On connect it emits the current device.state; then for each inbound
// command line it applies the command and writes the ordered receipt and
// terminal result. Under an ack_lost fault it writes NOTHING for that command
// — the effect is still applied, so the upstream must reconcile via a later
// state query before retrying.
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
	if err := sendInitialState(d, gate); err != nil {
		return err
	}
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 4096), maxFrameBytes)
	if err := scanDeviceFrames(scanner, d, gate); err != nil {
		return err
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

func malformedResult() map[string]any {
	return map[string]any{
		"message_type":     "result",
		"protocol_version": float64(ProtocolVersion),
		"command_id":       "unknown",
		"boot_id":          "unknown",
		"status":           "rejected",
		"error_code":       "malformed",
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
	go acceptDeviceConnections(listener, d)
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
	return clearSocketPath(path)
}

func sendInitialState(d *Device, gate *wireGate) error {
	frame, err := EncodeRecord(d.State())
	if err != nil {
		return fmt.Errorf("device: encode initial state: %w", err)
	}
	if err := gate.send(frame); err != nil {
		return fmt.Errorf("device: write initial state: %w", err)
	}
	return nil
}

func scanDeviceFrames(scanner *bufio.Scanner, d *Device, gate *wireGate) error {
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		frame := append(append([]byte{}, line...), '\n')
		if err := serveDeviceFrame(d, gate, frame); err != nil {
			return err
		}
	}
	return nil
}

func acceptDeviceConnections(listener *net.UnixListener, d *Device) {
	for {
		conn, err := listener.AcceptUnix()
		if err != nil {
			return
		}
		serveAcceptedConnection(conn, d)
	}
}

func serveAcceptedConnection(conn *net.UnixConn, d *Device) {
	if err := ServeConn(conn, d); err != nil && !errors.Is(err, ErrInjectedDisconnect) {
		slog.Error("device connection failed", "error", err)
	}
	if err := conn.Close(); err != nil {
		slog.Error("device connection close failed", "error", err)
	}
}

func clearSocketPath(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("device: clear stale socket %s: %w", path, err)
	}
	return nil
}
