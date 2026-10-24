package domain

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestDescribeListsEveryCatalogKindInOrderAsADeliveryPerturbation(t *testing.T) {
	t.Parallel()
	kinds := Describe()
	names := make([]string, len(kinds))
	for i, kind := range kinds {
		names[i] = kind.Kind
		if kind.Category != CategoryDelivery || kind.Description == "" || kind.Params == nil {
			t.Fatalf("%s = %+v, want a delivery category, a description and a parameter list", kind.Kind, kind)
		}
	}
	if !slices.Equal(names, CatalogNames()) {
		t.Fatalf("described %v, want the catalog %v", names, CatalogNames())
	}
}

func TestDescribeStatesTheParameterContractOfAKind(t *testing.T) {
	t.Parallel()
	zero, one, big := 0.0, 1.0, 1e9
	for name, tc := range map[string]struct {
		kind string
		want []ParamInfo
	}{
		"a rate":            {Drop, []ParamInfo{{Name: "rate", Type: ParamNumber, Minimum: &zero, Maximum: &one}}},
		"an enumeration":    {ClockSkew, []ParamInfo{{Name: "offset_s", Type: ParamNumber, Minimum: &zero, Maximum: &big}, {Name: "sign", Type: ParamEnum, Values: []string{"positive", "negative"}}}},
		"a list of strings": {InjectionProbe, []ParamInfo{{Name: "payloads", Type: ParamStringList}}},
		"no parameters":     {Malformed, []ParamInfo{}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := describedKind(t, tc.kind).Params; !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("%s params = %+v, want %+v", tc.kind, got, tc.want)
			}
		})
	}
}

func TestEveryDescribedBoundIsTheBoundCheckParamsEnforces(t *testing.T) {
	t.Parallel()
	for _, kind := range Describe() {
		for _, param := range kind.Params {
			t.Run(kind.Kind+"/"+param.Name, func(t *testing.T) {
				t.Parallel()
				requireBoundsEnforced(t, kind.Kind, param)
			})
		}
	}
}

func requireBoundsEnforced(t *testing.T, kind string, param ParamInfo) {
	t.Helper()
	switch param.Type {
	case ParamNumber, ParamInteger:
		requireAccepted(t, kind, param.Name, *param.Minimum)
		requireAccepted(t, kind, param.Name, *param.Maximum)
		requireRefused(t, kind, param.Name, *param.Minimum-1)
		requireRefused(t, kind, param.Name, *param.Maximum*2+1)
	case ParamEnum:
		requireAccepted(t, kind, param.Name, param.Values[0])
		requireRefused(t, kind, param.Name, "no_such_value")
	case ParamStringList:
		requireAccepted(t, kind, param.Name, []any{"x"})
		requireRefused(t, kind, param.Name, []any{})
	default:
		t.Fatalf("%s.%s has the undeclared type %q", kind, param.Name, param.Type)
	}
}

func TestAnIntegerParameterRefusesAFraction(t *testing.T) {
	t.Parallel()
	requireAccepted(t, Reorder, "max_displacement", 3.0)
	requireRefused(t, Reorder, "max_displacement", 2.5)
}

func TestDescribeReturnsACopy(t *testing.T) {
	t.Parallel()
	first := Describe()
	first[0].Description, first[0].Params = "changed", nil
	*describedKind(t, Drop).Params[0].Maximum = 99
	if again := describedKind(t, Drop); again.Params[0].Maximum == nil || *again.Params[0].Maximum != 1 || Describe()[0].Description == "changed" {
		t.Fatalf("changing the result changed the catalog: %+v", again)
	}
}

func describedKind(t *testing.T, kind string) KindInfo {
	t.Helper()
	for _, info := range Describe() {
		if info.Kind == kind {
			return info
		}
	}
	t.Fatalf("%s is not described", kind)
	return KindInfo{}
}

func requireAccepted(t *testing.T, kind, param string, value any) {
	t.Helper()
	if err := CheckParams(kind, map[string]any{param: value}); err != nil {
		t.Fatalf("%s %s=%v refused, want accepted: %v", kind, param, value, err)
	}
}

func requireRefused(t *testing.T, kind, param string, value any) {
	t.Helper()
	err := CheckParams(kind, map[string]any{param: value})
	if err == nil || !strings.Contains(err.Error(), kind+" "+param+":") {
		t.Fatalf("%s %s=%v gave %v, want a refusal naming the kind and the parameter", kind, param, value, err)
	}
}

func TestOnlyAKindThatReleasesWhatItHeldAtTheEndOfItsWindowRequiresAnEnd(t *testing.T) {
	t.Parallel()
	for _, kind := range Describe() {
		want := kind.Kind == GrossBackfill || kind.Kind == ProducerFlap
		if kind.RequiresEnd != want {
			t.Errorf("%s requires_end = %v, want %v", kind.Kind, kind.RequiresEnd, want)
		}
	}
}
