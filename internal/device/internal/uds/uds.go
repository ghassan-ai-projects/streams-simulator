// Package uds is the transport edge of the device module: the unix-domain
// socket listener, the newline-delimited session loop and the deterministic
// wire-fault gate that sits between the device rules and a gateway.
package uds

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

var ErrInjectedDisconnect = errors.New("device: injected disconnect")

func ServeConn(conn io.ReadWriter, d *domain.Device) error {
	return ServeConnWithFaults(conn, d, WireFaults{})
}

func ServeConnWithFaults(conn io.ReadWriter, d *domain.Device, faults WireFaults) error {
	gate := newWireGate(conn, faults)
	if err := sendInitialState(d, gate); err != nil {
		return err
	}
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 4096), domain.MaxFrameBytes)
	if err := scanDeviceFrames(scanner, d, gate); err != nil {
		return err
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("device: read connection: %w", err)
	}
	return gate.flush()
}

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
		"protocol_version": float64(domain.ProtocolVersion),
		"command_id":       "unknown",
		"boot_id":          "unknown",
		"accepted":         false,
		"reject_code":      "malformed",
	}
}

func malformedResult() map[string]any {
	return map[string]any{
		"message_type":     "result",
		"protocol_version": float64(domain.ProtocolVersion),
		"command_id":       "unknown",
		"boot_id":          "unknown",
		"status":           "rejected",
		"error_code":       "malformed",
	}
}

func Listen(path string, d *domain.Device) (*net.UnixListener, error) {
	listener, err := kernel.Listen(context.Background(), path)
	if err != nil {
		return nil, fmt.Errorf("device: listen on %s: %w", path, err)
	}
	go acceptDeviceConnections(listener, d)
	return listener, nil
}

func sendInitialState(d *domain.Device, gate *wireGate) error {
	frame, err := domain.EncodeRecord(d.State())
	if err != nil {
		return fmt.Errorf("device: encode initial state: %w", err)
	}
	if err := gate.send(frame); err != nil {
		return fmt.Errorf("device: write initial state: %w", err)
	}
	return nil
}

func scanDeviceFrames(scanner *bufio.Scanner, d *domain.Device, gate *wireGate) error {
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

func acceptDeviceConnections(listener *net.UnixListener, d *domain.Device) {
	for {
		conn, err := listener.AcceptUnix()
		if err != nil {
			return
		}
		serveAcceptedConnection(conn, d)
	}
}

func serveAcceptedConnection(conn *net.UnixConn, d *domain.Device) {
	if err := ServeConn(conn, d); err != nil && !errors.Is(err, ErrInjectedDisconnect) {
		slog.Error("device connection failed", "error", err)
	}
	if err := conn.Close(); err != nil {
		slog.Error("device connection close failed", "error", err)
	}
}
