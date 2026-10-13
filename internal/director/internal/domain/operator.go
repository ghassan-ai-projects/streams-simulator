package domain

import (
	"errors"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

type EffectorInvoker interface {
	InvokeEffectorNow(effector, entityID, commandID string, args map[string]any) (*world.InvokeResult, error)
}

type VerdictSink interface {
	SubmitVerdict(v *model.Verdict) error
	ReportQuiesced(throughNS int64)
}

type EffectorInfo struct {
	Name       string         `json:"name"`
	ArgsSchema map[string]any `json:"args_schema"`
	RiskClass  string         `json:"risk_class"`
}

type Nameplate struct {
	WorldID       string             `json:"world_id"`
	Domain        string             `json:"domain"`
	DomainVersion string             `json:"domain_version"`
	Entities      []NameplateEntity  `json:"entities"`
	Channels      []NameplateChannel `json:"channels"`
	Effectors     []EffectorInfo     `json:"effectors"`
	Simulated     bool               `json:"simulated"`
}

type NameplateEntity struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	BornAtNS int64  `json:"born_at_ns"`
}

type NameplateChannel struct {
	Name       string  `json:"name"`
	Unit       string  `json:"unit,omitempty"`
	RangeMin   float64 `json:"range_min,omitempty"`
	RangeMax   float64 `json:"range_max,omitempty"`
	Resolution float64 `json:"resolution"`
}

type OperatorView struct {
	WorldID string
	Token   string

	RunID     string
	Nameplate *Nameplate
	Invoker   EffectorInvoker
	Verdicts  VerdictSink

	effectors map[string]bool
}

func NewOperatorView(worldID, token string, nameplate *Nameplate, invoker EffectorInvoker, verdicts VerdictSink) *OperatorView {
	eff := map[string]bool{}
	for _, e := range nameplate.Effectors {
		eff[e.Name] = true
	}
	return &OperatorView{
		WorldID: worldID, Token: token, Nameplate: nameplate,
		Invoker: invoker, Verdicts: verdicts, effectors: eff,
	}
}

func (v *OperatorView) ReadNameplate(token string) (*Nameplate, error) {
	if !v.authorized(token) {
		return nil, errTool(CodeCapabilityDenied, "capability token required")
	}
	return v.Nameplate, nil
}

func (v *OperatorView) ListEffectors(token string) ([]EffectorInfo, error) {
	if !v.authorized(token) {
		return nil, errTool(CodeCapabilityDenied, "capability token required")
	}
	return v.Nameplate.Effectors, nil
}

func (v *OperatorView) Invoke(token, effector, entityID, commandID string, args map[string]any) (*model.EffectorAcknowledgement, error) {
	if err := v.admitInvocation(token, effector, commandID); err != nil {
		return nil, err
	}
	res, err := v.Invoker.InvokeEffectorNow(effector, entityID, commandID, args)
	if errors.Is(err, world.ErrCommandIDReused) {
		return nil, errTool(CodeInvalidArgs, "command_id was already used for a different request")
	}
	if err != nil {
		return nil, errTool(CodeEffectorRefused, "the effector declined")
	}
	return world.Acknowledge(res), nil
}

func (v *OperatorView) Report(token, runID string, quiescedThroughNS int64, verdict *model.Verdict) error {
	if err := v.admitReport(token, runID); err != nil {
		return err
	}

	if err := v.submit(verdict); err != nil {
		return err
	}
	if quiescedThroughNS > 0 {
		v.Verdicts.ReportQuiesced(quiescedThroughNS)
	}
	return nil
}

func (v *OperatorView) submit(verdict *model.Verdict) error {
	if verdict == nil {
		return nil
	}
	if err := v.Verdicts.SubmitVerdict(verdict); err != nil {
		return errTool(CodeInvalidArgs, "verdict rejected: %v", err)
	}
	return nil
}

func (v *OperatorView) admitReport(token, runID string) error {
	if !v.authorized(token) {
		return errTool(CodeCapabilityDenied, "capability token required")
	}
	if v.RunID != "" && runID != v.RunID {
		return errTool(CodeInvalidArgs, "run_id does not name this world's run")
	}
	return nil
}

func (v *OperatorView) authorized(token string) bool {
	return v.Token != "" && token == v.Token
}

func (v *OperatorView) admitInvocation(token, effector, commandID string) error {
	if !v.authorized(token) {
		return errTool(CodeCapabilityDenied, "capability token required")
	}
	if commandID == "" {
		return errTool(CodeMissingCommandID, "command_id is the idempotency key and is required")
	}
	if !v.effectors[effector] {
		return errTool(CodeUnknownEffector, "effector not declared by this world's domain")
	}

	return nil
}
