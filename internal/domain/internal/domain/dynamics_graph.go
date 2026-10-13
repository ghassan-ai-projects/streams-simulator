package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type cycleTraversal struct {
	dependencies map[string][]string
	state        map[string]uint8
	stack        []string
	cycle        func(name string, path []string) error
}

func dynamicsDependencies(kind *model.Kind) map[string][]string {
	deps := make(map[string][]string)
	for _, d := range kind.Dynamics {
		if d.F1 == nil {
			continue
		}
		for _, in := range d.F1.Inputs {
			if in.Link == nil {
				deps[d.Target] = append(deps[d.Target], in.State)
			}
		}
	}
	return deps
}

func (traversal *cycleTraversal) visit(name string) error {
	switch traversal.state[name] {
	case 1:
		return traversal.cycle(name, append(traversal.stack, name))
	case 2:
		return nil
	}
	return traversal.visitUnseen(name)
}

func (traversal *cycleTraversal) visitUnseen(name string) error {
	traversal.state[name] = 1
	traversal.stack = append(traversal.stack, name)
	if err := traversal.visitDependencies(name); err != nil {
		return err
	}
	traversal.stack = traversal.stack[:len(traversal.stack)-1]
	traversal.state[name] = 2
	return nil
}

func (traversal *cycleTraversal) visitDependencies(name string) error {
	for _, dep := range traversal.dependencies[name] {
		if err := traversal.visit(dep); err != nil {
			return err
		}
	}
	return nil
}

func dynamicsTargets(kind *model.Kind) map[string]int {
	targets := make(map[string]int)
	for i, d := range kind.Dynamics {
		targets[d.Target] = i
	}
	return targets
}
