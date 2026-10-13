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
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const pondDomain = "aquaculture-pond"

// suiteDomain is the shipped domain the suite command tests generate from:
// generating audits every candidate over a full simulated day, so they use the
// domain whose audit is cheapest. What the command does does not depend on it.
const suiteDomain = "shipment-exception"

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

func TestHelpPrintsUsageOnStderrAndNothingOnStdout(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		args   []string
		stderr string
	}{
		{"help", []string{"help"}, "Commands:"},
		{"short help flag", []string{"-h"}, "Commands:"},
		{"flag help", []string{"catalog", "-h"}, "-domains-dir"},
		{"verb flag help", []string{"domain", "validate", "-h"}, "Usage of domain"},
		{"subcommand help", []string{"device", "-h"}, "Usage: streamsim device serve"},
		{"subcommand flag help", []string{"device", "serve", "-h"}, "-capabilities"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invoke(t, tc.args...)
			if got.code != 0 || strings.Count(got.stderr, tc.stderr) != 1 || got.stdout != "" {
				t.Fatalf("exit %d stdout %q stderr %q; want exit 0, nothing on stdout, %q once on stderr", got.code, got.stdout, got.stderr, tc.stderr)
			}
		})
	}
}

func TestUsageFailuresExitTwoWithAUsageDocument(t *testing.T) {
	t.Parallel()
	domains, adapters := []string{"--domains-dir", testsupport.DomainsDir()}, []string{"--adapters-dir", testsupport.AdaptersDir()}
	cases := []struct {
		name   string
		args   []string
		stderr string
	}{
		{"no command", nil, "Usage: streamsim <command>"},
		{"unknown command", []string{"frobnicate"}, `unknown command "frobnicate"`},
		{"bad flag", []string{"catalog", "--no-such-flag"}, "flag provided but not defined"},
		{"catalog verb", append([]string{"catalog", "nope"}, domains...), `catalog: unknown verb "nope"`},
		{"describe without id", append([]string{"catalog", "describe"}, domains...), "catalog describe requires a domain id"},
		{"domain usage", []string{"domain"}, "usage: streamsim domain validate|format|lint <path>"},
		{"domain verb", []string{"domain", "bogus", "x"}, `domain: unknown verb "bogus"`},
		{"lint without path", []string{"domain", "lint"}, "usage: streamsim domain validate|format|lint <path>"},
		{"lint hours not positive", []string{"domain", "lint", "x", "--hours", "0"}, "domain lint: --hours must be positive and at most 720, got 0"},
		{"lint hours beyond the span", []string{"domain", "lint", "x", "--hours", "721"}, "at most 720"},
		{"lint flag on validate", []string{"domain", "validate", "x", "--seed", "3"}, "domain validate: --seed applies to lint only"},
		{"adapter verb", append([]string{"adapter", "nope"}, adapters...), `adapter: unknown verb "nope"`},
		{"adapter verify without path", append([]string{"adapter", "verify"}, adapters...), "adapter verify requires a path"},
		{"run without domain", []string{"run"}, "run requires --domain"},
		{"file sink without target", []string{"run", "--domain", pondDomain, "--sink", "file"}, "--sink=file requires --sink-target or --out"},
		{"unknown sink", []string{"run", "--domain", pondDomain, "--sink", "bogus"}, `--sink must be inproc, file or http-push, got "bogus"`},
		{"push sink without target", []string{"run", "--domain", pondDomain, "--sink", "http-push"}, "--sink=http-push requires --sink-target"},
		{"non-positive duration", []string{"run", "--domain", pondDomain, "--duration", "-5"}, "--duration must be positive"},
		{"duration beyond the run span", []string{"run", "--domain", pondDomain, "--duration", "1e12"}, "at most 2592000 seconds (30 days)"},
		{"duration not a number", []string{"run", "--domain", pondDomain, "--duration", "NaN"}, "--duration must be positive"},
		{"start beyond the world range", []string{"run", "--domain", pondDomain, "--start-time", "9223372036854775000"}, "--start-time must be in [0,"},
		{"fault offset not a number", []string{"run", "--domain", pondDomain, "--fault", "site-a/pond-1=aerator_failure@NaN"}, `--fault "site-a/pond-1=aerator_failure@NaN": offset "NaN" must be a number of seconds`},
		{"fault offset unreadable", []string{"run", "--domain", pondDomain, "--fault", "site-a/pond-1=aerator_failure@soon"}, `offset "soon" must be a number of seconds`},
		{"perturbation beyond the run span", []string{"run", "--domain", pondDomain, "--perturb", "drop@1e30"}, `--perturb "drop@1e30": offset "1e30"`},
		{"effector offset infinite", []string{"run", "--domain", pondDomain, "--effector", "start_aerator@site-a/pond-1@-Inf"}, `offset "-Inf" must be a number of seconds`},
		{"unknown run profile", append([]string{"run", "--domain", pondDomain, "--profile", "bogus"}, append(domains, adapters...)...), `has no profile "bogus"`},
		{"suite without domain", []string{"suite"}, "suite requires --domain"},
		{"suite of no scenarios", []string{"suite", "--domain", pondDomain, "--n", "0"}, "--n must be at least 1"},
		{"unknown suite profile", append([]string{"suite", "--domain", pondDomain, "--profile", "bogus"}, domains...), `has no profile "bogus"`},
		{"refconsumer empty window", []string{"refconsumer", "--trace", "t.jsonl", "--window", "0"}, "--window must be at least 1"},
		{"malformed fault", append([]string{"run", "--domain", pondDomain, "--duration", "1", "--fault", "no-separator"}, append(domains, adapters...)...), `bad triple "no-separator"`},
		{"replay without artifact", []string{"replay"}, "replay requires a run artifact path"},
		{"score without run", []string{"score"}, "score requires --run"},
		{"refconsumer without trace", []string{"refconsumer"}, "refconsumer requires --trace"},
		{"refconsumer effector without endpoint", []string{"refconsumer", "--trace", "t.jsonl", "--effector", "x"}, "--effector requires --mcp"},
		{"manifest without identity", []string{"manifest"}, "manifest requires --author and --reviewer"},
		{"device without subcommand", []string{"device"}, "device requires a subcommand: serve"},
		{"device unknown subcommand", []string{"device", "boot"}, `device: unknown subcommand "boot"`},
		{"device serve without socket", []string{"device", "serve"}, "device serve requires --socket"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invoke(t, tc.args...)
			if got.code != 2 || strings.Count(got.stderr, tc.stderr) != 1 {
				t.Fatalf("exit %d stderr %q; want exit 2 naming %q once", got.code, got.stderr, tc.stderr)
			}
			if failure := got.failure(t); failure["code"] != "usage_invalid" {
				t.Fatalf("failure = %v, want usage_invalid", failure)
			}
		})
	}
}

func TestMCPNeverPrintsADocumentOnTheStdoutItOwns(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{{"mcp"}, {"mcp", "--role", "operator"}} {
		got := invoke(t, args...)
		if got.code != 2 || got.stdout != "" || !strings.HasPrefix(got.stderr, "streamsim: ") {
			t.Fatalf("%v: exit %d stdout %q stderr %q; want exit 2, empty stdout, the reason on stderr", args, got.code, got.stdout, got.stderr)
		}
	}
}

func TestCatalogAndDomainCommandsPrintOneJSONDocument(t *testing.T) {
	t.Parallel()
	list := invoke(t, "catalog", "list", "--domains-dir", testsupport.DomainsDir()).mustSucceed(t).json(t)
	if domains, _ := list["domains"].([]any); len(domains) < 8 {
		t.Fatalf("catalog list = %v", list)
	}
	describe := invoke(t, "catalog", "describe", "--domains-dir", testsupport.DomainsDir(), pondDomain).mustSucceed(t).json(t)
	if describe["digest"] == "" || describe["spec"] == nil {
		t.Fatalf("catalog describe = %v", describe)
	}
	if coverage := invoke(t, "catalog", "coverage", "--domains-dir", testsupport.DomainsDir()).mustSucceed(t).json(t); len(coverage) == 0 {
		t.Fatal("catalog coverage printed nothing")
	}
	validate := invoke(t, "domain", "validate", filepath.Join(testsupport.DomainsDir(), pondDomain+".domain.json")).mustSucceed(t).json(t)
	if validate["valid"] != true || validate["id"] != pondDomain {
		t.Fatalf("domain validate = %v", validate)
	}
}

func TestAdapterListAndVerify(t *testing.T) {
	t.Parallel()
	list := invoke(t, "adapter", "list", "--adapters-dir", testsupport.AdaptersDir()).mustSucceed(t).json(t)
	if adapters, _ := list["adapters"].([]any); len(adapters) < 2 {
		t.Fatalf("adapter list = %v", list)
	}
	path := filepath.Join(testsupport.AdaptersDir(), "native-jsonl.adapter.json")
	verify := invoke(t, "adapter", "verify", "--adapters-dir", testsupport.AdaptersDir(), path).mustSucceed(t).json(t)
	if verify["schema_ok"] != true || verify["golden_match"] != true {
		t.Fatalf("adapter verify = %v", verify)
	}
}

func TestSuiteWritesTheGeneratedSuiteFile(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "suites")
	got := invoke(t, "suite", "--domains-dir", testsupport.DomainsDir(), "--domain", suiteDomain, "--n", "1", "--out", out).mustSucceed(t).json(t)
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
	got := invoke(t, "manifest", "--out", out, "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(),
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
	short := invoke(t, append(identity, "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(), "--key", key)...)
	if short.code != 1 || !strings.Contains(short.stderr, "manifest: key must be 64 hex chars") {
		t.Fatalf("short key: %+v", short)
	}
}

func TestFlagsMayPrecedeTheVerb(t *testing.T) {
	t.Parallel()
	path := testsupport.Adapter("native-jsonl")
	got := invoke(t, "adapter", "--adapters-dir", testsupport.AdaptersDir(), "verify", path).mustSucceed(t).document(t, "adapter_verify")
	if got["adapter"] != "native-jsonl" {
		t.Fatalf("adapter verify with leading flags = %v, want the native adapter verified", got)
	}
}
