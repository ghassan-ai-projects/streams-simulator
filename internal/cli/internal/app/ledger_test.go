package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadLedgerReadsJSONL(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "ledger.jsonl")
	raw := []byte(`{"delivery_id":1,"seq":1,"delivered":true,"delivery_reason":"ok"}
{"delivery_id":2,"seq":2,"delivered":false,"delivery_reason":"dropped_by_perturbation"}
`)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	ledger, err := loadLedger(path)
	if err != nil {
		t.Fatalf("load ledger: %v", err)
	}
	if len(ledger) != 2 {
		t.Fatalf("got %d ledger rows, want 2", len(ledger))
	}
	if ledger[0].DeliveryID != 1 || ledger[1].DeliveryID != 2 {
		t.Fatalf("unexpected delivery IDs: %+v", ledger)
	}
}

func TestLoadLedgerRejectsMalformedJSONL(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "ledger.jsonl")
	if err := os.WriteFile(path, []byte(`{"delivery_id":1}
not-json
`), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := loadLedger(path)
	if err == nil || !strings.HasPrefix(err.Error(), "decode ledger "+path+": ") {
		t.Fatalf("malformed JSONL: err = %v", err)
	}
}

func TestLoadLedgerNamesAPathThatIsADirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, err := loadLedger(dir)
	if err == nil || !strings.HasPrefix(err.Error(), "open ledger "+dir+": ") || !strings.Contains(err.Error(), "is a directory") {
		t.Fatalf("err = %v", err)
	}
}

func TestLoadLedgerNamesAMissingFile(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "absent.jsonl")
	_, err := loadLedger(path)
	if err == nil || !strings.HasPrefix(err.Error(), "open ledger "+path+": ") {
		t.Fatalf("err = %v", err)
	}
}
