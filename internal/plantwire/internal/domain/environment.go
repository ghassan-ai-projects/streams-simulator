package domain

import (
	"errors"
	"fmt"
	"slices"
)

const (
	EnvironmentSimulated = "simulated"
	EnvironmentHardware  = "hardware"
	EnvironmentMixed     = "mixed"
)

var ErrEnvironment = errors.New("plantwire: unknown environment")

type EnvironmentFault struct {
	Pointer string
	Reason  string
}

func CheckEnvironment(environment string) error {
	if !slices.Contains([]string{EnvironmentSimulated, EnvironmentHardware, EnvironmentMixed}, environment) {
		return fmt.Errorf("%w: %q is none of simulated, hardware, mixed", ErrEnvironment, clip(environment))
	}
	return nil
}

func EnvironmentClaimFault(m Message) *EnvironmentFault {
	switch m.Environment {
	case EnvironmentSimulated, EnvironmentHardware:
		return strayBoard(m)
	case EnvironmentMixed:
		return unmixed(m)
	}
	return &EnvironmentFault{Pointer: "/environment", Reason: fmt.Sprintf("environment %q is none of simulated, hardware, mixed", clip(m.Environment))}
}

func EnvironmentFaultFor(host string, m Message) *EnvironmentFault {
	if fault := EnvironmentClaimFault(m); fault != nil {
		return fault
	}
	if host == EnvironmentMixed || host == m.Environment {
		return nil
	}
	return &EnvironmentFault{Pointer: "/environment", Reason: fmt.Sprintf("the host accepts %s nodes only, the node announces %s", clip(host), m.Environment)}
}

func strayBoard(m Message) *EnvironmentFault {
	for i, board := range m.Boards {
		if board.Environment != m.Environment {
			return &EnvironmentFault{Pointer: fmt.Sprintf("/boards/%d/environment", i),
				Reason: fmt.Sprintf("node environment %s needs every board %s, board %q is %q", m.Environment, m.Environment, clip(board.Board), clip(board.Environment))}
		}
	}
	return nil
}

func unmixed(m Message) *EnvironmentFault {
	has := map[string]bool{}
	for _, board := range m.Boards {
		has[board.Environment] = true
	}
	if has[EnvironmentSimulated] && has[EnvironmentHardware] {
		return nil
	}
	missing := EnvironmentHardware
	if !has[EnvironmentSimulated] {
		missing = EnvironmentSimulated
	}
	return &EnvironmentFault{Pointer: "/boards", Reason: fmt.Sprintf("node environment mixed needs a simulated and a hardware board, none of its %d boards is %s", len(m.Boards), missing)}
}

func AnnouncedEnvironment(boards map[string]bool) string {
	switch {
	case boards[EnvironmentSimulated] && boards[EnvironmentHardware]:
		return EnvironmentMixed
	case boards[EnvironmentSimulated]:
		return EnvironmentSimulated
	}
	return EnvironmentHardware
}
