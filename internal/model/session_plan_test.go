package model

import (
	"errors"
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
