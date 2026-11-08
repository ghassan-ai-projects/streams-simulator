package domain

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const PolicySampleNS int64 = 1_000_000_000

var ErrUnknownPolicy = errors.New("run: the domain declares no such baseline policy")

type PolicyCall struct {
	Effector  string
	EntityID  string
	CommandID string
	Args      map[string]any
	AtNS      int64
}

type PolicyPlay struct {
	policy     model.BaselinePolicy
	effectors  map[string]*model.Effector
	nextSample int64
	since      map[string]int64
	fired      map[string]bool
	pending    []PolicyCall
	issued     int
}

func NewPolicyPlay(spec *model.DomainSpec, name string, startNS int64) (*PolicyPlay, error) {
	index := slices.IndexFunc(spec.GroundTruth.BaselinePolicies, func(p model.BaselinePolicy) bool { return p.Name == name })
	if index < 0 {
		return nil, fmt.Errorf("%w: %q", ErrUnknownPolicy, name)
	}
	return &PolicyPlay{policy: spec.GroundTruth.BaselinePolicies[index], effectors: domainEffectors(spec), nextSample: startNS,
		since: map[string]int64{}, fired: map[string]bool{}}, nil
}

func domainEffectors(spec *model.DomainSpec) map[string]*model.Effector {
	effectors := map[string]*model.Effector{}
	for i := range spec.Kinds {
		for j := range spec.Kinds[i].Effectors {
			effectors[spec.Kinds[i].Effectors[j].Name] = &spec.Kinds[i].Effectors[j]
		}
	}
	return effectors
}

func (p *PolicyPlay) Kind() string { return p.policy.When.Kind }

func (p *PolicyPlay) Condition() model.Condition {
	when := p.policy.When
	return model.Condition{State: when.State, Operator: when.Operator, Threshold: when.Threshold}
}

func (p *PolicyPlay) NextStop(throughNS int64) (int64, bool) {
	stop := p.nextSample
	if len(p.pending) > 0 {
		stop = min(stop, p.pending[0].AtNS)
	}
	return stop, stop <= throughNS
}

func (p *PolicyPlay) SampleDue(atNS int64) bool { return atNS == p.nextSample }

func (p *PolicyPlay) Observe(atNS int64, holding map[string]bool) {
	for _, entity := range slices.Sorted(maps.Keys(holding)) {
		p.observeEntity(atNS, entity, holding[entity])
	}
	p.nextSample = atNS + PolicySampleNS
}

func (p *PolicyPlay) observeEntity(atNS int64, entity string, holds bool) {
	if !holds {
		delete(p.since, entity)
		delete(p.fired, entity)
		return
	}
	since, held := p.since[entity]
	if !held {
		since, p.since[entity] = atNS, atNS
	}
	if !p.fired[entity] && atNS-since >= seconds(p.policy.ForS) {
		p.fired[entity] = true
		p.fire(atNS, entity)
	}
}

func (p *PolicyPlay) fire(atNS int64, entity string) {
	for _, call := range p.policy.Then {
		target := call.Target
		if call.TargetsMatched() {
			target = entity
		}
		p.issued++
		p.pending = append(p.pending, PolicyCall{Effector: call.Effector, EntityID: target, AtNS: atNS + seconds(call.DelayS),
			CommandID: fmt.Sprintf("policy-%s-%d", p.policy.Name, p.issued), Args: p.callArgs(call, target)})
	}
	slices.SortStableFunc(p.pending, func(a, b PolicyCall) int { return cmp.Compare(a.AtNS, b.AtNS) })
}

func (p *PolicyPlay) callArgs(call model.PolicyCall, target string) map[string]any {
	args := maps.Clone(call.Args)
	if args == nil {
		args = map[string]any{}
	}
	for _, name := range p.effectors[call.Effector].EntityArguments() {
		args[name] = target
	}
	return args
}

func (p *PolicyPlay) Due(atNS int64) []PolicyCall {
	count := 0
	for count < len(p.pending) && p.pending[count].AtNS <= atNS {
		count++
	}
	due := slices.Clone(p.pending[:count])
	p.pending = p.pending[count:]
	return due
}

func seconds(s float64) int64 { return int64(math.Round(s * 1e9)) }
