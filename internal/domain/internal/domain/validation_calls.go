package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonvalue"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (checks domainChecks) effectorResponder(e *model.Effector) error {
	r := e.Responder
	if r == nil {
		return nil
	}
	if !argumentOfType(e, r.MatchArg, "string") {
		return checks.bad("/responder/match_arg", "responder of %q matches argument %q, which its args_schema does not declare as a string", e.Name, r.MatchArg)
	}
	if err := checks.deltaStates(e, "/responder/on_match", r.OnMatch); err != nil {
		return err
	}
	return checks.deltaStates(e, "/responder/on_miss", r.OnMiss)
}

func (checks domainChecks) effectorLease(e *model.Effector) error {
	l := e.Lease
	if l == nil {
		return nil
	}
	if err := checks.leaseArgument(e, l.Arg); err != nil {
		return err
	}
	if err := checks.leaseFallback(e, l); err != nil {
		return err
	}
	return checks.safeStates(l.SafeDeltas)
}

func (checks domainChecks) leaseFallback(e *model.Effector, l *model.Lease) error {
	switch {
	case l.Fallback == model.FallbackSafeOff && len(l.SafeDeltas) == 0:
		return checks.bad("/lease/safe_deltas", "a safe_off lease of %q needs safe_deltas", e.Name)
	case l.Fallback == model.FallbackRevert && assignsState(e.Effect.StateDeltas):
		return checks.bad("/lease/fallback", "a revert lease of %q cannot undo an assigned state", e.Name)
	}
	return nil
}

func (checks domainChecks) leaseArgument(e *model.Effector, arg string) error {
	if !argumentOfType(e, arg, "number", "integer") {
		return checks.bad("/lease/arg", "lease of %q reads argument %q, which its args_schema does not declare as a number", e.Name, arg)
	}
	if !argumentNonNegative(e, arg) {
		return checks.bad("/lease/arg", "lease of %q reads argument %q, whose args_schema allows a negative lease; declare minimum or exclusiveMinimum of at least 0", e.Name, arg)
	}
	return nil
}

func (checks domainChecks) safeStates(deltas []model.StateDelta) error {
	for i, delta := range deltas {
		if delta.FromArg != "" || delta.AssignFromArg != "" {
			return checks.bad(fmt.Sprintf("/lease/safe_deltas/%d", i), "a safe state is a constant: from_arg and assign_from_arg do not apply")
		}
	}
	return checks.deltaStates(nil, "/lease/safe_deltas", deltas)
}

func (checks domainChecks) deltaStates(e *model.Effector, member string, deltas []model.StateDelta) error {
	for i, delta := range deltas {
		at := fmt.Sprintf("%s/%d", member, i)
		if !checks.hasState(delta.State) {
			return checks.bad(at+"/state", "delta on undeclared state %q", delta.State)
		}
		if err := checks.deltaArguments(e, at, delta); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) deltaArguments(e *model.Effector, at string, delta model.StateDelta) error {
	for _, read := range [][2]string{{"/from_arg", delta.FromArg}, {"/assign_from_arg", delta.AssignFromArg}} {
		if read[1] != "" && (e == nil || !argumentOfType(e, read[1], "number", "integer")) {
			return checks.bad(at+read[0], "delta reads argument %q, which is no numeric argument of the effector", read[1])
		}
	}
	return nil
}

func argumentOfType(e *model.Effector, arg string, kinds ...string) bool {
	properties, _ := e.ArgsSchema["properties"].(map[string]any)
	property, _ := properties[arg].(map[string]any)
	kind, _ := property["type"].(string)
	for _, want := range kinds {
		if kind == want {
			return true
		}
	}
	return false
}

func argumentNonNegative(e *model.Effector, arg string) bool {
	properties, _ := e.ArgsSchema["properties"].(map[string]any)
	property, _ := properties[arg].(map[string]any)
	for _, key := range []string{"minimum", "exclusiveMinimum"} {
		if bound, ok := jsonvalue.Number(property[key]); ok && bound >= 0 {
			return true
		}
	}
	return false
}

func assignsState(deltas []model.StateDelta) bool {
	for _, delta := range deltas {
		if delta.AssignFromArg != "" {
			return true
		}
	}
	return false
}
