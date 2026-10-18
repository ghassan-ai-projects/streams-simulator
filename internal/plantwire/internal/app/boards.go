package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func (d *devices) board(id string) string {
	if board := d.domainBoard(id); board != "" {
		return board
	}
	return d.wireOf(id)
}

func (d *devices) domainBoard(id string) string {
	for _, board := range d.spec.Spec.Boards {
		for _, member := range board.Members {
			if member == id {
				return board.Name
			}
		}
	}
	return ""
}

func (d *devices) boot(id string) string { return d.board(id) + "-boot-1" }

func (d *devices) boards() []model.Board {
	boards := make([]model.Board, 0, len(d.spec.Spec.Boards))
	for _, board := range d.spec.Spec.Boards {
		boards = simulatedBoard(boards, board.Name, d.wireMembers(board.Members))
	}
	for _, id := range d.run.World.EntityIDs() {
		if d.domainBoard(id) == "" {
			boards = simulatedBoard(boards, d.wireOf(id), d.wireMembers([]string{id}))
		}
	}
	return boards
}

func simulatedBoard(boards []model.Board, name string, members []string) []model.Board {
	if len(members) == 0 {
		return boards
	}
	return append(boards, model.Board{Name: name, Members: members, Environment: domain.EnvironmentSimulated})
}

func (d *devices) wireMembers(members []string) []string {
	out := make([]string, 0, len(members))
	for _, member := range members {
		if wireID := d.wireOf(member); !d.withheld[member] {
			out = append(out, wireID)
		}
	}
	return out
}

func (d *devices) safeStopsUnguarded() error {
	for _, kind := range d.kinds() {
		for i := range kind.Effectors {
			if eff := &kind.Effectors[i]; eff.SafeStop && (len(eff.Requires) > 0 || eff.Interlock != nil) {
				return fmt.Errorf("%w: safe stop %q carries preconditions or an interlock", ErrBadOptions, eff.Name)
			}
		}
	}
	return nil
}
