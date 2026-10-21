package app

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func latestSync(received []map[string]any) int64 {
	times := syncTimes(received)
	return times[len(times)-1]
}

func TestAMappedConnectedSessionTranslatesACommandRecordsItRawAndReplays(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-1")
	host := serveWireHost(t, hostPath, false)
	options := connectOptions(t, strings.TrimSuffix(planText(hostPath, ""), "}")+`,"policies":["`+policy+`"]}`, out)
	cancel, done := startConnect(t, options, false)
	host.waitFor(func(r []map[string]any) bool {
		return hasRecord(r, "DBIRTH", "device", "fan-01") && len(syncTimes(r)) > 0
	})
	received := host.snapshot()
	fields := set(commandFields("c-cool", "fan-01", "set_cooling", map[string]any{"level_percent": 45}, latestSync(received)+3000e9),
		"expected_boot_id", bootOf(received, "fan-01"))
	fields["authority_epoch"] = 7
	host.write(domain.Record{Type: "DCMD", Lane: domain.LaneControl, Fields: withKey(fields)})
	host.waitFor(func(r []map[string]any) bool { return hasRecord(r, "DRECEIPT", "command_id", "c-cool") })
	cancel()
	if ended := <-done; ended.err != nil {
		t.Fatal(ended.err)
	}
	if status, reason := receiptOf(host.snapshot(), "c-cool"); status != "accepted" {
		t.Fatalf("set_cooling 45: %s %s, want accepted", status, reason)
	}
	log := commandLog(t, out)
	invoked := slices.IndexFunc(log, func(c model.Command) bool { return c.Op == model.OpEffectorInvoke })
	if invoked < 0 || log[invoked].Args["effector"] != "set_pwm_lease" || fmt.Sprint(log[invoked].Args["args"]) != "map[duty_permille:450 lease_ms:5000]" {
		t.Fatalf("command log %v, want the raw set_pwm_lease 450/5000 the twin ran", log)
	}
	replaysWhole(t, out, options.DomainsDir)
}
