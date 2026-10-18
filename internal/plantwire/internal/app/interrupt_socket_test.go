package app

import (
	"bufio"
	"encoding/json"
	"errors"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

type controlClient struct {
	t     *testing.T
	conn  net.Conn
	lines *bufio.Scanner
}

func dialControl(t *testing.T, path string) *controlClient {
	t.Helper()
	conn, err := (&net.Dialer{}).DialContext(t.Context(), "unix", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1<<16), 1<<20)
	return &controlClient{t: t, conn: conn, lines: scanner}
}

func (c *controlClient) write(line string) {
	c.t.Helper()
	if _, err := c.conn.Write([]byte(line + "\n")); err != nil {
		c.t.Fatal(err)
	}
}

func (c *controlClient) read() map[string]any {
	c.t.Helper()
	if !c.lines.Scan() {
		c.t.Fatalf("the control socket closed without an answer: %v", c.lines.Err())
	}
	var reply map[string]any
	if err := json.Unmarshal(c.lines.Bytes(), &reply); err != nil {
		c.t.Fatal(err)
	}
	return reply
}

func (c *controlClient) ask(line string) map[string]any {
	c.t.Helper()
	c.write(line)
	return c.read()
}

func holdCount(reply map[string]any) int {
	state, _ := reply["state"].(map[string]any)
	holds, _ := state["holds"].([]any)
	return len(holds)
}

func TestAnOperatorsRequestsReachARunningNodeOverItsOwnerOnlySocketAndTheStoppedSessionKeepsThemInTheRun(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), t.TempDir()
	control := filepath.Join(filepath.Dir(hostPath), "c.sock")
	host := serveWireHost(t, hostPath, true)
	options := connectOptions(t, planText(hostPath, ""), out)
	options.ControlSocket = control
	cancel, done := startConnect(t, options, false)
	host.waitFor(func(r []map[string]any) bool { return hasRecord(r, "NSYNC", "", nil) })
	client := dialControl(t, control)
	if state := client.ask(`{"op":"state"}`); state["result"] != "applied" || holdCount(state) != 0 {
		t.Fatalf("state = %v, want it answered at once with nothing held", state)
	}
	if info, err := os.Stat(control); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("control socket %v (%v), want an owner-only socket", info, err)
	}
	client.write(`{"op":"estop","scope":{"kind":"device","id":"fan-01"}}`)
	host.release()
	if reply := client.read(); reply["result"] != "applied" || reply["stopped"].([]any)[0] != "fan-01" || holdCount(reply) != 1 {
		t.Fatalf("estop reply = %v, want fan-01 stopped and a hold latched", reply)
	}
	if reply := client.ask(`not json`); reply["result"] != "refused" || reply["reason"] != "malformed" {
		t.Fatalf("a request that is no JSON = %v, want it refused as malformed", reply)
	}
	if reply := client.ask(`{"op":"release","scope":{"kind":"device","id":"fan-01"}}`); reply["result"] != "applied" || holdCount(reply) != 0 {
		t.Fatalf("release reply = %v, want the latch cleared", reply)
	}
	cancel()
	ended := <-done
	if ended.err != nil || !ended.result.Stopped {
		t.Fatalf("session = %+v, want it stopped by the interrupt", ended)
	}
	replayed, err := VerifyInterruptLog(out)
	if err != nil || replayed.Records < 6 {
		t.Fatalf("VerifyInterruptLog = %+v, %v, want the control, the stop and the release in a log that replays", replayed, err)
	}
	replaysWhole(t, out, options.DomainsDir)
	if _, err := os.Stat(control); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Stat of the control socket = %v, want it removed with the session", err)
	}
}

func tickAt(t *testing.T, ticks chan<- time.Time, seconds int64) {
	t.Helper()
	select {
	case ticks <- time.Unix(1767225600+seconds, 0):
	case <-t.Context().Done():
		t.Fatal("the relay took no tick")
	}
}

func TestAMixedRelayServesItsControlsOnItsTicksAndKeepsThemInTheRun(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), t.TempDir()
	host := serveWireHost(t, hostPath, true)
	options := mixedOptions(t, out)
	options.Control = NewControlPort()
	ticks := make(chan time.Time)
	relay := startMixedRelayTicked(t, hostPath, options, ticks)
	gateway := dialGateway(t, relay.gateway, mixedGateway(t, "gw-1"))
	gateway.handshake()
	host.waitFor(func(r []map[string]any) bool {
		return hasRecord(r, "DBIRTH", "device", "fan-01") && hasRecord(r, "DBIRTH", "device", "fan-02") && hasRecord(r, "NSYNC", "", nil)
	})
	hold := controlThroughTicks(t, options.Control, ticks, `{"op":"hold","scope":{"kind":"device","id":"fan-02"}}`)
	now := lastWorldTime(host.snapshot())
	host.write(commandRecord("hw-cmd", "fan-02", "boot-1", now))
	host.waitFor(func(r []map[string]any) bool { return hasRecord(r, "DRECEIPT", "command_id", "hw-cmd") })
	stop := controlThroughTicks(t, options.Control, ticks, `{"op":"estop","scope":{"kind":"device","id":"fan-01"}}`)
	host.release()
	host.waitFor(func(r []map[string]any) bool { return hasRecord(r, "DDEATH", "device", "zone-01") })
	ended := relay.stop()
	assertRelayControls(t, hold, stop, host.snapshot(), ended)
	assertRelayRun(t, out, options.Simulated.DomainsDir)
}

func controlThroughTicks(t *testing.T, port *ControlPort, ticks chan<- time.Time, line string) map[string]any {
	t.Helper()
	request, err := domain.DecodeControl([]byte(line))
	if err != nil {
		t.Fatal(err)
	}
	call, queued := port.enqueue(t.Context(), request)
	if !queued {
		t.Fatal("the control port took no request")
	}
	tickAt(t, ticks, 1)
	tickAt(t, ticks, 2)
	var reply map[string]any
	if err := json.Unmarshal(<-call.Reply, &reply); err != nil {
		t.Fatal(err)
	}
	return reply
}

func assertRelayControls(t *testing.T, hold, stop map[string]any, received []map[string]any, ended relayRun) {
	t.Helper()
	if hold["result"] != "applied" || stop["result"] != "applied" || stop["stopped"].([]any)[0] != "fan-01" {
		t.Fatalf("hold %v stop %v, want both applied and fan-01 stopped", hold, stop)
	}
	var refused map[string]any
	for _, receipt := range ofType(received, "DRECEIPT") {
		if receipt["command_id"] == "hw-cmd" {
			refused = receipt
		}
		if receipt["command_id"] == "node-stop-1" {
			t.Fatalf("the host received %v, an answer to a command it never sent", receipt)
		}
	}
	if refused["status"] != "refused" || refused["reason"] != "local_control" || refused["boot_id"] != "none" {
		t.Fatalf("receipt = %v, want the held command refused as local_control", refused)
	}
	if ended.err != nil || ended.stats.Down != 1 {
		t.Fatalf("relay %+v, want only the host's STATE to have gone to the gateway: the held command never did", ended)
	}
}

func assertRelayRun(t *testing.T, out, domains string) {
	t.Helper()
	art, err := run.LoadArtifact(filepath.Join(out, model.FileRunArtifact))
	if err != nil || art.InterruptLogDigest == "" || countInvokes(art) != 1 {
		t.Fatalf("artifact %+v (%v), want the interrupt log covered and one effector call, the safe stop", art, err)
	}
	if replayed, err := VerifyInterruptLog(out); err != nil || replayed.Digest != art.InterruptLogDigest {
		t.Fatalf("VerifyInterruptLog = %+v, %v, want the log to replay under the recorded digest", replayed, err)
	}
	raw, err := os.ReadFile(filepath.Join(out, model.FileInterruptLog))
	if err != nil || !strings.Contains(string(raw), `"event":"stop"`) || !strings.Contains(string(raw), `"event":"effect"`) {
		t.Fatalf("interrupt log %q (%v), want the stop and its effect", raw, err)
	}
	replaysWhole(t, out, domains)
}
