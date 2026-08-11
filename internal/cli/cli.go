// Package cli implements the streamsim command-line interface: catalog,
// domain/adapter tooling, scripted runs, replay/verify, the MCP stdio
// servers, suite generation, and offline scoring. The entrypoint in
// cmd/streamsim is a thin dispatch over this package.
package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
	"github.com/ghassan-ai-projects/streams-simulator/internal/suite"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// Main dispatches to a subcommand. It is the whole CLI surface.
func Main(args []string) int {
	if len(args) < 2 {
		usage()
		return 2
	}
	cmd, rest := args[1], args[2:]
	var err error
	switch cmd {
	case "catalog":
		err = cmdCatalog(rest)
	case "domain":
		err = cmdDomain(rest)
	case "adapter":
		err = cmdAdapter(rest)
	case "run":
		err = cmdRun(rest)
	case "replay":
		err = cmdReplay(rest, false)
	case "verify":
		err = cmdReplay(rest, true)
	case "mcp":
		err = cmdMCP(rest)
	case "refconsumer":
		err = cmdRefconsumer(rest)
	case "suite":
		err = cmdSuite(rest)
	case "score":
		err = cmdScore(rest)
	case "help", "-h", "--help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "streamsim: unknown command %q\n\n", cmd)
		usage()
		return 2
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "streamsim: %v\n", err)
		return 1
	}
	return 0
}

func usage() {
	fmt.Fprintf(os.Stderr, `streamsim — deterministic world simulator for testing stream processors

Usage: streamsim <command> [flags]

Commands:
  catalog list|describe|coverage        Installed domains and axis coverage
  domain validate <path>                Validate a domain spec against the contract
  adapter list|verify [path]            Installed adapters; verify one against its golden
  run                                   One-shot scripted run
  replay <run.json>                     Reproduce a run from its artifact
  verify <run.json>                     Verify a run artifact reproduces
  mcp --role director|operator          Serve the MCP role over stdio
  refconsumer --trace <file>            Detect episodes in a trace; write verdict.json
  suite generate --domain --profile     Generate an audited graded suite
  score --run --verdict --label         Offline scorecard from artifacts
  help
`)
}

// loadCatalog loads every domain in a directory.
func loadCatalog(dir string) (*domain.Catalog, error) {
	if dir == "" {
		dir = "domains"
	}
	specs, err := domain.LoadAll(dir)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	if len(specs) == 0 {
		return nil, fmt.Errorf("no domains found in %s", dir)
	}
	return domain.NewCatalog(specs), nil
}

// loadAdapters loads every adapter in a directory.
func loadAdapters(dir string) (map[string]*model.Adapter, error) {
	if dir == "" {
		dir = "adapters"
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	out := map[string]*model.Adapter{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		a, err := adapter.Load(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("streamsim: %w", err)
		}
		out[a.ID] = a
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no adapters found in %s", dir)
	}
	return out, nil
}

func cmdCatalog(args []string) error {
	fs := flag.NewFlagSet("catalog", flag.ExitOnError)
	dir := fs.String("domains-dir", "", "directory of domain specs")
	verb := "list"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		verb = args[0]
		args = args[1:]
	}
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	cat, err := loadCatalog(*dir)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	switch verb {
	case "list":
		return printJSON(map[string]any{"domains": cat.List("")})
	case "describe":
		if fs.NArg() == 0 {
			return fmt.Errorf("catalog describe requires a domain id")
		}
		c, err := cat.Describe(fs.Arg(0))
		if err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		return printJSON(map[string]any{"spec": c.Spec, "digest": c.Digest})
	case "coverage":
		return printJSON(cat.Coverage())
	}
	return fmt.Errorf("catalog: unknown verb %q", verb)
}

func cmdDomain(args []string) error {
	fs := flag.NewFlagSet("domain", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	if fs.NArg() < 2 {
		return fmt.Errorf("usage: streamsim domain validate <path>")
	}
	switch fs.Arg(0) {
	case "validate":
		c, err := domain.Load(fs.Arg(1))
		if err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		return printJSON(map[string]any{"valid": true, "id": c.Spec.ID, "digest": c.Digest, "channels": len(c.Spec.Channels), "faults": len(c.Spec.Faults), "effectors": len(c.Spec.Effectors)})
	}
	return fmt.Errorf("domain: unknown verb %q", fs.Arg(0))
}

func cmdAdapter(args []string) error {
	fs := flag.NewFlagSet("adapter", flag.ExitOnError)
	dir := fs.String("adapters-dir", "", "directory of adapter files")
	verb := "list"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		verb = args[0]
		args = args[1:]
	}
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	switch verb {
	case "list":
		adapters, err := loadAdapters(*dir)
		if err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		var out []map[string]any
		for id, a := range adapters {
			out = append(out, map[string]any{"id": id, "version": a.Version, "encoding": a.Encoding, "title": a.Title})
		}
		return printJSON(map[string]any{"adapters": out})
	case "verify":
		if fs.NArg() < 1 {
			return fmt.Errorf("adapter verify requires a path")
		}
		base := *dir
		if base == "" {
			base = "adapters"
		}
		res, err := adapter.Verify(fs.Arg(0), "", base)
		if err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		if !res.SchemaOK || !res.GoldenMatch {
			return fmt.Errorf("adapter verify FAILED: %s", res.FirstDivergence)
		}
		return printJSON(map[string]any{"adapter": res.Adapter, "schema_ok": true, "golden_match": true, "records": res.RecordCount})
	}
	return fmt.Errorf("adapter: unknown verb %q", verb)
}

func cmdRun(args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	domainsDir := fs.String("domains-dir", "domains", "domain specs directory")
	adaptersDir := fs.String("adapters-dir", "adapters", "adapters directory")
	domainID := fs.String("domain", "", "domain id")
	adapterID := fs.String("adapter", "native-jsonl", "adapter id")
	seed := fs.Uint64("seed", 1, "world seed")
	sinkName := fs.String("sink", "file", "sink: inproc|file")
	sinkTarget := fs.String("sink-target", "", "sink target (file path)")
	outDir := fs.String("out", "", "output directory for artifacts")
	durationS := fs.Float64("duration", 6*3600, "run duration (seconds)")
	startTime := fs.Int64("start-time", model.DefaultStartTimeNS, "world start (ns epoch)")
	faults := fs.String("fault", "", "repeatable: entity=fault@offset_s")
	perts := fs.String("perturb", "", "repeatable: name@from_s[-until_s]")
	effectors := fs.String("effector", "", "repeatable: effector@entity@offset_s[=arg:val]")
	profile := fs.String("profile", "", "scenario profile")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	if *domainID == "" {
		return fmt.Errorf("run requires --domain")
	}
	cat, err := loadCatalog(*domainsDir)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	adapters, err := loadAdapters(*adaptersDir)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	spec, err := cat.Describe(*domainID)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	adap, ok := adapters[*adapterID]
	if !ok {
		return fmt.Errorf("unknown adapter %q", *adapterID)
	}
	cfg := run.Config{
		Domain: spec, Adapter: adap, Seed: *seed, SinkName: *sinkName,
		SinkTarget: *sinkTarget, TimeMode: model.TimeStepped, StartTimeNS: *startTime,
		ScenarioProfile: *profile,
	}
	r, err := run.New(context.Background(), cfg)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	start := *startTime
	// Setup first, then faults/perturbations at their offsets, then run.
	if *faults != "" {
		for _, f := range splitCSV(*faults) {
			entity, fault, offsetS, err := parseTriple(f, "=", "@")
			if err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
			if _, err := r.InjectFault(entity, fault, start+int64(offsetS*1e9), nil); err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
		}
	}
	if *perts != "" {
		for _, p := range splitCSV(*perts) {
			parts := strings.SplitN(p, "@", 3)
			name := parts[0]
			fromS := 0.0
			untilS := 0.0
			if len(parts) > 1 {
				fromS = parseF(parts[1])
			}
			if len(parts) > 2 {
				untilS = parseF(parts[2])
			}
			from := start + int64(fromS*1e9)
			until := int64(0)
			if untilS > 0 {
				until = start + int64(untilS*1e9)
			}
			if _, err := r.ApplyPerturb(name, nil, from, until); err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
		}
	}
	if *effectors != "" {
		for _, e := range splitCSV(*effectors) {
			eff, entity, offsetS, err := parseTriple(e, "@", "@")
			if err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
			cmdID := fmt.Sprintf("cli-%d", timeNanos())
			if _, err := r.InvokeEffector(eff, entity, cmdID, map[string]any{}, start+int64(offsetS*1e9)); err != nil {
				return fmt.Errorf("effector %s: %w", eff, err)
			}
		}
	}
	if _, err := r.Advance(start+int64(*durationS*1e9), false); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	art, err := r.End(*outDir)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return printJSON(map[string]any{
		"run_id": art.RunID, "trace_digest": art.ExpectedTraceDigest,
		"reproducible": art.Reproducible, "emitted": art.Counts.Emitted,
		"ledger": len(r.Ledger()),
	})
}

func cmdReplay(args []string, verifyOnly bool) error {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	domainsDir := fs.String("domains-dir", "domains", "domain specs directory")
	adaptersDir := fs.String("adapters-dir", "adapters", "adapters directory")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("replay requires a run artifact path")
	}
	art, err := run.LoadArtifact(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	cat, err := loadCatalog(*domainsDir)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	adapters, err := loadAdapters(*adaptersDir)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	spec, err := cat.Describe(art.Domain.ID)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	adap, ok := adapters[art.Adapter.ID]
	if !ok {
		return fmt.Errorf("unknown adapter %q", art.Adapter.ID)
	}
	res, err := run.ReplayArtifact(context.Background(), art, spec, adap, "")
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return printJSON(map[string]any{
		"matches": res.Matches, "version_match": res.VersionMatch,
		"got": res.GotDigest, "want": res.WantDigest,
		"first_divergence": res.FirstDivergence, "detail": res.Detail,
	})
}

func cmdMCP(args []string) error {
	fs := flag.NewFlagSet("mcp", flag.ExitOnError)
	role := fs.String("role", "", "director|operator")
	fs.String("world", "", "world id (operator role)")
	domainsDir := fs.String("domains-dir", "domains", "domain specs directory")
	adaptersDir := fs.String("adapters-dir", "adapters", "adapters directory")
	outDir := fs.String("out", "runs", "run artifact output directory")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	switch *role {
	case "director":
		cat, err := loadCatalog(*domainsDir)
		if err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		adapters, err := loadAdapters(*adaptersDir)
		if err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		d := mcp.NewDirector(context.Background(), cat, adapters, *outDir)
		srv := mcp.NewDirectorServer(d)
		if err := srv.Run(context.Background(), &mcpsdk.StdioTransport{}); err != nil {
			return fmt.Errorf("mcp director: %w", err)
		}
		return nil
	case "operator":
		return fmt.Errorf("the operator role is served per-world from a director session; use the director's sim.world.create token")
	default:
		return fmt.Errorf("mcp requires --role director|operator")
	}
}

func cmdRefconsumer(args []string) error {
	fs := flag.NewFlagSet("refconsumer", flag.ExitOnError)
	trace := fs.String("trace", "", "native-format trace file (JSONL)")
	out := fs.String("out", "verdict.json", "verdict output path")
	threshold := fs.Float64("threshold", 4, "z-score threshold")
	window := fs.Int("window", 30, "detector window")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	if *trace == "" {
		return fmt.Errorf("refconsumer requires --trace")
	}
	// #nosec G703 -- a CLI flag naming a trace file is user intent, not attacker input.
	raw, err := os.ReadFile(*trace)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	np := &refconsumer.Nameplate{WorldID: "cli"}
	verdictSink := &fileSink{}
	r := refconsumer.New(refconsumer.Config{Threshold: *threshold, Window: *window, MinConsecutive: 3, AbsenceFactor: 3}, np, nil, verdictSink, "cli")
	v, err := r.Process(raw, 0)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	rawV, _ := json.MarshalIndent(v, "", "  ")
	// #nosec G703 -- the CLI output path is user intent.
	if err := os.WriteFile(*out, rawV, 0o600); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return printJSON(map[string]any{"verdict_written": *out, "detections": len(v.Detections)})
}

type fileSink struct{}

func (f *fileSink) SubmitVerdict(v *model.Verdict) error { return nil }

func cmdSuite(args []string) error {
	fs := flag.NewFlagSet("suite", flag.ExitOnError)
	domainsDir := fs.String("domains-dir", "domains", "domain specs directory")
	domainID := fs.String("domain", "", "domain id")
	profile := fs.String("profile", "nominal", "profile name")
	n := fs.Int("n", 100, "target scenario count")
	seed := fs.Uint64("seed", 1, "suite seed")
	out := fs.String("out", "suites", "output directory")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	cat, err := loadCatalog(*domainsDir)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	spec, err := cat.Describe(*domainID)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	s, err := suite.Generate(suite.Config{Domain: spec, Profile: *profile, N: *n, Seed: *seed, SampleNS: 120 * 1e9})
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	if err := os.MkdirAll(*out, 0o700); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	raw, _ := json.MarshalIndent(s, "", "  ")
	path := filepath.Join(*out, fmt.Sprintf("%s-%s-suite.json", spec.Spec.ID, *profile))
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return printJSON(map[string]any{
		"suite": path, "scenarios": len(s.Scenarios), "trivial_excluded": len(s.TrivialExcluded),
		"terminal_state": s.TerminalState,
	})
}

func cmdScore(args []string) error {
	fs := flag.NewFlagSet("score", flag.ExitOnError)
	runArtifact := fs.String("run", "", "run artifact (run.json)")
	labelPath := fs.String("label", "", "ground truth label (JSON)")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	if *runArtifact == "" {
		return fmt.Errorf("score requires --run")
	}
	dir := filepath.Dir(*runArtifact)
	verdictPath := filepath.Join(dir, "verdict.json")
	label := *labelPath
	if label == "" {
		label = filepath.Join(dir, "label.json")
	}
	var verdict model.Verdict
	if err := loadJSON(verdictPath, &verdict); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	var gt model.GroundTruthRecord
	if err := loadJSON(label, &gt); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	var ledger []model.LedgerRecord
	if err := loadJSON(filepath.Join(dir, "ledger.jsonl"), &ledger); err != nil {
		ledger = nil
	}
	var calls []world.EffectorCall
	sc := score.Offline(&verdict, &gt, ledger, calls)
	return printJSON(sc)
}

func loadJSON(path string, dst any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func printJSON(v any) error {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	fmt.Println(string(raw))
	return nil
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// parseTriple splits "a=b@c" by two separators into three parts.
func parseTriple(s, sep1, sep2 string) (string, string, float64, error) {
	rest := s
	a := ""
	if i := strings.Index(rest, sep1); i >= 0 {
		a = rest[:i]
		rest = rest[i+len(sep1):]
	} else {
		return "", "", 0, fmt.Errorf("bad triple %q", s)
	}
	b := ""
	offset := 0.0
	if i := strings.Index(rest, sep2); i >= 0 {
		b = rest[:i]
		offset = parseF(rest[i+len(sep2):])
	} else {
		b = rest
	}
	return a, b, offset, nil
}

func parseF(s string) float64 {
	var f float64
	_, _ = fmt.Sscanf(s, "%f", &f)
	return f
}

func timeNanos() int64 { return time.Now().UnixNano() }
