package app

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func thermalDomains(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(testsupport.ThermalBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	testsupport.WriteFileIn(t, dir, "arduino-thermal.domain.json", raw)
	return dir
}

func benchPlan(instance string) string {
	return `{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/tmp/h.sock","gateway_socket":"/tmp/g.sock",` +
		`"real":[{"instance":"` + instance + `","board":"uno-07","device":"zone-01"}]}`
}

func checkArgs(domains string) []string {
	return []string{"plantwire", "connect", "--stdin", "--check", "--domains-dir", domains, "--adapters-dir", testsupport.AdaptersDir(),
		"--kinds-dir", testsupport.ThermalBoardKindsDir()}
}

func TestPlantWireConnectUsageIsChecked(t *testing.T) {
	t.Parallel()
	requireUsageRefusals(t, [][]string{{"plantwire", "connect"}, {"plantwire", "connect", "--plan", "/tmp/p.json", "--stdin", "--check"},
		{"plantwire", "connect", "--stdin"}, {"plantwire", "connect", "--stdin", "--check", "extra"},
		{"plantwire", "connect", "--stdin", "--check", "--duration", "60"}})
}

func TestAConnectCheckNamesEachInstanceRealOrSimulatedInTheOutputContract(t *testing.T) {
	t.Parallel()
	got := invokeWithInput(t, benchPlan("zone-01"), checkArgs(thermalDomains(t))...)
	if got.code != 0 {
		t.Fatalf("exit %d: %s %s", got.code, got.stdout, got.stderr)
	}
	var doc any
	if err := model.DecodeBytes([]byte(got.stdout), &doc); err != nil {
		t.Fatal(err)
	}
	if errs := cliOutputSchema(t, "plantwire_connect_check").Validate(doc); len(errs) > 0 {
		t.Fatalf("the check does not match cli-output plantwire_connect_check:\n%s\n%s", jsonschema.FormatErrors(errs), got.stdout)
	}
	var check struct {
		Environment string `json:"environment"`
		Instances   []struct {
			Instance string `json:"instance"`
			Origin   string `json:"origin"`
			Board    string `json:"board"`
		} `json:"instances"`
	}
	if err := json.Unmarshal([]byte(got.stdout), &check); err != nil {
		t.Fatal(err)
	}
	origins := map[string]string{}
	for _, instance := range check.Instances {
		origins[instance.Instance] = instance.Origin + " " + instance.Board
	}
	if check.Environment != "mixed" || origins["zone-01"] != "real uno-07" || origins["fan-01"] != "simulated " {
		t.Fatalf("check = %+v, want zone-01 real on uno-07 and fan-01 simulated in a mixed session", check)
	}
}

func TestAConnectPlanThatDoesNotHoldIsASessionInvalidFailureWithItsIssue(t *testing.T) {
	t.Parallel()
	got := invokeWithInput(t, benchPlan("ghost-01"), checkArgs(thermalDomains(t))...)
	body := got.failure(t)
	issues, _ := body["issues"].([]any)
	first, _ := issues[0].(map[string]any)
	if got.code != 1 || body["code"] != "plantwire_failed" || body["source"] != "stdin" || len(issues) != 1 || first["pointer"] != "/real/0/instance" ||
		first["keyword"] != "unknown_instance" {
		t.Fatalf("exit %d with %v, want plantwire_failed from stdin with one unknown_instance issue at /real/0/instance", got.code, body)
	}
}

func TestAConnectPlanFileThatIsAbsentIsReportedAsSuch(t *testing.T) {
	t.Parallel()
	got := invoke(t, "plantwire", "connect", "--plan", filepath.Join(t.TempDir(), "absent.json"), "--check")
	if body := got.failure(t); got.code != 1 || body["code"] != "file_not_found" {
		t.Fatalf("exit %d with %v, want file_not_found", got.code, body)
	}
}

func TestTheControlSocketFlagTakesAPathAndAConnectCheckStartsNothingForIt(t *testing.T) {
	t.Parallel()
	requireUsageRefusals(t, [][]string{{"plantwire", "connect", "--stdin", "--check", "--control-socket"}})
	args := append(checkArgs(thermalDomains(t)), "--control-socket", filepath.Join(t.TempDir(), "c.sock"))
	got := invokeWithInput(t, benchPlan("zone-01"), args...)
	if got.code != 0 {
		t.Fatalf("exit %d: %s %s, want a check that ignores the control socket: it starts no session", got.code, got.stdout, got.stderr)
	}
	if _, err := os.Stat(args[len(args)-1]); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Stat of the control socket = %v, want a check to create none", err)
	}
}

func policyPlan(policies ...string) string {
	return `{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/tmp/h.sock","policies":["` + strings.Join(policies, `","`) + `"]}`
}

func TestAConnectCheckEchoesThePlansPolicyListSortedAndEmptyWhenThePlanHasNone(t *testing.T) {
	t.Parallel()
	low, high := "sha256:"+strings.Repeat("a", 64), "sha256:"+strings.Repeat("c", 64)
	cases := map[string]struct {
		plan string
		want []any
	}{
		"a list in any order": {policyPlan(high, low), []any{low, high}},
		"no list":             {`{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/tmp/h.sock"}`, []any{}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			doc := invokeWithInput(t, tc.plan, checkArgs(thermalDomains(t))...).mustSucceed(t).document(t, "plantwire_connect_check")
			if !reflect.DeepEqual(doc["policies"], tc.want) {
				t.Fatalf("policies = %#v, want %#v", doc["policies"], tc.want)
			}
		})
	}
}

func TestAConnectPlanWithABrokenPolicyListIsRefusedAtTheEntryOrTheList(t *testing.T) {
	t.Parallel()
	digest := func(char string) string { return "sha256:" + strings.Repeat(char, 64) }
	nine := make([]string, 9)
	for i := range nine {
		nine[i] = digest(string("abcdef012"[i]))
	}
	cases := []struct {
		name, plan, pointer, keyword string
	}{
		{"a repeat", policyPlan(digest("a"), digest("a")), "/policies/1", "policy_listed_twice"},
		{"the reserved stop digest", policyPlan(digest("0")), "/policies/0", "policy_reserved"},
		{"a digest in upper case", policyPlan(strings.ToUpper(digest("a"))), "/policies/0", "pattern"},
		{"nine digests", policyPlan(nine...), "/policies", "maxItems"},
		{"an empty digest", policyPlan(), "/policies/0", "pattern"},
		{"an empty list", `{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/tmp/h.sock","policies":[]}`, "/policies", "minItems"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invokeWithInput(t, tc.plan, checkArgs(thermalDomains(t))...)
			body := got.failure(t)
			issues, _ := body["issues"].([]any)
			if got.code != 1 || len(issues) == 0 {
				t.Fatalf("exit %d with %v, want a refusal with its issue", got.code, body)
			}
			if first, _ := issues[0].(map[string]any); first["pointer"] != tc.pointer || first["keyword"] != tc.keyword {
				t.Fatalf("first issue = %v, want %s at %s", first, tc.keyword, tc.pointer)
			}
		})
	}
}

func TestAConnectIntoAFolderThatHoldsARunFailsAndKeepsTheEarlierRun(t *testing.T) {
	t.Parallel()
	out := t.TempDir()
	testsupport.WriteFileIn(t, out, "run.json", []byte("{}\n"))
	plan := `{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/tmp/h.sock"}`
	got := invokeWithInput(t, plan, "plantwire", "connect", "--stdin", "--out", out, "--domains-dir", thermalDomains(t),
		"--adapters-dir", testsupport.AdaptersDir(), "--kinds-dir", testsupport.ThermalBoardKindsDir())
	body := got.failure(t)
	if message, _ := body["message"].(string); got.code != 1 || body["code"] != "plantwire_failed" || !strings.Contains(message, "already holds a run") {
		t.Fatalf("exit %d with %v, want plantwire_failed saying the folder already holds a run", got.code, body)
	}
	if raw, err := os.ReadFile(filepath.Join(out, "run.json")); err != nil || string(raw) != "{}\n" {
		t.Fatalf("run.json after the refusal = %q, %v, want the earlier run kept", raw, err)
	}
}
