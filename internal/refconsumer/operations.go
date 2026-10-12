package refconsumer

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer/internal/app"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer/internal/domain"
)

func DefaultConfig() Config {
	return layer.DefaultConfig()
}

func Consume(options ConsumeOptions) (ConsumeSummary, error) {
	return app.Consume(options)
}

func (r *Runner) Process(trace []byte, endNS int64) (*model.Verdict, error) {
	return r.runner.Process(trace, endNS)
}

func (o *MCPOperator) Close() error {
	return o.operator.Close()
}

func (o *MCPOperator) Nameplate() (*Nameplate, error) {
	return o.operator.Nameplate()
}

func (o *MCPOperator) ListEffectors() ([]EffectorInfo, error) {
	return o.operator.ListEffectors()
}

func (o *MCPOperator) InvokeEffector(effector, entityID, commandID string, args map[string]any, atNS int64) (*model.EffectorAcknowledgement, error) {
	return o.operator.InvokeEffector(effector, entityID, commandID, args, atNS)
}

func (o *MCPOperator) SubmitVerdict(v *model.Verdict) error {
	return o.operator.SubmitVerdict(v)
}

func (o *MCPOperator) ReportQuiesced(throughNS int64) error {
	return o.operator.ReportQuiesced(throughNS)
}
