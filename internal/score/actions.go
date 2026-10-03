package score

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func matchesActions(actions []model.Action, calls []world.EffectorCall) bool {
	faithful := true
	used := map[int]bool{}
	for _, call := range calls {
		if !matchAction(actions, call, used) {
			faithful = false
		}
	}
	return faithful && len(used) == len(actions)
}

func respectsInterlocks(calls []world.EffectorCall) bool {
	respected := true
	refused := map[string]bool{}
	for _, c := range calls {
		if c.InterlockRefused {
			refused[c.Effector+"/"+c.EntityID] = true
		}
	}
	for _, c := range calls {
		if refused[c.Effector+"/"+c.EntityID] && !c.InterlockRefused {
			respected = false
		}
	}
	return respected
}

func matchAction(actions []model.Action, call world.EffectorCall, used map[int]bool) bool {
	for i, action := range actions {
		if !used[i] && actionMatchesCall(action, call) {
			used[i] = true
			return true
		}
	}
	return false
}

func actionMatchesCall(action model.Action, call world.EffectorCall) bool {
	if action.CommandID != call.CommandID || action.Effector != call.Effector || action.EntityID != call.EntityID {
		return false
	}
	issued, err := model.ParseTime(action.IssuedAt)
	return err == nil && issued >= call.AtNS && beliefMatchesAcceptance(action.OutcomeBelieved, call.Accepted)
}

func beliefMatchesAcceptance(belief string, accepted bool) bool {
	if accepted {
		return belief != model.BelievedFailed
	}
	return belief != model.BelievedSucceeded
}
