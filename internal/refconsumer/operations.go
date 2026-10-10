package refconsumer

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// DefaultConfig is the reference detector's default tuning.
func DefaultConfig() Config {
	return layer.DefaultConfig()
}

// Process detects episodes in a native JSONL trace, actuates when configured
// and returns the verdict it reports.
func (r *Runner) Process(trace []byte, endNS int64) (*model.Verdict, error) {
	return r.runner.Process(trace, endNS)
}

// Close ends the operator session.
func (o *MCPOperator) Close() error {
	return o.operator.Close()
}

// Nameplate fetches the world description over the operator surface.
func (o *MCPOperator) Nameplate() (*Nameplate, error) {
	return o.operator.Nameplate()
}

// ListEffectors lists the effectors the operator surface exposes.
func (o *MCPOperator) ListEffectors() ([]EffectorInfo, error) {
	return o.operator.ListEffectors()
}

// InvokeEffector actuates an effector over the operator surface.
func (o *MCPOperator) InvokeEffector(effector, entityID, commandID string, args map[string]any, atNS int64) (*world.InvokeResult, error) {
	return o.operator.InvokeEffector(effector, entityID, commandID, args, atNS)
}

// SubmitVerdict reports the verdict over the operator surface.
func (o *MCPOperator) SubmitVerdict(v *model.Verdict) error {
	return o.operator.SubmitVerdict(v)
}

// ReportQuiesced reports the consumer's quiescence watermark.
func (o *MCPOperator) ReportQuiesced(throughNS int64) error {
	return o.operator.ReportQuiesced(throughNS)
}
