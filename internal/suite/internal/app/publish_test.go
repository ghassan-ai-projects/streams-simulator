package app

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// publishDomain is the shipped domain these tests publish. Publishing audits
// every candidate over a full simulated day, so the test uses the shipped
// domain whose audit is cheapest; the rules under test (file naming,
// determinism, refusals) do not depend on which domain is published.
const publishDomain = "shipment-exception"

func cheapOptions(t *testing.T) PublishOptions {
	t.Helper()
	return PublishOptions{DomainsDir: testsupport.DomainsDir(), DomainID: publishDomain, Profile: "nominal",
		Out: filepath.Join(t.TempDir(), "suites"), N: 1, Seed: 1}
}

func TestPublishWritesTheSuiteFileNamedForItsDomainAndProfile(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "generates and audits a suite on one goroutine")
	options := cheapOptions(t)
	published, err := Publish(options)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(options.Out, publishDomain+"-nominal-suite.json"); published.Path != want {
		t.Fatalf("path = %q, want %q", published.Path, want)
	}
	raw, err := os.ReadFile(published.Path)
	if err != nil {
		t.Fatal(err)
	}
	var written rules.Suite
	if err := json.Unmarshal(raw, &written); err != nil || len(written.Scenarios) != published.Scenarios || written.TerminalState != published.TerminalState {
		t.Fatalf("file holds %d scenarios (%v), published says %+v", len(written.Scenarios), err, published)
	}
}

func TestPublishIsDeterministicForTheSameOptions(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "generates and audits a suite on one goroutine")
	first, err := Publish(cheapOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	second, err := Publish(cheapOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	a, _ := os.ReadFile(first.Path)
	b, _ := os.ReadFile(second.Path)
	if string(a) != string(b) {
		t.Fatal("the same options must publish the same suite bytes")
	}
}

func TestPublishRefusesAnUnknownProfileDomainAndDirectory(t *testing.T) {
	t.Parallel()
	options := cheapOptions(t)
	options.Profile = "no-such-profile"
	var unknown *rules.UnknownProfileError
	_, err := Publish(options)
	if !errors.As(err, &unknown) || err.Error() != `suite: domain "`+publishDomain+`" has no profile "no-such-profile"` {
		t.Fatalf("an unknown profile: err = %v", err)
	}
	if _, statErr := os.Stat(options.Out); !errors.Is(statErr, fs.ErrNotExist) {
		t.Fatalf("a refused suite must write nothing: %v", statErr)
	}
	options = cheapOptions(t)
	options.DomainID = "no-such"
	if _, err := Publish(options); !errors.Is(err, domain.ErrUnknownDomain) {
		t.Fatalf("an unknown domain: err = %v", err)
	}
	options = cheapOptions(t)
	options.DomainsDir = t.TempDir()
	if _, err := Publish(options); err == nil || !strings.Contains(err.Error(), "no domains found in") {
		t.Fatalf("an empty domains directory: err = %v", err)
	}
}

func TestPublishNamesAnOutputDirectoryItCannotCreate(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "generates and audits a suite on one goroutine")
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	options := cheapOptions(t)
	options.Out = filepath.Join(blocker, "suites")
	var pathFailure *fs.PathError
	if _, err := Publish(options); !errors.As(err, &pathFailure) || pathFailure.Path != blocker || !errors.Is(err, syscall.ENOTDIR) {
		t.Fatalf("err = %v, want a path failure naming %s", err, blocker)
	}
}
