package app

import (
	"cmp"
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func cmdCatalog(s *session, args []string) (any, error) {
	fs := newFlagSet("catalog", s.stderr)
	dir := fs.String("domains-dir", "", "directory of domain specs")
	verb, err := parseCommandVerb(fs, args, "list")
	if err != nil {
		return nil, err
	}
	cat, err := domain.NewCatalogFromDir(*dir)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return catalogVerb(cat, verb, fs.Args())
}

// parseCommandVerb reads a command's verb and flags, with the flags allowed
// before or after the verb; the arguments after them stay in fs.Args. With
// no verb it returns defaultVerb. A known verb given more arguments than it
// takes is a usage failure.
func parseCommandVerb(fs *flag.FlagSet, args []string, defaultVerb string) (string, error) {
	verb, err := readCommandVerb(fs, args, defaultVerb)
	if err != nil {
		return "", err
	}
	if arity, known := verbArity(verb); known {
		return verb, refuseExtraArguments(fs, arity)
	}
	return verb, nil
}

func readCommandVerb(fs *flag.FlagSet, args []string, defaultVerb string) (string, error) {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		return args[0], parseFlags(fs, args[1:])
	}
	if err := parseFlags(fs, args); err != nil {
		return "", err
	}
	if fs.NArg() == 0 {
		return defaultVerb, nil
	}
	return fs.Arg(0), keepPositional(fs, fs.Args()[1:])
}

// verbArity is how many arguments a verb takes, and whether it is a verb.
func verbArity(verb string) (int, bool) {
	switch verb {
	case "list", "coverage":
		return 0, true
	case "describe", "validate", "format", "verify", "from-run":
		return 1, true
	}
	return 0, false
}

func catalogVerb(cat *domain.Catalog, verb string, args []string) (any, error) {
	switch verb {
	case "list":
		return map[string]any{"domains": cat.List("")}, nil
	case "describe":
		return describeCatalogDomain(cat, args)
	case "coverage":
		return cat.Coverage(), nil
	}
	return nil, usagef("catalog: unknown verb %q", verb)
}

func describeCatalogDomain(cat *domain.Catalog, args []string) (any, error) {
	if len(args) == 0 {
		return nil, usagef("catalog describe requires a domain id")
	}
	compiled, err := cat.Describe(args[0])
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return map[string]any{"spec": compiled.Spec, "digest": compiled.Digest}, nil
}

func cmdDomain(s *session, args []string) (any, error) {
	fs := newFlagSet("domain", s.stderr)
	verb, err := parseCommandVerb(fs, args, "")
	if err != nil {
		return nil, err
	}
	return domainVerb(verb, fs.Args())
}

func domainVerb(verb string, args []string) (any, error) {
	switch {
	case verb == "validate" && len(args) > 0:
		return validateDomainFile(args[0])
	case verb == "format" && len(args) > 0:
		return formatDomainFile(args[0])
	case verb == "" || verb == "validate" || verb == "format":
		return nil, usagef("usage: streamsim domain validate|format <path>")
	}
	return nil, usagef("domain: unknown verb %q", verb)
}

func formatDomainFile(path string) (any, error) {
	result, err := domain.Format(path)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return result, nil
}

func validateDomainFile(path string) (any, error) {
	c, err := domain.Load(path)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return map[string]any{"valid": true, "id": c.Spec.ID, "digest": c.Digest, "channels": len(c.ChannelNames()), "faults": c.FaultCount(), "effectors": c.EffectorCount()}, nil
}

func cmdAdapter(s *session, args []string) (any, error) {
	fs := newFlagSet("adapter", s.stderr)
	dir := fs.String("adapters-dir", "", "directory of adapter files")
	verb, err := parseCommandVerb(fs, args, "list")
	if err != nil {
		return nil, err
	}
	return adapterVerb(verb, *dir, fs.Args())
}

func adapterVerb(verb, dir string, args []string) (any, error) {
	switch verb {
	case "list":
		return listAdapters(dir)
	case "verify":
		return verifyAdapterCommand(dir, args)
	}
	return nil, usagef("adapter: unknown verb %q", verb)
}

func listAdapters(dir string) (any, error) {
	adapters, err := adapter.LoadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return map[string]any{"adapters": adapterListings(adapters)}, nil
}

func adapterListings(adapters map[string]*model.Adapter) []map[string]any {
	var out []map[string]any
	for _, id := range sortedAdapterIDs(adapters) {
		a := adapters[id]
		out = append(out, map[string]any{"id": id, "version": a.Version, "encoding": a.Encoding, "title": a.Title})
	}
	return out
}

func sortedAdapterIDs(adapters map[string]*model.Adapter) []string {
	ids := make([]string, 0, len(adapters))
	// determinism-safe: collected ids are sorted below.
	for id := range adapters {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func verifyAdapterCommand(dir string, args []string) (any, error) {
	if len(args) < 1 {
		return nil, usagef("adapter verify requires a path")
	}
	return verifyAdapterFile(args[0], cmp.Or(dir, "adapters"))
}

func verifyAdapterFile(path, base string) (any, error) {
	res, err := adapter.Verify(path, "", base)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return adapterVerdict(res)
}

// adapterVerdict fails the command when the adapter facade finds the
// verification failed (a declared check failed, or none was declared).
func adapterVerdict(res *adapter.VerifyResult) (any, error) {
	report := adapterVerifyReport(res)
	message, diverged := adapter.VerifyFailure(res)
	if message == "" {
		return report, nil
	}
	if diverged {
		report["first_divergence"], report["detail"] = res.FirstDivergence, res.Detail
	}
	return nil, adapterVerifyFailure(report, message)
}

func adapterVerifyReport(res *adapter.VerifyResult) map[string]any {
	return map[string]any{"adapter": res.Adapter, "schema_ok": res.SchemaOK, "golden_match": res.GoldenMatch,
		"schema_checked": res.SchemaChecked, "golden_checked": res.GoldenChecked, "records": res.RecordCount}
}

func adapterVerifyFailure(report map[string]any, message string) error {
	return &outcomeError{code: codeAdapterVerifyFailed, message: message, result: report}
}
