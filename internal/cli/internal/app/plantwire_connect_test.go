package app

import (
	"encoding/json"
	"os"
	"path/filepath"
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
