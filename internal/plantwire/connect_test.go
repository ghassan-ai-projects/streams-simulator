package plantwire

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func facadeConnectOptions(t *testing.T, plan string) ConnectOptions {
	t.Helper()
	raw, err := os.ReadFile(testsupport.ThermalBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	domains := t.TempDir()
	testsupport.WriteFileIn(t, domains, "arduino-thermal.domain.json", raw)
	return ConnectOptions{Plan: strings.NewReader(plan), Source: "plan.json", DomainsDir: domains, AdaptersDir: testsupport.AdaptersDir(),
		KindsDirs: []string{testsupport.ThermalBoardKindsDir()}, SliceNS: 1e9}
}

func TestTheFacadeChecksAConnectPlanAndRefusesASessionWithoutARunFolder(t *testing.T) {
	t.Parallel()
	plan := `{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/tmp/h.sock","gateway_socket":"/tmp/g.sock",` +
		`"real":[{"instance":"zone-01","board":"uno-07","device":"zone-01"}]}`
	check, err := CheckConnect(facadeConnectOptions(t, plan))
	if err != nil || check.Environment != "mixed" || len(check.Instances) != 3 {
		t.Fatalf("CheckConnect = %+v, %v, want the three instances of a mixed session", check, err)
	}
	if _, err := Connect(t.Context(), facadeConnectOptions(t, plan)); err == nil || !strings.Contains(err.Error(), "needs a run folder") {
		t.Fatalf("Connect without a run folder = %v, want the missing folder refused before anything starts", err)
	}
	unknown := strings.Replace(plan, `"instance":"zone-01"`, `"instance":"ghost"`, 1)
	if _, err := CheckConnect(facadeConnectOptions(t, unknown)); !errors.Is(err, ErrPlanRefused) {
		t.Fatalf("CheckConnect of an unknown instance = %v, want ErrPlanRefused", err)
	}
}

func TestTheFacadeEchoesThePlansPolicyListSorted(t *testing.T) {
	t.Parallel()
	low, high := "sha256:"+strings.Repeat("a", 64), "sha256:"+strings.Repeat("c", 64)
	plan := `{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/tmp/h.sock","policies":["` + high + `","` + low + `"]}`
	check, err := CheckConnect(facadeConnectOptions(t, plan))
	if err != nil || len(check.Policies) != 2 || check.Policies[0] != low || check.Policies[1] != high {
		t.Fatalf("CheckConnect = %+v, %v, want the plan's policies sorted", check, err)
	}
}
