package device

import (
	"io"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/uds"
)

// Test seams for the facade's black-box transport tests: the transport and
// the wire codec are reached through the module, not exported to other
// modules.
var (
	EncodeRecord = layer.EncodeRecord
	DecodeRecord = layer.DecodeRecord
)

const (
	QueryStateControl = uds.QueryStateControl
	FaultAckLost      = layer.FaultAckLost
	FaultDuplicate    = layer.FaultDuplicate
)

// ServeConn runs the session loop for a facade device over a connection.
func ServeConn(conn io.ReadWriter, d *Device) error {
	return uds.ServeConn(conn, d.device)
}

// ServeConnWithFaults is ServeConn with a deterministic wire-fault plan.
func ServeConnWithFaults(conn io.ReadWriter, d *Device, faults WireFaults) error {
	return uds.ServeConnWithFaults(conn, d.device, faults)
}

// AcceptedCommandCount reports how many commands the device accepted.
func (d *Device) AcceptedCommandCount() int {
	return d.device.AcceptedCommandCount()
}
