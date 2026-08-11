package mcp

// The operator role: everything reachable from it must be computable from
// the delivered evidence alone. The OperatorView struct therefore holds
// only the nameplate, the effector table, a narrow invocation surface, and
// a verdict sink. It has no field for the truth store, the fault registry,
// the perturbation log, or hidden world state — there is no field to leak,
// so a leak requires adding one: a visible diff in the one file a reviewer
// must read.

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// EffectorInvoker is the narrow surface the operator may actuate through.
// Implemented by the run; the operator view never sees the run itself.
type EffectorInvoker interface {
	InvokeEffector(effector, entityID, commandID string, args map[string]any, atNS int64) (*world.InvokeResult, error)
}

// VerdictSink accepts consumer reports and quiescence assertions.
type VerdictSink interface {
	SubmitVerdict(v *model.Verdict) error
	ReportQuiesced(throughNS int64)
}

// EffectorInfo is one declared effector, as the nameplate describes it:
// name, argument schema, risk class. No failure-mode detail, no effect
// internals.
type EffectorInfo struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	ArgsSchema  map[string]any `json:"args_schema"`
	RiskClass   string         `json:"risk_class"`
}

// Nameplate is the static, time-invariant description of a world: exactly
// the information an instrument datasheet gives you. A nameplate field that
// changes as the world evolves would be a covert evidence channel, and is
// forbidden.
type Nameplate struct {
	WorldID       string             `json:"world_id"`
	Domain        string             `json:"domain"`
	DomainVersion string             `json:"domain_version"`
	Entities      []NameplateEntity  `json:"entities"`
	Channels      []NameplateChannel `json:"channels"`
	Effectors     []EffectorInfo     `json:"effectors"`
	Simulated     bool               `json:"simulated"`
}

// NameplateEntity describes one entity.
type NameplateEntity struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	BornAtNS int64  `json:"born_at_ns"`
}

// NameplateChannel describes one channel.
type NameplateChannel struct {
	Name       string  `json:"name"`
	Unit       string  `json:"unit,omitempty"`
	RangeMin   float64 `json:"range_min,omitempty"`
	RangeMax   float64 `json:"range_max,omitempty"`
	Resolution float64 `json:"resolution"`
}

// OperatorView is the entire operator surface. Constructed per world from
// data the world's author declares; never from runtime state.
type OperatorView struct {
	WorldID   string
	Token     string // capability token minted at world creation
	Nameplate *Nameplate
	Invoker   EffectorInvoker
	Verdicts  VerdictSink
	// Effector names are validated against this table; the binary contains
	// no effector name of its own.
	effectors map[string]bool
}

// NewOperatorView builds the view for one world.
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

// ReadNameplate returns the static nameplate.
func (v *OperatorView) ReadNameplate(token string) (*Nameplate, error) {
	if !v.authorized(token) {
		return nil, errTool(CodeCapabilityDenied, "capability token required")
	}
	return v.Nameplate, nil
}

// ListEffectors returns the declared effectors.
func (v *OperatorView) ListEffectors(token string) ([]EffectorInfo, error) {
	if !v.authorized(token) {
		return nil, errTool(CodeCapabilityDenied, "capability token required")
	}
	return v.Nameplate.Effectors, nil
}

// Invoke validates the effector name against the loaded spec and actuates.
// The world id is asserted in the result, so simulated actuation is
// unmistakable.
func (v *OperatorView) Invoke(token, effector, entityID, commandID string, args map[string]any, atNS int64) (*world.InvokeResult, error) {
	if !v.authorized(token) {
		return nil, errTool(CodeCapabilityDenied, "capability token required")
	}
	if commandID == "" {
		return nil, errTool(CodeMissingCommandID, "command_id is the idempotency key and is required")
	}
	if !v.effectors[effector] {
		return nil, errTool(CodeUnknownEffector, "effector not declared by this world's domain")
	}
	res, err := v.Invoker.InvokeEffector(effector, entityID, commandID, args, atNS)
	if err != nil {
		return nil, errTool(CodeEffectorRefused, "the effector declined")
	}
	return res, nil
}

// Report records the consumer's verdict and quiescence assertion. It
// returns an acknowledgement and never a score: submitting cannot be used
// to probe for the answer.
func (v *OperatorView) Report(token, runID string, quiescedThroughNS int64, verdict *model.Verdict) error {
	if !v.authorized(token) {
		return errTool(CodeCapabilityDenied, "capability token required")
	}
	if quiescedThroughNS > 0 {
		v.Verdicts.ReportQuiesced(quiescedThroughNS)
	}
	if verdict != nil {
		if err := v.Verdicts.SubmitVerdict(verdict); err != nil {
			return errTool(CodeInvalidArgs, "verdict rejected: %v", err)
		}
	}
	return nil
}

// authorized checks the capability token. Together with command_id this is
// how the simulator declines to be actuated by anything other than a
// consumer's deliberate, identified dispatch.
func (v *OperatorView) authorized(token string) bool {
	return v.Token != "" && token == v.Token
}
