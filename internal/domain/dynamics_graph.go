package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type dynamicsTraversal struct {
	dependencies map[string][]string
	state        map[string]uint8
	stack        []string
	source       string
}

func dynamicsDependencies(spec *model.DomainSpec) map[string][]string {
	deps := make(map[string][]string)
	for _, d := range spec.Dynamics {
		if d.F1 == nil {
			continue
		}
		for _, in := range d.F1.Inputs {
			deps[d.Target] = append(deps[d.Target], in.State)
		}
	}
	return deps
}

func (traversal *dynamicsTraversal) visit(name string) error {
	switch traversal.state[name] {
	case 1:
		return fmt.Errorf("domain: %s: dynamics dependency cycle at state %q (path %v)", traversal.source, name, append(traversal.stack, name))
	case 2:
		return nil
	}
	return traversal.visitUnseen(name)
}

func (traversal *dynamicsTraversal) visitUnseen(name string) error {
	traversal.state[name] = 1
	traversal.stack = append(traversal.stack, name)
	if err := traversal.visitDependencies(name); err != nil {
		return err
	}
	traversal.stack = traversal.stack[:len(traversal.stack)-1]
	traversal.state[name] = 2
	return nil
}

func (traversal *dynamicsTraversal) visitDependencies(name string) error {
	for _, dep := range traversal.dependencies[name] {
		if err := traversal.visit(dep); err != nil {
			return err
		}
	}
	return nil
}
