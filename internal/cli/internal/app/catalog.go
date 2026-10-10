package app

import (
	"flag"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

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
	names, err := files.FileNames(dir)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return decodeAdapterDirectory(dir, names)
}

func decodeAdapterDirectory(dir string, names []string) (map[string]*model.Adapter, error) {
	out, err := loadAdapterEntries(dir, names)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no adapters found in %s", dir)
	}
	return out, nil
}

func loadAdapterEntries(dir string, names []string) (map[string]*model.Adapter, error) {
	out := map[string]*model.Adapter{}
	for _, name := range names {
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		a, err := adapter.Load(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("streamsim: %w", err)
		}
		out[a.ID] = a
	}
	return out, nil
}

func cmdCatalog(s *session, args []string) (any, error) {
	fs := newFlagSet("catalog", s.stderr)
	dir := fs.String("domains-dir", "", "directory of domain specs")
	verb, err := parseCommandVerb(fs, args)
	if err != nil {
		return nil, err
	}
	cat, err := loadCatalog(*dir)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return catalogVerb(cat, verb, fs.Args())
}

func parseCommandVerb(fs *flag.FlagSet, args []string) (string, error) {
	verb := "list"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		verb, args = args[0], args[1:]
	}
	if err := parseFlags(fs, args); err != nil {
		return "", err
	}
	return verb, nil
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
	return nil, fmt.Errorf("catalog: unknown verb %q", verb)
}

func describeCatalogDomain(cat *domain.Catalog, args []string) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("catalog describe requires a domain id")
	}
	compiled, err := cat.Describe(args[0])
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return map[string]any{"spec": compiled.Spec, "digest": compiled.Digest}, nil
}

func cmdDomain(s *session, args []string) (any, error) {
	fs := newFlagSet("domain", s.stderr)
	if err := parseFlags(fs, args); err != nil {
		return nil, err
	}
	if fs.NArg() < 2 {
		return nil, fmt.Errorf("usage: streamsim domain validate <path>")
	}
	switch fs.Arg(0) {
	case "validate":
		return validateDomainFile(fs.Arg(1))
	}
	return nil, fmt.Errorf("domain: unknown verb %q", fs.Arg(0))
}

func validateDomainFile(path string) (any, error) {
	c, err := domain.Load(path)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return map[string]any{"valid": true, "id": c.Spec.ID, "digest": c.Digest, "channels": len(c.Spec.Channels), "faults": len(c.Spec.Faults), "effectors": len(c.Spec.Effectors)}, nil
}

func cmdAdapter(s *session, args []string) (any, error) {
	fs := newFlagSet("adapter", s.stderr)
	dir := fs.String("adapters-dir", "", "directory of adapter files")
	verb, err := parseCommandVerb(fs, args)
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
	return nil, fmt.Errorf("adapter: unknown verb %q", verb)
}

func listAdapters(dir string) (any, error) {
	adapters, err := loadAdapters(dir)
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
		return nil, fmt.Errorf("adapter verify requires a path")
	}
	if dir == "" {
		dir = "adapters"
	}
	return verifyAdapterFile(args[0], dir)
}

func verifyAdapterFile(path, base string) (any, error) {
	res, err := adapter.Verify(path, "", base)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return adapterVerdict(res)
}

// adapterVerdict fails the command when a declared check failed, or when the
// adapter declares none, so "verified" always means something was compared.
func adapterVerdict(res *adapter.VerifyResult) (any, error) {
	if !res.SchemaChecked && !res.GoldenChecked {
		return nil, fmt.Errorf("adapter verify: %s declares no conformance schema or golden to check", res.Adapter)
	}
	if (res.SchemaChecked && !res.SchemaOK) || (res.GoldenChecked && !res.GoldenMatch) {
		return nil, fmt.Errorf("adapter verify FAILED: %s %s", res.FirstDivergence, res.Detail)
	}
	return map[string]any{"adapter": res.Adapter, "schema_ok": res.SchemaOK, "golden_match": res.GoldenMatch,
		"schema_checked": res.SchemaChecked, "golden_checked": res.GoldenChecked, "records": res.RecordCount}, nil
}
