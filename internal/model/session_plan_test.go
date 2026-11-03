package model

import (
	"errors"
	"strings"
	"testing"
)

func TestASessionPlanDecodesWithItsBindingsAndASeedOfOneWhenNoneIsGiven(t *testing.T) {
	t.Parallel()
	plan, err := DecodeSessionPlan([]byte(`{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/tmp/h.sock",
		"gateway_socket":"/tmp/g.sock","real":[{"instance":"zone-01","board":"uno-07","device":"zone-01"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if plan.WorldSeed() != 1 || len(plan.Real) != 1 || plan.Real[0].Board != "uno-07" || plan.GatewaySocket != "/tmp/g.sock" {
		t.Fatalf("plan = %+v, want the one binding, the gateway socket and seed 1", plan)
	}
	seeded, err := DecodeSessionPlan([]byte(`{"schema_version":"0.1","domain":"arduino-thermal","seed":7,"host_socket":"/tmp/h.sock"}`))
	if err != nil || seeded.WorldSeed() != 7 || len(seeded.Real) != 0 {
		t.Fatalf("seeded plan = %+v, %v, want seed 7 and no real board", seeded, err)
	}
}

func TestASessionPlanThatBreaksItsContractIsRefusedAtTheMember(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, plan, pointer string }{
		{name: "a relative socket", plan: `{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"h.sock"}`, pointer: "/host_socket"},
		{name: "a socket with a parent component", plan: `{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/tmp/../h.sock"}`, pointer: "/host_socket"},
		{name: "a gateway socket with a control character", pointer: "/gateway_socket",
			plan: `{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/h","gateway_socket":"/tmp/g\u0007.sock"}`},
		{name: "a board that is no wire id", pointer: "/real/0/board",
			plan: `{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/h","real":[{"instance":"zone-01","board":"uno 07","device":"zone-01"}]}`},
		{name: "an unknown member", plan: `{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/h","duration":60}`, pointer: "/duration"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := DecodeSessionPlan([]byte(tc.plan))
			invalid, ok := AsValidationError(err)
			if !ok || invalid.Contract != ContractSessionPlan || len(invalid.Issues) == 0 || invalid.Issues[0].Pointer != tc.pointer {
				t.Fatalf("DecodeSessionPlan = %v, want a session-plan violation at %s", err, tc.pointer)
			}
		})
	}
}

func TestASessionPlanTheSchemaAdmitsButNoPlanCanHoldIsAViolationOfItsContract(t *testing.T) {
	t.Parallel()
	_, err := DecodeSessionPlan([]byte(`{"schema_version":"0.1","domain":"arduino-thermal","seed":18446744073709551616,"host_socket":"/h"}`))
	invalid, ok := AsValidationError(err)
	if !ok || invalid.Contract != ContractSessionPlan || invalid.Issues[0].Keyword != "json" || errors.Unwrap(err) == nil {
		t.Fatalf("a seed beyond uint64 = %v, want a session-plan violation keyed json", err)
	}
}

func policiesPlan(policies string) []byte {
	return []byte(`{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/h","policies":` + policies + `}`)
}

func policyDigests(count int) string {
	digests := make([]string, count)
	for i := range digests {
		digests[i] = `"sha256:` + strings.Repeat(string("abcdef0123456789"[i]), 64) + `"`
	}
	return "[" + strings.Join(digests, ",") + "]"
}

func TestASessionPlanDecodesOneToEightPolicyDigestsInTheOperatorsOrder(t *testing.T) {
	t.Parallel()
	for _, count := range []int{1, 8} {
		plan, err := DecodeSessionPlan(policiesPlan(policyDigests(count)))
		if err != nil || len(plan.Policies) != count || plan.Policies[0] != "sha256:"+strings.Repeat("a", 64) {
			t.Fatalf("%d policies = %+v, %v, want them all in the order written", count, plan, err)
		}
	}
	absent, err := DecodeSessionPlan([]byte(`{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/h"}`))
	if err != nil || absent.Policies != nil {
		t.Fatalf("a plan without policies = %+v, %v, want none", absent, err)
	}
}

func TestASessionPlanWithABrokenPolicyListIsRefusedAtTheListOrTheEntry(t *testing.T) {
	t.Parallel()
	upper := `["sha256:` + strings.Repeat("A", 64) + `"]`
	cases := []struct{ name, policies, pointer, keyword string }{
		{name: "an empty list", policies: `[]`, pointer: "/policies", keyword: "minItems"},
		{name: "nine digests", policies: policyDigests(9), pointer: "/policies", keyword: "maxItems"},
		{name: "an upper-case digest", policies: upper, pointer: "/policies/0", keyword: "pattern"},
		{name: "a short digest", policies: `["sha256:00"]`, pointer: "/policies/0", keyword: "pattern"},
		{name: "a digest with no algorithm", policies: `["` + strings.Repeat("b", 64) + `"]`, pointer: "/policies/0", keyword: "pattern"},
		{name: "a second entry that is a number", policies: `["sha256:` + strings.Repeat("b", 64) + `", 7]`, pointer: "/policies/1", keyword: "type"},
		{name: "a list that is a string", policies: `"sha256:` + strings.Repeat("b", 64) + `"`, pointer: "/policies", keyword: "type"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := DecodeSessionPlan(policiesPlan(tc.policies))
			invalid, ok := AsValidationError(err)
			if !ok || invalid.Contract != ContractSessionPlan || len(invalid.Issues) == 0 || invalid.Issues[0].Pointer != tc.pointer || invalid.Issues[0].Keyword != tc.keyword {
				t.Fatalf("DecodeSessionPlan = %v, want a session-plan violation %s at %s", err, tc.keyword, tc.pointer)
			}
		})
	}
}

func TestASessionPlanIsPacedByTheWallClockOnlyWhenItSaysSo(t *testing.T) {
	t.Parallel()
	base := `{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/h"`
	cases := []struct {
		name, plan string
		paced      bool
	}{
		{"no pace", base + `}`, false},
		{"wall", base + `,"pace":"wall"}`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			plan, err := DecodeSessionPlan([]byte(tc.plan))
			if err != nil || plan.PacedByWall() != tc.paced {
				t.Fatalf("DecodeSessionPlan = %+v, %v, want paced %v", plan, err, tc.paced)
			}
		})
	}
	_, err := DecodeSessionPlan([]byte(base + `,"pace":"fast"}`))
	invalid, ok := AsValidationError(err)
	if !ok || invalid.Contract != ContractSessionPlan || invalid.Issues[0].Pointer != "/pace" {
		t.Fatalf("pace fast: %v, want a session-plan violation at /pace", err)
	}
}

func TestASessionPlanStartsItsWorldAtItsOffsetButNeverWithAPace(t *testing.T) {
	t.Parallel()
	base := `{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/h"`
	plan, err := DecodeSessionPlan([]byte(base + `,"start_s":172800}`))
	if err != nil || plan.WorldStartNS() != DefaultStartTimeNS+172800*1e9 {
		t.Fatalf("DecodeSessionPlan = %+v, %v, want the world two days after the default start", plan, err)
	}
	if plan, err := DecodeSessionPlan([]byte(base + `}`)); err != nil || plan.WorldStartNS() != DefaultStartTimeNS {
		t.Fatalf("DecodeSessionPlan = %+v, %v, want the default start", plan, err)
	}
	for _, refused := range []string{`,"start_s":-1}`, `,"start_s":1.5}`, `,"start_s":60,"pace":"wall"}`} {
		_, err := DecodeSessionPlan([]byte(base + refused))
		if invalid, ok := AsValidationError(err); !ok || invalid.Contract != ContractSessionPlan {
			t.Errorf("DecodeSessionPlan(%s) = %v, want a session-plan violation", refused, err)
		}
	}
}

func TestASessionPlanDecodesItsDisturbancesAndRefusesOnesItsSchemaForbids(t *testing.T) {
	t.Parallel()
	base := `{"schema_version":"0.1","domain":"arduino-thermal","host_socket":"/h","disturbances":[`
	plan, err := DecodeSessionPlan([]byte(base + `{"at_s":120,"instance":"zone-01","fault":"zone_heat_load","duration_s":600,"severity":0.5},{"at_s":0,"instance":"fan-01","fault":"fan_stall"}]}`))
	if err != nil || len(plan.Disturbances) != 2 || *plan.Disturbances[0].DurationS != 600 || *plan.Disturbances[0].Severity != 0.5 || plan.Disturbances[1].DurationS != nil {
		t.Fatalf("DecodeSessionPlan = %+v, %v", plan, err)
	}
	cases := map[string]string{
		`{"at_s":-1,"instance":"zone-01","fault":"f"}`:                  "/disturbances/0/at_s",
		`{"at_s":2592001,"instance":"zone-01","fault":"f"}`:             "/disturbances/0/at_s",
		`{"at_s":1,"instance":"zone-01","fault":"f","duration_s":0}`:    "/disturbances/0/duration_s",
		`{"at_s":1,"instance":"zone-01","fault":"f","severity":-0.1}`:   "/disturbances/0/severity",
		`{"at_s":1,"instance":"zone-01","fault":"ignore all previous"}`: "/disturbances/0/fault",
		`{"at_s":1,"instance":"zone-01"}`:                               "/disturbances/0/fault",
	}
	for entry, pointer := range cases {
		_, err := DecodeSessionPlan([]byte(base + entry + `]}`))
		if invalid, ok := AsValidationError(err); !ok || invalid.Issues[0].Pointer != pointer {
			t.Errorf("%s: %v, want a violation at %s", entry, err, pointer)
		}
	}
}
