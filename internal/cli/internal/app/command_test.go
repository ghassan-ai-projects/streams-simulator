package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/process"
)

const (
	domainsDir  = "../../../../domains"
	adaptersDir = "../../../../adapters"
	pondDomain  = "aquaculture-pond"
)

var fixedNow = time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)

type outcome struct {
	code           int
	stdout, stderr string
}

// invoke runs one streamsim command in-process against buffers.
func invoke(t *testing.T, args ...string) outcome {
	t.Helper()
	var stdout, stderr bytes.Buffer
	env := process.Env{Stdout: &stdout, Stderr: &stderr, Now: func() time.Time { return fixedNow }}
	code := Run(append([]string{"streamsim"}, args...), Build{Version: "v-test", Commit: "c-test"}, env)
	return outcome{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func (o outcome) json(t *testing.T) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal([]byte(o.stdout), &doc); err != nil {
		t.Fatalf("stdout is not one JSON object: %v\n%s", err, o.stdout)
	}
	return doc
}

func (o outcome) mustSucceed(t *testing.T) outcome {
	t.Helper()
	if o.code != 0 {
		t.Fatalf("exit %d\nstdout: %s\nstderr: %s", o.code, o.stdout, o.stderr)
	}
	return o
}

func TestUsageAndExitStatuses(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		args   []string
		code   int
		stderr string
	}{
		{"no command", nil, 2, "Usage: streamsim <command>"},
		{"help", []string{"help"}, 0, "Commands:"},
		{"short help flag", []string{"-h"}, 0, "Commands:"},
		{"unknown command", []string{"frobnicate"}, 2, `unknown command "frobnicate"`},
		{"bad flag", []string{"catalog", "--no-such-flag"}, 2, "flag provided but not defined"},
		{"flag help", []string{"catalog", "-h"}, 0, "-domains-dir"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invoke(t, tc.args...)
			if got.code != tc.code || !strings.Contains(got.stderr, tc.stderr) || got.stdout != "" {
				t.Fatalf("exit %d stdout %q stderr %q; want exit %d with %q", got.code, got.stdout, got.stderr, tc.code, tc.stderr)
			}
		})
	}
}

func TestCommandErrorsExitOneWithThePrefixedMessage(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"catalog verb", []string{"catalog", "nope", "--domains-dir", domainsDir}, `catalog: unknown verb "nope"`},
		{"describe without id", []string{"catalog", "describe", "--domains-dir", domainsDir}, "catalog describe requires a domain id"},
		{"describe unknown", []string{"catalog", "describe", "no-such", "--domains-dir", domainsDir}, "streamsim: "},
		{"domain usage", []string{"domain"}, "usage: streamsim domain validate <path>"},
		{"domain verb", []string{"domain", "lint", "x"}, `domain: unknown verb "lint"`},
		{"adapter verb", []string{"adapter", "nope", "--adapters-dir", adaptersDir}, `adapter: unknown verb "nope"`},
		{"adapter verify without path", []string{"adapter", "verify", "--adapters-dir", adaptersDir}, "adapter verify requires a path"},
		{"run without domain", []string{"run"}, "run requires --domain"},
		{"replay without artifact", []string{"replay"}, "replay requires a run artifact path"},
		{"score without run", []string{"score"}, "score requires --run"},
		{"refconsumer without trace", []string{"refconsumer"}, "refconsumer requires --trace"},
		{"manifest without identity", []string{"manifest"}, "manifest requires --author and --reviewer"},
		{"mcp without role", []string{"mcp"}, "mcp requires --role director|operator"},
		{"mcp operator role", []string{"mcp", "--role", "operator"}, "the operator role is served from a director process"},
		{"device without subcommand", []string{"device"}, "device requires a subcommand: serve"},
		{"device unknown subcommand", []string{"device", "boot"}, `device: unknown subcommand "boot"`},
		{"device serve without socket", []string{"device", "serve"}, "device serve requires --socket"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invoke(t, tc.args...)
			if got.code != 1 || !strings.Contains(got.stderr, "streamsim: ") || !strings.Contains(got.stderr, tc.want) || got.stdout != "" {
				t.Fatalf("exit %d stdout %q stderr %q; want exit 1 naming %q", got.code, got.stdout, got.stderr, tc.want)
			}
		})
	}
}

func TestCatalogAndDomainCommandsPrintOneJSONDocument(t *testing.T) {
	t.Parallel()
	list := invoke(t, "catalog", "list", "--domains-dir", domainsDir).mustSucceed(t).json(t)
	if domains, _ := list["domains"].([]any); len(domains) < 8 {
		t.Fatalf("catalog list = %v", list)
	}
	describe := invoke(t, "catalog", "describe", "--domains-dir", domainsDir, pondDomain).mustSucceed(t).json(t)
	if describe["digest"] == "" || describe["spec"] == nil {
		t.Fatalf("catalog describe = %v", describe)
	}
	if coverage := invoke(t, "catalog", "coverage", "--domains-dir", domainsDir).mustSucceed(t).json(t); len(coverage) == 0 {
		t.Fatal("catalog coverage printed nothing")
	}
	validate := invoke(t, "domain", "validate", filepath.Join(domainsDir, pondDomain+".domain.json")).mustSucceed(t).json(t)
	if validate["valid"] != true || validate["id"] != pondDomain {
		t.Fatalf("domain validate = %v", validate)
	}
}

func TestAdapterListAndVerify(t *testing.T) {
	t.Parallel()
	list := invoke(t, "adapter", "list", "--adapters-dir", adaptersDir).mustSucceed(t).json(t)
	if adapters, _ := list["adapters"].([]any); len(adapters) < 2 {
		t.Fatalf("adapter list = %v", list)
	}
	path := filepath.Join(adaptersDir, "native-jsonl.adapter.json")
	verify := invoke(t, "adapter", "verify", "--adapters-dir", adaptersDir, path).mustSucceed(t).json(t)
	if verify["schema_ok"] != true || verify["golden_match"] != true {
		t.Fatalf("adapter verify = %v", verify)
	}
}

func TestSuiteWritesTheGeneratedSuiteFile(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "suites")
	got := invoke(t, "suite", "--domains-dir", domainsDir, "--domain", pondDomain, "--n", "4", "--out", out).mustSucceed(t).json(t)
	path, _ := got["suite"].(string)
	if filepath.Dir(path) != out || got["scenarios"] == nil {
		t.Fatalf("suite result = %v", got)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("suite file missing: %v", err)
	}
}

func TestManifestRecordsBuildClockAndDigests(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "manifest.json")
	got := invoke(t, "manifest", "--out", out, "--domains-dir", domainsDir, "--adapters-dir", adaptersDir,
		"--author", "A <a@x>", "--reviewer", "R <r@x>").mustSucceed(t).json(t)
	if got["sim"] != "v-test@c-test" || got["signed"] != false {
		t.Fatalf("manifest result = %v", got)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"created_at":"2026-10-10T12:00:00Z"`, `"version":"v-test"`, `"commit":"c-test"`} {
		if !strings.Contains(strings.ReplaceAll(string(raw), `": "`, `":"`), want) {
			t.Fatalf("manifest lacks %s:\n%s", want, raw)
		}
	}
}

func TestManifestRefusesAMissingDirectoryAndAShortKey(t *testing.T) {
	t.Parallel()
	identity := []string{"manifest", "--author", "A", "--reviewer", "R", "--out", filepath.Join(t.TempDir(), "m.json")}
	missing := invoke(t, append(identity, "--domains-dir", filepath.Join(t.TempDir(), "absent"))...)
	if missing.code != 1 || !strings.Contains(missing.stderr, "manifest: domain dir:") {
		t.Fatalf("missing domains dir: %+v", missing)
	}
	key := filepath.Join(t.TempDir(), "key")
	if err := os.WriteFile(key, []byte("abcd"), 0o600); err != nil {
		t.Fatal(err)
	}
	short := invoke(t, append(identity, "--domains-dir", domainsDir, "--adapters-dir", adaptersDir, "--key", key)...)
	if short.code != 1 || !strings.Contains(short.stderr, "manifest: key must be 64 hex chars") {
		t.Fatalf("short key: %+v", short)
	}
}
