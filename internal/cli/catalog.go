package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
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
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return decodeAdapterDirectory(dir, entries)
}

func decodeAdapterDirectory(dir string, entries []os.DirEntry) (map[string]*model.Adapter, error) {
	out, err := loadAdapterEntries(dir, entries)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no adapters found in %s", dir)
	}
	return out, nil
}

func loadAdapterEntries(dir string, entries []os.DirEntry) (map[string]*model.Adapter, error) {
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
	return out, nil
}

func cmdCatalog(args []string) error {
	fs := flag.NewFlagSet("catalog", flag.ExitOnError)
	dir := fs.String("domains-dir", "", "directory of domain specs")
	verb, err := parseCommandVerb(fs, args)
	if err != nil {
		return err
	}
	cat, err := loadCatalog(*dir)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return catalogVerb(cat, verb, fs.Args())
}

func parseCommandVerb(fs *flag.FlagSet, args []string) (string, error) {
	verb := "list"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		verb, args = args[0], args[1:]
	}
	if err := fs.Parse(args); err != nil {
		return "", fmt.Errorf("streamsim: %w", err)
	}
	return verb, nil
}

func catalogVerb(cat *domain.Catalog, verb string, args []string) error {
	switch verb {
	case "list":
		return printJSON(map[string]any{"domains": cat.List("")})
	case "describe":
		return describeCatalogDomain(cat, args)
	case "coverage":
		return printJSON(cat.Coverage())
	}
	return fmt.Errorf("catalog: unknown verb %q", verb)
}

func describeCatalogDomain(cat *domain.Catalog, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("catalog describe requires a domain id")
	}
	compiled, err := cat.Describe(args[0])
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return printJSON(map[string]any{"spec": compiled.Spec, "digest": compiled.Digest})
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
		return validateDomainFile(fs.Arg(1))
	}
	return fmt.Errorf("domain: unknown verb %q", fs.Arg(0))
}

func validateDomainFile(path string) error {
	c, err := domain.Load(path)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return printJSON(map[string]any{"valid": true, "id": c.Spec.ID, "digest": c.Digest, "channels": len(c.Spec.Channels), "faults": len(c.Spec.Faults), "effectors": len(c.Spec.Effectors)})
}

func cmdAdapter(args []string) error {
	fs := flag.NewFlagSet("adapter", flag.ExitOnError)
	dir := fs.String("adapters-dir", "", "directory of adapter files")
	verb, err := parseCommandVerb(fs, args)
	if err != nil {
		return err
	}
	return adapterVerb(verb, *dir, fs.Args())
}

func adapterVerb(verb, dir string, args []string) error {
	switch verb {
	case "list":
		return listAdapters(dir)
	case "verify":
		return verifyAdapterCommand(dir, args)
	}
	return fmt.Errorf("adapter: unknown verb %q", verb)
}

func listAdapters(dir string) error {
	adapters, err := loadAdapters(dir)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return printJSON(map[string]any{"adapters": adapterListings(adapters)})
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
	for id := range adapters {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func verifyAdapterCommand(dir string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("adapter verify requires a path")
	}
	if dir == "" {
		dir = "adapters"
	}
	return verifyAdapterFile(args[0], dir)
}

func verifyAdapterFile(path, base string) error {
	res, err := adapter.Verify(path, "", base)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	if !res.SchemaOK || !res.GoldenMatch {
		return fmt.Errorf("adapter verify FAILED: %s", res.FirstDivergence)
	}
	return printJSON(map[string]any{"adapter": res.Adapter, "schema_ok": true, "golden_match": true, "records": res.RecordCount})
}
