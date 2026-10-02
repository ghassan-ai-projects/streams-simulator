package cli

import (
	"flag"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
		ids := make([]string, 0, len(adapters))
		for id := range adapters {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		var out []map[string]any
		for _, id := range ids {
			a := adapters[id]
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
