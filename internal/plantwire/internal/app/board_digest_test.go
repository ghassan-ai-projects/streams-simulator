package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const thermalBoardFile = "arduino-thermal.board.json"

func thermalKindsCopy(t *testing.T, edit func(board string) string) string {
	t.Helper()
	dir := t.TempDir()
	entries, err := os.ReadDir(testsupport.ThermalBoardKindsDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".kind.json") && name != thermalBoardFile {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(testsupport.ThermalBoardKindsDir(), name))
		if err != nil {
			t.Fatal(err)
		}
		if name == thermalBoardFile && edit != nil {
			raw = []byte(edit(string(raw)))
		}
		testsupport.WriteFileIn(t, dir, name, raw)
	}
	return dir
}

func checkWithKinds(t *testing.T, kindsDir string) (ConnectCheck, error) {
	t.Helper()
	options := connectOptions(t, planText("/tmp/h.sock", "/tmp/g.sock", "fan-01", "mega-bench", "fan-01"), "")
	options.KindsDirs = []string{kindsDir}
	return CheckConnect(options)
}

func boardDigestOf(check ConnectCheck, kind string) string {
	for _, entry := range check.ImplementsChecked {
		if entry.Kind == kind && entry.BoardType == "arduino-thermal" {
			return entry.BoardDigest
		}
	}
	return ""
}

func TestTheCheckNamesTheBoardMappingSoAChangedLeaseOrRowGivesAnotherDocument(t *testing.T) {
	t.Parallel()
	shipped, err := checkWithKinds(t, thermalKindsCopy(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	longer, err := checkWithKinds(t, thermalKindsCopy(t, func(board string) string { return strings.Replace(board, `"ms": 5000`, `"ms": 10000`, 1) }))
	if err != nil {
		t.Fatal(err)
	}
	digest := boardDigestOf(shipped, "fan")
	if !strings.HasPrefix(digest, "sha256:") || boardDigestOf(longer, "fan") == digest || boardDigestOf(shipped, "thermal_sensor") != "" {
		t.Fatalf("board digests %q and %q, want the mapping named only for mapped kinds and a changed lease to change it", digest, boardDigestOf(longer, "fan"))
	}
	for i := range shipped.ImplementsChecked {
		if shipped.ImplementsChecked[i].KindDigest != longer.ImplementsChecked[i].KindDigest {
			t.Fatal("a lease change in the board mapping changed a kind digest")
		}
	}
	for _, instance := range shipped.Instances {
		if instance.Instance == "fan-01" && (instance.BoardType != "arduino-thermal" || instance.BoardDigest != digest) {
			t.Errorf("real fan-01 = %+v, want the board mapping the gateway must hold", instance)
		}
	}
}

func TestAMappingThatReadsBackAnotherWordThanItWroteIsRefused(t *testing.T) {
	t.Parallel()
	swapped := func(board string) string {
		board = strings.Replace(board, `"off": {"brightness_permille": 0, "pattern": "off"}`, `"off": {"brightness_permille": 0, "pattern": "solid"}`, 1)
		return strings.Replace(board, `"alert": {"brightness_permille": 1000, "pattern": "solid"}`, `"alert": {"brightness_permille": 1000, "pattern": "off"}`, 1)
	}
	_, err := checkWithKinds(t, thermalKindsCopy(t, swapped))
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || !errors.Is(err, domain.ErrBoardRule) || invalid.Issues[0].Keyword != domain.KeywordBoardReadbackMismatch {
		t.Fatalf("check = %v, want the swapped LED rows refused board_readback_mismatch", err)
	}
}

func TestARunRecordsTheBoardMappingsItReadInItsArtifactAndSessionRecord(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-b")
	host := serveWireHost(t, hostPath, true)
	options := connectOptions(t, planText(hostPath, ""), out)
	check, err := CheckConnect(connectOptions(t, planText(hostPath, ""), ""))
	if err != nil {
		t.Fatal(err)
	}
	cancel, done := startConnect(t, options, false)
	host.waitFor(func(r []map[string]any) bool { return hasRecord(r, "DBIRTH", "device", "fan-01") })
	cancel()
	if ended := <-done; ended.err != nil {
		t.Fatal(ended.err)
	}
	raw, err := os.ReadFile(filepath.Join(out, model.FileRunArtifact))
	if err != nil {
		t.Fatal(err)
	}
	var artifact model.RunArtifact
	if err := json.Unmarshal(raw, &artifact); err != nil {
		t.Fatal(err)
	}
	if want := boardDigestOf(check, "fan"); want == "" || artifact.BoardDigests["arduino-thermal"] != want || len(artifact.BoardDigests) != 1 {
		t.Fatalf("run.json board_digests = %v, want arduino-thermal: %s", artifact.BoardDigests, want)
	}
	session := journalOfType(t, out, "session")
	if len(session) != 1 || !strings.Contains(string(mustJSON(t, session[0])), boardDigestOf(check, "fan")) {
		t.Fatalf("session record = %v, want the board digest among its implements_checked", session)
	}
	replaysWhole(t, out, options.DomainsDir)
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestAMappingThatLeavesARequiredRawArgumentUnsetIsRefusedAtLoad(t *testing.T) {
	t.Parallel()
	withoutLease := func(board string) string {
		return strings.Replace(board, "},\n          \"lease\": {\"arg\": \"lease_ms\", \"ms\": 5000}", "}", 1)
	}
	_, err := checkWithKinds(t, thermalKindsCopy(t, withoutLease))
	if !errors.Is(err, domain.ErrKindDrift) || !strings.Contains(err.Error(), `"lease_ms" is required`) {
		t.Fatalf("check = %v, want the fan refused at load because nothing sets lease_ms", err)
	}
}
