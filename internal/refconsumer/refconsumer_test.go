package refconsumer_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

type recordingSink struct{ verdicts []*model.Verdict }

func (s *recordingSink) SubmitVerdict(v *model.Verdict) error {
	s.verdicts = append(s.verdicts, v)
	return nil
}

type noInvoker struct{}

func (noInvoker) InvokeEffector(string, string, string, map[string]any, int64) (*world.InvokeResult, error) {
	return nil, nil
}

func TestDefaultConfigIsAUsableDetectorTuning(t *testing.T) {
	t.Parallel()
	cfg := refconsumer.DefaultConfig()
	if cfg.Threshold <= 0 || cfg.Window <= 0 || cfg.MinConsecutive <= 0 {
		t.Fatalf("default config = %+v", cfg)
	}
}

func TestProcessOfAnEmptyTraceReportsAVerdictToTheSink(t *testing.T) {
	t.Parallel()
	sink := &recordingSink{}
	np := &refconsumer.Nameplate{WorldID: "w-1"}
	runner := refconsumer.New(refconsumer.DefaultConfig(), np, noInvoker{}, sink, "r-1")
	verdict, err := runner.Process(nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if verdict == nil || len(verdict.Detections) != 0 {
		t.Fatalf("verdict = %+v", verdict)
	}
	if len(sink.verdicts) != 1 {
		t.Fatalf("the verdict must reach the sink exactly once, got %d", len(sink.verdicts))
	}
}

func TestNewMCPOperatorRefusesAnIncompleteEndpointDescription(t *testing.T) {
	t.Parallel()
	if _, err := refconsumer.NewMCPOperator("", "token", "run"); err == nil || !strings.Contains(err.Error(), "required together") {
		t.Fatalf("err = %v", err)
	}
}

func TestMCPOperatorWorksOverAStreamableHTTPOperatorEndpoint(t *testing.T) {
	t.Parallel()
	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "operator"}, nil)
	object := map[string]any{"type": "object"}
	tool := func(name string, result any) {
		mcpsdk.AddTool(server, &mcpsdk.Tool{Name: name, InputSchema: object},
			func(context.Context, *mcpsdk.CallToolRequest, map[string]any) (*mcpsdk.CallToolResult, any, error) {
				return nil, result, nil
			})
	}
	tool("sim.nameplate.read", map[string]any{"world_id": "w-9", "entities": []any{}, "channels": []any{}, "effectors": []any{}})
	tool("sim.effector.list", []any{map[string]any{"name": "act", "args_schema": map[string]any{}}})
	tool("sim.effector.invoke", map[string]any{"accepted": true, "command_id": "c-1"})
	tool("sim.consumer.report", map[string]any{"ok": true})
	endpoint := httptest.NewServer(mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server { return server }, nil))
	defer endpoint.Close()

	operator, err := refconsumer.NewMCPOperator(endpoint.URL, "tok", "r-1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := operator.Close(); err != nil {
			t.Error(err)
		}
	}()
	if np, err := operator.Nameplate(); err != nil || np.WorldID != "w-9" {
		t.Fatalf("nameplate = %+v (%v)", np, err)
	}
	if list, err := operator.ListEffectors(); err != nil || len(list) != 1 {
		t.Fatalf("effectors = %+v (%v)", list, err)
	}
	if res, err := operator.InvokeEffector("act", "e-1", "c-1", nil, 1); err != nil || !res.Accepted {
		t.Fatalf("invoke = %+v (%v)", res, err)
	}
	if err := operator.SubmitVerdict(&model.Verdict{}); err != nil {
		t.Fatal(err)
	}
	if err := operator.ReportQuiesced(5); err != nil {
		t.Fatal(err)
	}
}
