package domain

import (
	"errors"
	"testing"
)

func capabilityView() *OperatorView {
	np := &Nameplate{WorldID: "w-1", Effectors: []EffectorInfo{{Name: "start_aerator"}}}
	return NewOperatorView("w-1", "t-secret", np, nil, nil)
}

func TestOperatorViewReadsOnlyWithItsCapabilityToken(t *testing.T) {
	t.Parallel()
	view := capabilityView()
	var refusal *ToolError
	if _, err := view.ReadNameplate("t-wrong"); !errors.As(err, &refusal) || refusal.Code != CodeCapabilityDenied {
		t.Fatalf("nameplate with a wrong token: %v", err)
	}
	if _, err := view.ListEffectors(""); !errors.As(err, &refusal) || refusal.Code != CodeCapabilityDenied {
		t.Fatalf("effectors without a token: %v", err)
	}
	np, err := view.ReadNameplate("t-secret")
	if err != nil || np.WorldID != "w-1" {
		t.Fatalf("nameplate = %+v, %v", np, err)
	}
	effectors, err := view.ListEffectors("t-secret")
	if err != nil || len(effectors) != 1 || effectors[0].Name != "start_aerator" {
		t.Fatalf("effectors = %+v, %v", effectors, err)
	}
}

func TestInvokeRefusesWhatIsNotDeclaredOrNotIdentified(t *testing.T) {
	t.Parallel()
	view := capabilityView()
	cases := map[string]struct {
		token, effector, commandID string
		want                       Code
	}{
		"wrong token":     {"t-wrong", "start_aerator", "c-1", CodeCapabilityDenied},
		"missing command": {"t-secret", "start_aerator", "", CodeMissingCommandID},
		"undeclared":      {"t-secret", "no_such", "c-1", CodeUnknownEffector},
	}
	for name, tc := range cases {
		var refusal *ToolError
		_, err := view.Invoke(tc.token, tc.effector, "e-1", tc.commandID, nil, 0)
		if !errors.As(err, &refusal) || refusal.Code != tc.want {
			t.Errorf("%s: err = %v, want code %s", name, err, tc.want)
		}
	}
}

func TestToolErrorRendersItsCodeAndMessage(t *testing.T) {
	t.Parallel()
	if got := ToolErrorf(CodeWorldNotFound, "unknown world %q", "w-9").Error(); got != `world_not_found: unknown world "w-9"` {
		t.Fatalf("error text = %q", got)
	}
}
