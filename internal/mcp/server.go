package mcp

// Wiring the director and operator tool handlers onto the official MCP Go
// SDK. One process, one endpoint, two handler sets constructed from
// different view structs; a session's role never changes, and the server
// advertises only that role's tools.

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Implementation identity advertised by the server.
var implementation = &mcp.Implementation{Name: "streamsim", Version: "0.1.0"}

// toolDef describes one tool for the SDK wiring.
type toolDef struct {
	name        string
	description string
	schema      json.RawMessage
	// handler receives the decoded arguments and returns a JSON value.
	handler func(ctx context.Context, args map[string]any) (any, error)
}

// addTool registers one tool with the SDK. The SDK validates every call
// against the tool's declared schema before the handler runs; handlers keep
// the cross-field and domain-dependent checks the static schema cannot
// express (e.g. effector arguments validated against the loaded domain).
func addTool(s *mcp.Server, def toolDef) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        def.name,
		Description: def.description,
		InputSchema: def.schema,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
		out, err := def.handler(ctx, args)
		return nil, out, err
	})
}

// NewDirectorServer builds the director-role server.
func NewDirectorServer(d *Director) *mcp.Server {
	s := mcp.NewServer(implementation, nil)

	addTool(s, toolDef{name: "sim.catalog.list", description: "List the installed domains with their property vectors and stresses.", schema: toolSchema("sim.catalog.list"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return map[string]any{"domains": d.Catalog.List(str(args, "group"))}, nil
	}})
	addTool(s, toolDef{name: "sim.catalog.describe", description: "Describe a domain: channels, faults, effectors, profiles, fidelity tiers.", schema: toolSchema("sim.catalog.describe"), handler: func(_ context.Context, args map[string]any) (any, error) {
		c, err := d.Catalog.Describe(str(args, "domain"))
		if err != nil {
			return nil, errTool(CodeDomainInvalid, "%v", err)
		}
		return map[string]any{"spec": c.Spec, "digest": c.Digest}, nil
	}})
	addTool(s, toolDef{name: "sim.catalog.coverage", description: "The axis-coverage matrix; which axes are thin.", schema: toolSchema("sim.catalog.coverage"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return d.Catalog.Coverage(), nil
	}})
	addTool(s, toolDef{name: "sim.adapter.list", description: "Installed output adapters.", schema: toolSchema("sim.adapter.list"), handler: func(_ context.Context, args map[string]any) (any, error) {
		ids := make([]string, 0, len(d.Adapters))
		// determinism-safe: collected here, sorted below before output.
		for id := range d.Adapters {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		var out []map[string]any
		for _, id := range ids {
			a := d.Adapters[id]
			out = append(out, map[string]any{"id": id, "version": a.Version, "encoding": a.Encoding, "title": a.Title})
		}
		return map[string]any{"adapters": out}, nil
	}})
	addTool(s, toolDef{name: "sim.world.create", description: "Create a world: domain, seed, sink, adapter, time mode.", schema: toolSchema("sim.world.create"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return d.CreateWorld(args)
	}})
	addTool(s, toolDef{name: "sim.world.describe", description: "Describe a world: config, digest, clock, emitted count.", schema: toolSchema("sim.world.describe"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return d.DescribeWorld(str(args, "world_id"))
	}})
	addTool(s, toolDef{name: "sim.world.destroy", description: "Destroy a world; final counts and run artifact.", schema: toolSchema("sim.world.destroy"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return d.DestroyWorld(str(args, "world_id"))
	}})
	addTool(s, toolDef{name: "sim.clock.advance", description: "Advance by a relative nanosecond delta or to an absolute epoch nanosecond; await_consumer blocks on quiescence.", schema: toolSchema("sim.clock.advance"), handler: func(ctx context.Context, args map[string]any) (any, error) {
		worldID := str(args, "world_id")
		toNS, hasTo := intArg(args, "to_ns")
		byNS, hasBy := intArg(args, "by_ns")
		if hasTo == hasBy {
			return nil, errTool(CodeInvalidArgs, "exactly one of by_ns or to_ns is required")
		}
		if hasBy {
			if byNS < 0 {
				return nil, errTool(CodeInvalidArgs, "by_ns must be non-negative")
			}
			w := d.World(worldID)
			if w == nil {
				return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
			}
			const maxInt64 = int64(1<<63 - 1)
			if byNS > maxInt64-w.Run.World.Clock() {
				return nil, errTool(CodeInvalidArgs, "by_ns overflows the world clock")
			}
			toNS = w.Run.World.Clock() + byNS
		}
		return d.Advance(ctx, worldID, toNS, boolArg(args, "await_consumer"))
	}})
	addTool(s, toolDef{name: "sim.clock.state", description: "The clock, next scheduled event, pending effects.", schema: toolSchema("sim.clock.state"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return d.ClockState(str(args, "world_id"))
	}})
	addTool(s, toolDef{name: "sim.fault.inject", description: "Inject a world fault into an entity.", schema: toolSchema("sim.fault.inject"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return d.InjectFault(str(args, "world_id"), str(args, "entity_id"), str(args, "fault"), num(args, "onset_ns", 0), mapArg(args, "params"))
	}})
	addTool(s, toolDef{name: "sim.fault.clear", description: "Clear a fault by id.", schema: toolSchema("sim.fault.clear"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return d.ClearFault(str(args, "world_id"), str(args, "fault_id"))
	}})
	addTool(s, toolDef{name: "sim.fault.list", description: "Active faults. Director only.", schema: toolSchema("sim.fault.list"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return d.ListFaults(str(args, "world_id"))
	}})
	addTool(s, toolDef{name: "sim.perturb.apply", description: "Apply a delivery perturbation.", schema: toolSchema("sim.perturb.apply"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return d.ApplyPerturb(str(args, "world_id"), str(args, "perturbation"), mapArg(args, "params"), num(args, "from_ns", 0), num(args, "until_ns", 0))
	}})
	addTool(s, toolDef{name: "sim.perturb.clear", description: "Clear a perturbation by id.", schema: toolSchema("sim.perturb.clear"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return d.ClearPerturb(str(args, "world_id"), str(args, "perturb_id"))
	}})
	addTool(s, toolDef{name: "sim.env.inject", description: "Inject an environment fault against a configured target.", schema: toolSchema("sim.env.inject"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return d.EnvInject(str(args, "world_id"), str(args, "target"), str(args, "fault"), mapArg(args, "params"))
	}})
	addTool(s, toolDef{name: "sim.run.begin", description: "Open a run; truth is sealed.", schema: toolSchema("sim.run.begin"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return d.BeginRun(str(args, "world_id"), str(args, "label"))
	}})
	addTool(s, toolDef{name: "sim.run.end", description: "Close the run; writes the run artifact.", schema: toolSchema("sim.run.end"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return d.EndRun(str(args, "world_id"))
	}})
	addTool(s, toolDef{name: "sim.truth.seal", description: "Install and seal a director-only ground-truth record before run.begin.", schema: toolSchema("sim.truth.seal"), handler: func(_ context.Context, args map[string]any) (any, error) {
		runID := str(args, "run_id")
		raw, ok := args["ground_truth"].(map[string]any)
		if !ok {
			return nil, errTool(CodeInvalidArgs, "ground_truth object is required")
		}
		b, err := json.Marshal(raw)
		if err != nil {
			return nil, errTool(CodeInvalidArgs, "%v", err)
		}
		rec := &model.GroundTruthRecord{}
		if err := json.Unmarshal(b, rec); err != nil {
			return nil, errTool(CodeInvalidArgs, "%v", err)
		}
		if err := d.SealTruth(runID, rec); err != nil {
			return nil, err
		}
		return map[string]any{"run_id": runID, "sealed": true}, nil
	}})
	addTool(s, toolDef{name: "sim.run.verify", description: "Verify a run artifact reproduces.", schema: toolSchema("sim.run.verify"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return d.VerifyRun(str(args, "run_artifact_path"))
	}})
	addTool(s, toolDef{name: "sim.truth.reveal", description: "Reveal sealed truth; unblind stamps the run permanently.", schema: toolSchema("sim.truth.reveal"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return d.RevealTruth(str(args, "run_id"), boolArg(args, "unblind"))
	}})
	addTool(s, toolDef{name: "sim.truth.seal_status", description: "Sealing state of a run.", schema: toolSchema("sim.truth.seal_status"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return d.SealStatus(str(args, "run_id"))
	}})
	addTool(s, toolDef{name: "sim.score", description: "Score a run from its submitted verdict, sealed truth, ledger and effector log.", schema: toolSchema("sim.score"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return d.Score(str(args, "run_id"))
	}})
	addTool(s, toolDef{name: "sim.scenario.audit", description: "Trivial-baseline audit of one injection.", schema: toolSchema("sim.scenario.audit"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return d.AuditScenario(str(args, "domain"), str(args, "entity_id"), str(args, "fault"), num(args, "onset_ns", 0), num(args, "start_ns", 0), num(args, "duration_ns", 0))
	}})
	addTool(s, toolDef{name: "sim.entity.retire", description: "Retire an entity.", schema: toolSchema("sim.entity.retire"), handler: func(_ context.Context, args map[string]any) (any, error) {
		w := d.World(str(args, "world_id"))
		if w == nil {
			return nil, errTool(CodeWorldNotFound, "unknown world")
		}
		if err := w.Run.RetireEntity(str(args, "entity_id"), str(args, "reason"), w.Run.World.Clock()); err != nil {
			return nil, errTool(CodeDomainInvalid, "%v", err)
		}
		return map[string]any{"retired": true}, nil
	}})

	// Resources: catalog and per-domain specs.
	s.AddResource(&mcp.Resource{URI: "sim://catalog", Name: "Domain catalog", MIMEType: "application/json"}, func(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		entries := d.Catalog.List("")
		b, _ := json.Marshal(entries)
		return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{URI: "sim://catalog", MIMEType: "application/json", Text: string(b)}}}, nil
	})
	s.AddResourceTemplate(&mcp.ResourceTemplate{URITemplate: "sim://domains/{id}/spec", Name: "Domain spec"}, func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		id, err := templateParam(req, "id")
		if err != nil {
			return nil, fmt.Errorf("mcp: %w", err)
		}
		c, err := d.Catalog.Describe(id)
		if err != nil {
			return nil, errTool(CodeDomainInvalid, "%v", err)
		}
		b, _ := json.Marshal(c.Spec)
		return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{URI: req.Params.URI, MIMEType: "application/json", Text: string(b)}}}, nil
	})
	return s
}

// NewOperatorServer builds the operator-role server from one world's view.
// It advertises exactly the four operator tools and nothing else.
func NewOperatorServer(v *OperatorView) *mcp.Server {
	s := mcp.NewServer(implementation, nil)
	addTool(s, toolDef{name: "sim.nameplate.read", description: "The static world nameplate: entities, channels, effectors.", schema: toolSchema("sim.nameplate.read"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return v.ReadNameplate(str(args, "token"))
	}})
	addTool(s, toolDef{name: "sim.effector.list", description: "The declared effectors and their argument schemas.", schema: toolSchema("sim.effector.list"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return v.ListEffectors(str(args, "token"))
	}})
	addTool(s, toolDef{name: "sim.effector.invoke", description: "Invoke an effector by name with a command_id (idempotency key) and capability token.", schema: toolSchema("sim.effector.invoke"), handler: func(_ context.Context, args map[string]any) (any, error) {
		return v.Invoke(str(args, "token"), str(args, "effector"), str(args, "entity_id"), str(args, "command_id"), mapArg(args, "args"), num(args, "at_ns", 0))
	}})
	addTool(s, toolDef{name: "sim.consumer.report", description: "Report quiescence and submit a consumer verdict. Write-only; never returns a score.", schema: toolSchema("sim.consumer.report"), handler: func(_ context.Context, args map[string]any) (any, error) {
		var verdict *model.Verdict
		if vd, ok := args["verdict"].(map[string]any); ok {
			b, err := json.Marshal(vd)
			if err != nil {
				return nil, errTool(CodeInvalidArgs, "%v", err)
			}
			verdict = &model.Verdict{}
			if err := json.Unmarshal(b, verdict); err != nil {
				return nil, errTool(CodeInvalidArgs, "%v", err)
			}
		}
		if err := v.Report(str(args, "token"), str(args, "run_id"), num(args, "quiesced_through_ns", 0), verdict); err != nil {
			return nil, fmt.Errorf("mcp: %w", err)
		}
		return map[string]any{"accepted": true, "world_id": v.WorldID, "simulated": true}, nil
	}})
	return s
}

func boolArg(args map[string]any, key string) bool {
	if args == nil {
		return false
	}
	b, _ := args[key].(bool)
	return b
}

func mapArg(args map[string]any, key string) map[string]any {
	if args == nil {
		return nil
	}
	if m, ok := args[key].(map[string]any); ok {
		return m
	}
	return nil
}

// templateParam extracts a value from a resolved resource URI template. The
// SDK resolves templates before invoking the handler, so the parameter is
// parsed from the concrete URI.
func templateParam(req *mcp.ReadResourceRequest, key string) (string, error) {
	uri := ""
	if req.Params != nil {
		uri = req.Params.URI
	}
	switch key {
	case "id":
		const prefix = "sim://domains/"
		if len(uri) > len(prefix) && uri[:len(prefix)] == prefix {
			id := uri[len(prefix):]
			if end := indexByte(id, '/'); end >= 0 {
				id = id[:end]
			}
			if id != "" {
				return id, nil
			}
		}
		return "", errTool(CodeDomainInvalid, "cannot parse domain id from %q", uri)
	}
	return "", errTool(CodeDomainInvalid, "unknown template parameter %q", key)
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}
