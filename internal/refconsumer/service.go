package refconsumer

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer/internal/mcpclient"
)

// Runner is the reference consumer: it detects episodes in a trace and, when
// configured, actuates and reports through its ports.
type Runner struct {
	runner *layer.Runner
}

// New builds a runner over a nameplate with its actuation and reporting
// ports.
func New(cfg Config, np *Nameplate, invoker EffectorInvoker, report VerdictSink, runID string) *Runner {
	return &Runner{runner: layer.New(cfg, np, invoker, report, runID)}
}

// MCPOperator is the reference consumer's out-of-process operator client: it
// actuates and reports through the director process's operator endpoint and
// never sees a director tool, the ledger or truth. It is an EffectorInvoker,
// a VerdictSink and a QuiescenceReporter.
type MCPOperator struct {
	operator *mcpclient.MCPOperator
}

// NewMCPOperator connects to an operator endpoint with a capability token.
func NewMCPOperator(endpoint, token, runID string) (*MCPOperator, error) {
	inner, err := mcpclient.NewMCPOperator(endpoint, token, runID)
	if err != nil {
		return nil, err
	}
	return &MCPOperator{operator: inner}, nil
}
