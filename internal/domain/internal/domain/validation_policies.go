package domain

import (
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (checks domainChecks) checkBaselinePolicies() error {
	policies := checks.compiled.Spec.GroundTruth.BaselinePolicies
	return checks.within("/ground_truth").each("baseline_policies", len(policies), func(at domainChecks, i int) error {
		return at.baselinePolicy(policies[:i], &policies[i])
	})
}

func (checks domainChecks) baselinePolicy(earlier []model.BaselinePolicy, p *model.BaselinePolicy) error {
	if slices.ContainsFunc(earlier, func(e model.BaselinePolicy) bool { return e.Name == p.Name }) {
		return checks.bad("/name", "baseline policy %q is declared twice", p.Name)
	}
	kind, err := checks.policyCondition(p)
	if err != nil {
		return err
	}
	return checks.each("then", len(p.Then), func(at domainChecks, i int) error { return at.policyCall(kind, &p.Then[i]) })
}

func (checks domainChecks) policyCondition(p *model.BaselinePolicy) (*model.Kind, error) {
	kind := checks.compiled.Kind(p.When.Kind)
	if kind == nil {
		return nil, checks.bad("/when/kind", "baseline policy %q watches kind %q, which the domain does not declare", p.Name, p.When.Kind)
	}
	if kind.StateDecl(p.When.State) == nil {
		return nil, checks.bad("/when/state", "baseline policy %q watches %q, which is not a state of kind %q", p.Name, p.When.State, kind.Name)
	}
	return kind, nil
}

func (checks domainChecks) policyCall(matched *model.Kind, call *model.PolicyCall) error {
	if !checks.compiled.HasEffector(call.Effector) {
		return checks.bad("/effector", "the domain declares no effector %q", call.Effector)
	}
	target, err := checks.targetKind(matched, call)
	if err != nil {
		return err
	}
	effector := target.EffectorDecl(call.Effector)
	if effector == nil {
		return checks.bad("/effector", "the call goes to a %q instance, whose kind does not declare effector %q", target.Name, call.Effector)
	}
	return checks.policyArguments(effector, call.Args)
}

func (checks domainChecks) targetKind(matched *model.Kind, call *model.PolicyCall) (*model.Kind, error) {
	if call.TargetsMatched() {
		return matched, nil
	}
	if !checks.compiled.IsInstanceName(call.Target) {
		return nil, checks.bad("/target", "target %q is neither matched nor a declared instance name", call.Target)
	}
	set := checks.compiled.InstanceSetFor(call.Target)
	return checks.compiled.Kind(checks.compiled.Spec.Instances[set].Kind), nil
}
