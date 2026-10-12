package app

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

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
		boards = append(boards, model.Board{Name: board.Name, Members: d.wireMembers(board.Members)})
	}
	for _, id := range d.run.World.EntityIDs() {
		if d.domainBoard(id) == "" {
			boards = append(boards, model.Board{Name: d.wireOf(id), Members: []string{d.wireOf(id)}})
		}
	}
	return boards
}

func (d *devices) wireMembers(members []string) []string {
	out := make([]string, len(members))
	for i, member := range members {
		out[i] = d.wireOf(member)
	}
	return out
}
