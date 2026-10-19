package app

import (
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func policiesPlanText(hostSocket string, policies ...string) string {
	plan := `{"schema_version":"0.1","domain":"arduino-thermal","seed":3,"host_socket":"` + hostSocket + `"`
	if len(policies) > 0 {
		plan += `,"policies":["` + strings.Join(policies, `","`) + `"]`
	}
	return plan + "}"
}

func bootOf(received []map[string]any, device string) string {
	for _, birth := range ofType(received, "DBIRTH") {
		if birth["device"] == device {
			boot, _ := birth["boot_id"].(string)
			return boot
		}
	}
	return ""
}

func receiptOf(received []map[string]any, commandID string) (status, reason string) {
	for _, receipt := range ofType(received, "DRECEIPT") {
		if receipt["command_id"] == commandID {
			status, _ = receipt["status"].(string)
			reason, _ = receipt["reason"].(string)
		}
	}
	return status, reason
}

type policySession struct {
	out     string
	options ConnectOptions
	status  string
	reason  string
	outcome string
}

func runPolicySession(t *testing.T, policies []string, sent string) policySession {
	t.Helper()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-policy")
	host := serveWireHost(t, hostPath, true)
	options := connectOptions(t, policiesPlanText(hostPath, policies...), out)
	cancel, done := startConnect(t, options, false)
	host.waitFor(func(r []map[string]any) bool {
		return hasRecord(r, "DBIRTH", "device", "fan-01") && hasRecord(r, "NSYNC", "", nil)
	})
	received := host.snapshot()
	command := commandRecord("cmd-1", "fan-01", bootOf(received, "fan-01"), lastWorldTime(received))
	command.Fields["policy_digest"] = sent
	host.write(command)
	host.release()
	host.waitFor(func(r []map[string]any) bool {
		status, _ := receiptOf(r, "cmd-1")
		return status == "refused" || outcomeOf(r, "cmd-1") != ""
	})
	cancel()
	if ended := <-done; ended.err != nil || !ended.result.Stopped {
		t.Fatalf("Connect = %+v, want a stopped session", ended)
	}
	status, reason := receiptOf(host.snapshot(), "cmd-1")
	return policySession{out: out, options: options, status: status, reason: reason, outcome: outcomeOf(host.snapshot(), "cmd-1")}
}

func effectorCalls(t *testing.T, out string) int {
	t.Helper()
	artifact, err := run.LoadArtifact(filepath.Join(out, model.FileRunArtifact))
	if err != nil {
		t.Fatal(err)
	}
	return len(slices.DeleteFunc(slices.Clone(artifact.CommandLog), func(c model.Command) bool { return c.Op != model.OpEffectorInvoke }))
}

func lastLiveCommand(t *testing.T, out string) map[string]any {
	t.Helper()
	live := journalOfType(t, out, "live")
	if len(live) == 0 {
		t.Fatal("no live projection recorded")
	}
	commands, _ := live[len(live)-1]["commands"].([]any)
	for _, entry := range commands {
		if command, _ := entry.(map[string]any); command["command_id"] == "cmd-1" {
			return command
		}
	}
	t.Fatalf("the last live record holds no cmd-1: %v", live[len(live)-1]["commands"])
	return nil
}

func TestAConnectedSessionExecutesACommandWhoseDigestThePlanListsAndRecordsTheListSorted(t *testing.T) {
	t.Parallel()
	list := []string{digestOf("c"), mixedPolicy}
	session := runPolicySession(t, list, mixedPolicy)
	if session.status != "accepted" || session.outcome != "succeeded" || effectorCalls(t, session.out) != 1 {
		t.Fatalf("receipt %q result %q, effector calls %d, want an executed command", session.status, session.outcome, effectorCalls(t, session.out))
	}
	plan, _ := journalOfType(t, session.out, "session")[0]["plan"].(map[string]any)
	if !reflect.DeepEqual(plan["policies"], []any{mixedPolicy, digestOf("c")}) {
		t.Fatalf("session plan policies = %v, want the list sorted", plan["policies"])
	}
	if command := lastLiveCommand(t, session.out); command["policy_digest"] != nil || command["receipt"] != "accepted" {
		t.Fatalf("live command = %v, want an accepted command with no policy digest", command)
	}
	replaysWhole(t, session.out, session.options.DomainsDir)
}

func TestAConnectedSessionRefusesADigestThePlanDoesNotListAndShowsItOnTheLiveCommand(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		listed []string
	}{
		{"a list that names other digests", []string{digestOf("c")}},
		{"a plan with no list", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			session := runPolicySession(t, tc.listed, mixedPolicy)
			if session.status != "refused" || session.reason != "policy_not_allowed" || session.outcome != "" || effectorCalls(t, session.out) != 0 {
				t.Fatalf("receipt %q/%q result %q, want a policy refusal and no effector call", session.status, session.reason, session.outcome)
			}
			if command := lastLiveCommand(t, session.out); command["policy_digest"] != mixedPolicy || command["reason"] != "policy_not_allowed" {
				t.Fatalf("live command = %v, want the refused digest %s", command, mixedPolicy)
			}
			plan, _ := journalOfType(t, session.out, "session")[0]["plan"].(map[string]any)
			want := []any{}
			for _, digest := range tc.listed {
				want = append(want, digest)
			}
			if !reflect.DeepEqual(plan["policies"], want) {
				t.Fatalf("session plan policies = %#v, want %#v", plan["policies"], want)
			}
			replaysWhole(t, session.out, session.options.DomainsDir)
		})
	}
}

func TestACheckEchoesTheSortedListInEffectAndNeverNull(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		plan []string
		want []string
	}{
		{"two digests in any order", []string{digestOf("c"), digestOf("a")}, []string{digestOf("a"), digestOf("c")}},
		{"no list", nil, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			check, err := CheckConnect(connectOptions(t, policiesPlanText("/tmp/h.sock", tc.plan...), ""))
			if err != nil || !reflect.DeepEqual(check.Policies, tc.want) {
				t.Fatalf("CheckConnect = %+v, %v, want the policies %#v", check, err, tc.want)
			}
		})
	}
}

func TestAPlanWithABrokenPolicyListIsAViolationAtTheEntryAndNothingStarts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		list    []string
		pointer string
		keyword string
	}{
		{"a repeat", []string{digestOf("a"), digestOf("b"), digestOf("a")}, "/policies/2", "policy_listed_twice"},
		{"the reserved stop digest", []string{digestOf("a"), digestOf("0")}, "/policies/1", "policy_reserved"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := CheckConnect(connectOptions(t, policiesPlanText("/tmp/h.sock", tc.list...), ""))
			invalid, ok := model.AsValidationError(err)
			if !ok || invalid.Contract != model.ContractSessionPlan || invalid.Issues[0].Pointer != tc.pointer || invalid.Issues[0].Keyword != tc.keyword {
				t.Fatalf("CheckConnect = %v, want a session-plan violation %s at %s", err, tc.keyword, tc.pointer)
			}
		})
	}
}
