package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/suite"
	"os"
	"path/filepath"
)

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
