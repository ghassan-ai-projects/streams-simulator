package domain

import (
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func violationIn(t *testing.T, path string, change func(doc map[string]any), pointer string) {
	t.Helper()
	_, err := Parse(documentOf(t, path, change), "stability.json")
	var invalid *model.ValidationError
	if pointer == "" {
		if err != nil {
			t.Fatalf("want a valid domain, got %v", err)
		}
		return
	}
	if !errors.As(err, &invalid) || len(invalid.Issues) != 1 || invalid.Issues[0].Pointer != pointer {
		t.Fatalf("want one issue at %s, got %v", pointer, err)
	}
}

func TestAnIntegratedStateMustStepWithinItsStabilityBound(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		path    string
		change  func(d map[string]any)
		pointer string
	}{
		{"a lag at dt/τ 2.8", testsupport.CoupledDomain(), func(d map[string]any) {
			put(d, 25.0, "kinds", 1, "dynamics", 3, "f1", "time_constant_s")
		}, "/kinds/1/dynamics/3/dt_ms"},
		{"a lag at dt/τ 2.33", testsupport.CoupledDomain(), func(d map[string]any) {
			put(d, 30.0, "kinds", 1, "dynamics", 3, "f1", "time_constant_s")
		}, ""},
		{"a stiff outlet", testsupport.TransferDomain(), func(d map[string]any) {
			put(d, 10.0, "kinds", 0, "flows", 1, "conductance")
		}, "/kinds/0/dynamics/2/dt_ms"},
		{"a transfer inside the bound at full pump speed", testsupport.TransferDomain(), func(d map[string]any) {
			put(d, 1.3, "kinds", 0, "flows", 0, "conductance")
		}, ""},
		{"a transfer stiff only at full pump speed", testsupport.TransferDomain(), func(d map[string]any) {
			put(d, 1.3, "kinds", 0, "flows", 0, "conductance")
			put(d, 2.2, "kinds", 0, "state", 1, "max")
		}, "/kinds/0/dynamics/2/dt_ms"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			violationIn(t, tc.path, tc.change, tc.pointer)
		})
	}
}

func TestEachIntegratedFormHasItsStabilityLimit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		f1    *model.F1Dyn
		limit float64
		taus  []float64
	}{
		{&model.F1Dyn{Form: "first_order_lag", TimeConstantS: 5}, rk4StableRatio, []float64{5, 0}},
		{&model.F1Dyn{Form: "rc_network", TimeConstantS: 5, TimeConstant2S: 7}, rk4StableRatio, []float64{5, 7}},
		{&model.F1Dyn{Form: "dead_time", TimeConstantS: 3}, eulerStableRatio, []float64{3}},
		{&model.F1Dyn{Form: model.FormRegime, Below: &model.Regime{TimeConstantS: 2}}, rk4StableRatio, []float64{2, 0}},
		{&model.F1Dyn{Form: "integrator"}, 0, nil},
	}
	for _, tc := range cases {
		limit, taus := stabilityLimit(tc.f1)
		if limit != tc.limit || len(taus) != len(tc.taus) {
			t.Fatalf("%s: limit %v taus %v", tc.f1.Form, limit, taus)
		}
		for i := range taus {
			if taus[i] != tc.taus[i] {
				t.Fatalf("%s: taus %v, want %v", tc.f1.Form, taus, tc.taus)
			}
		}
	}
}

func TestALeaseArgumentMustNotAllowANegativeLease(t *testing.T) {
	t.Parallel()
	runFixtureViolations(t, testsupport.UtilityDomain(), []violationCase{
		{"no lower bound", func(d map[string]any) {
			delete(at(d, "kinds", 2, "effectors", 1, "args_schema", "properties", "lease_s").(map[string]any), "exclusiveMinimum")
		}, "/kinds/2/effectors/1/lease/arg"},
		{"a negative lower bound", func(d map[string]any) {
			put(d, -5.0, "kinds", 2, "effectors", 1, "args_schema", "properties", "lease_s", "exclusiveMinimum")
		}, "/kinds/2/effectors/1/lease/arg"},
	})
}
