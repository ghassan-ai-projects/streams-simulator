package domain

import (
	"fmt"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	NoteOrigin  = "origin"
	NoteBinding = "binding"
)

type Note struct {
	Type   string
	Fields map[string]any
}

type relayPlan struct {
	bound     []Bound
	instances map[string]string
}

func (r *Relay) Plan(simInstances map[string]string, bound []Bound) {
	r.plan = relayPlan{bound: bound, instances: simInstances}
}

func BindingIssues(gateway Message, bound []Bound) []model.Issue {
	issues := []model.Issue{}
	for _, b := range bound {
		if issue, broken := bindingIssue(gateway, b); broken {
			issues = append(issues, issue)
		}
	}
	return issues
}

func bindingIssue(gateway Message, b Bound) (model.Issue, bool) {
	at := fmt.Sprintf("/real/%d", b.Index)
	board := slices.IndexFunc(gateway.Boards, func(entry Board) bool { return entry.Board == b.Board })
	switch {
	case board < 0:
		return model.RuleIssue(at+"/board", "board_missing", "board", fmt.Sprintf("the gateway announces no board %q", clip(b.Board))), true
	case !slices.Contains(gateway.Boards[board].Members, b.Device):
		return model.RuleIssue(at+"/device", "device_missing", "device",
			fmt.Sprintf("board %q of the gateway holds no device %q", clip(b.Board), clip(b.Device))), true
	}
	return kindIssue(gateway.Kinds, b, at)
}

func kindIssue(kinds []any, b Bound, at string) (model.Issue, bool) {
	at += "/instance"
	index := slices.IndexFunc(kinds, func(entry any) bool { return kindNamed(entry) == b.Kind })
	switch {
	case index < 0:
		return model.RuleIssue(at, "kind_missing", "instance",
			fmt.Sprintf("the gateway announces no kind %q, the kind file of %s", clip(b.Kind), clip(b.Instance))), true
	case digestOfKind(kinds[index]) != b.KindDigest:
		return model.RuleIssue(at, "kind_mismatch", "instance", fmt.Sprintf(
			"the gateway announces kind %q with kind_digest %s, the kind file of %s digests to %s",
			clip(b.Kind), clip(digestOfKind(kinds[index])), clip(b.Instance), b.KindDigest)), true
	}
	return model.Issue{}, false
}

func (r *Relay) bindingRefusal(gateway Message) string {
	if issues := BindingIssues(gateway, r.plan.bound); len(issues) > 0 {
		return issues[0].Message
	}
	return ""
}

func (r *Relay) noteBinding(m *Moves, in Inbound, refusal, reason string) {
	if len(r.plan.bound) == 0 {
		return
	}
	m.Notes = append(m.Notes, bindingNote(refusal, reason, BindingIssues(in.Message, r.plan.bound)))
}

func bindingNote(refusal, reason string, issues []model.Issue) Note {
	fields := map[string]any{"outcome": "matched", "issues": issues}
	if refusal != "" {
		fields["outcome"], fields["refusal"], fields["reason"] = "refused", refusal, reason
	}
	return Note{Type: NoteBinding, Fields: fields}
}

func (r *Relay) boundAt(in Inbound) (Bound, bool) {
	board := text(in.Record, "board")
	index := slices.IndexFunc(r.plan.bound, func(b Bound) bool { return b.Board == board && b.Device == in.Device })
	if index < 0 {
		return Bound{}, false
	}
	return r.plan.bound[index], true
}

func (r *Relay) misbound(in Inbound) (Bound, bool) {
	b, bound := r.boundAt(in)
	return b, in.Type == "DBIRTH" && bound && in.Kind != b.Kind
}

func (r *Relay) noteMisbound(m *Moves, in Inbound) {
	b, misbound := r.misbound(in)
	if !misbound {
		return
	}
	issue := model.RuleIssue(fmt.Sprintf("/real/%d/device", b.Index), "device_kind_mismatch", "device",
		fmt.Sprintf("device %q was born as kind %q, %s is of kind %q", clip(in.Device), clip(in.Kind), clip(b.Instance), clip(b.Kind)))
	m.Notes = append(m.Notes, bindingNote("device_kind_mismatch", issue.Message, []model.Issue{issue}))
}

func (r *Relay) noteOrigin(m *Moves, in Inbound, environment string) {
	if r.sim == nil || in.Type != "DBIRTH" {
		return
	}
	m.Notes = append(m.Notes, OriginNote(in.Record, environment, r.instanceOf(in, environment)))
}

func OriginNote(birth map[string]any, environment, instance string) Note {
	fields := map[string]any{"device": text(birth, "device"), "board": text(birth, "board"), "environment": environment,
		"boot_id": text(birth, "boot_id")}
	if instance != "" {
		fields["instance"] = instance
	}
	return Note{Type: NoteOrigin, Fields: fields}
}

func (r *Relay) instanceOf(in Inbound, environment string) string {
	if environment == EnvironmentSimulated {
		return r.plan.instances[in.Device]
	}
	if _, misbound := r.misbound(in); misbound {
		return ""
	}
	b, _ := r.boundAt(in)
	return b.Instance
}
