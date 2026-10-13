package refconsumer

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer/internal/mcpclient"
)

type Runner struct {
	runner *layer.Runner
}

func New(cfg Config, np *Nameplate, invoker EffectorInvoker, report VerdictSink, runID string) *Runner {
	return &Runner{runner: layer.New(cfg, np, invoker, report, runID)}
}

type MCPOperator struct {
	operator *mcpclient.MCPOperator
}

func NewMCPOperator(endpoint, token, runID string) (*MCPOperator, error) {
	inner, err := mcpclient.NewMCPOperator(endpoint, token, runID)
	if err != nil {
		return nil, err
	}
	return &MCPOperator{operator: inner}, nil
}
