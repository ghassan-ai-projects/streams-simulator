package domain

import (
	"fmt"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var unfollowableForms = []string{model.FormStore, model.FormPeak, model.FormTrip, model.FormAntiRecycle, model.FormSequencer}

func (checks domainChecks) f1Input(form string, in model.F1Input) error {
	if in.Link != nil {
		return checks.linkInput(in)
	}
	if !checks.hasState(in.State) {
		return checks.bad("", "f1 form %q input %q is not a declared state", form, in.State)
	}
	return nil
}

func (checks domainChecks) linkInput(in model.F1Input) error {
	if err := checks.linkRead(in.Link.Type, in.State, "/link/type", "/state"); err != nil {
		return err
	}
	if in.Link.NeedsDefault() && in.Link.Default == nil {
		return checks.bad("/link/default", "input %q takes the %s over link %q and needs a default for an instance with no linked instances", in.State, in.Link.Aggregate, in.Link.Type)
	}
	return nil
}

func (checks domainChecks) dynamicFollow(d *model.Dynamics) error {
	if d.Follow == nil {
		return nil
	}
	if d.F1 != nil && slices.Contains(unfollowableForms, d.F1.Form) {
		return checks.bad("/follow", "state %q keeps its own %s form and cannot follow linked instances", d.Target, d.F1.Form)
	}
	return checks.linkRead(d.Follow.Link, d.Follow.State, "/follow/link", "/follow/state")
}

func (checks domainChecks) linkRead(linkType, state, typeMember, stateMember string) error {
	lt := checks.linkType(linkType)
	if lt == nil {
		return checks.bad(typeMember, "link type %q is not declared", linkType)
	}
	if lt.To != checks.kind.Name {
		return checks.bad(typeMember, "link type %q ends at kind %q, so kind %q cannot read over it", linkType, lt.To, checks.kind.Name)
	}
	if source := checks.compiled.Kind(lt.From); source == nil || source.StateDecl(state) == nil {
		return checks.bad(stateMember, "kind %q, linked over %q, declares no state %q", lt.From, linkType, state)
	}
	return nil
}

func (checks domainChecks) checkCoupling() error {
	graph := checks.couplingGraph()
	kinds := checks.compiled.Spec.Kinds
	traversal := cycleTraversal{dependencies: graph, state: map[string]uint8{}, cycle: func(name string, path []string) error {
		at := itemPointer("kinds", slices.IndexFunc(kinds, func(k model.Kind) bool { return k.Name == name }))
		return invalidDomain(checks.source, at, fmt.Sprintf("kinds read each other over links in a cycle (path %v); a link read must not lead back to its own kind", path))
	}}
	for i := range kinds {
		if err := traversal.visit(kinds[i].Name); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) couplingGraph() map[string][]string {
	graph := map[string][]string{}
	for _, kind := range checks.readingKinds() {
		for _, linkType := range kind.LinkReads() {
			if lt := checks.linkType(linkType); lt != nil && !slices.Contains(graph[kind.Name], lt.From) {
				graph[kind.Name] = append(graph[kind.Name], lt.From)
			}
		}
	}
	return graph
}

func (checks domainChecks) readingKinds() []*model.Kind {
	spec := checks.compiled.Spec
	kinds := make([]*model.Kind, 0, len(spec.Kinds)+len(spec.Instances))
	for i := range spec.Kinds {
		kinds = append(kinds, &spec.Kinds[i])
	}
	for set := range spec.Instances {
		kinds = append(kinds, checks.compiled.InstanceKind(set))
	}
	return kinds
}
