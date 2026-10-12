package contract

import (
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestThePlantWireContractMatchesItsPinnedDigest(t *testing.T) {
	t.Parallel()
	pinned, err := os.ReadFile(filepath.Join(plantWireDir(), "DIGEST"))
	if err != nil {
		t.Fatal(err)
	}
	if got := contractDigest(t); strings.TrimSpace(string(pinned)) != got {
		t.Fatalf("schemas digest %s, pinned %s: a contract change needs a reviewed DIGEST change", got, strings.TrimSpace(string(pinned)))
	}
}

var safetyFloor = map[string][]string{
	"DCMD":     {"command_id", "idempotency_key", "device", "command", "args", "expected_boot_id", "deadline_ns", "not_before_ns", "authority_epoch", "policy_digest", "fallback_on_lease_loss"},
	"NBIRTH":   {"protocol_min", "protocol_max", "contract_version", "contract_digest", "environment", "bd_seq", "capability_digest", "kinds"},
	"DBIRTH":   {"device", "kind", "board", "boot_id", "birth_time_ns", "metrics"},
	"DDATA":    {"device", "boot_id", "historical", "metrics"},
	"STATE":    {"online", "heartbeat", "authority_epoch", "protocol_min", "protocol_max"},
	"DSTATE":   {"boot_id", "self_reported", "safe_state"},
	"DRECEIPT": {"command_id", "status", "boot_id"},
	"DRESULT":  {"command_id", "reported_outcome", "boot_id"},
}

var safetyValues = map[string]map[string][]string{
	"DRECEIPT": {"reason": {"expired", "wrong_boot", "duplicate", "stale_epoch", "policy_not_allowed", "interlocked", "out_of_range", "locked_out", "maintenance_isolated", "local_control", "permissive_not_met", "session_ending"}},
	"DRESULT":  {"reported_outcome": {"unknown"}},
	"DCMD":     {"fallback_on_lease_loss": {"safe_off"}},
	"NBIRTH":   {"environment": {"simulated", "hardware"}},
	"STATE":    {"refusal": {"environment_mismatch", "version_unsupported", "contract_unsupported"}},
}

func TestTheSafetyFieldsCannotBeWeakened(t *testing.T) {
	t.Parallel()
	for record, fields := range safetyFloor {
		schema := readJSON(t, filepath.Join(plantWireDir(), "schemas", record+".schema.json"))
		for _, field := range fields {
			if !slices.Contains(member(schema, "required").([]any), any(field)) {
				t.Errorf("%s no longer requires %s", record, field)
			}
		}
		for field, values := range safetyValues[record] {
			for _, value := range values {
				if !slices.Contains(member(schema, "properties", field, "enum").([]any), any(value)) {
					t.Errorf("%s.%s no longer offers %s", record, field, value)
				}
			}
		}
	}
	if member(readJSON(t, filepath.Join(plantWireDir(), "schemas", "DCMD.schema.json")), "additionalProperties") != false {
		t.Error("DCMD no longer refuses unknown members")
	}
}

func TestTheVendoredContractMatchesItsMasterWhenOneIsConfigured(t *testing.T) {
	t.Parallel()
	master := os.Getenv("PLANT_WIRE_CONTRACT_DIR")
	if master == "" {
		t.Skip("PLANT_WIRE_CONTRACT_DIR is not set; the drift check runs where a master is configured")
	}
	if local, upstream := folderContents(t, os.DirFS(plantWireDir())), folderContents(t, os.DirFS(master)); !maps.Equal(local, upstream) {
		t.Errorf("contracts/plant-wire differs from the master in %s: run make contract-sync", master)
	}
}

func folderContents(t *testing.T, folder fs.FS) map[string]string {
	t.Helper()
	contents := map[string]string{}
	err := fs.WalkDir(folder, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		raw, err := fs.ReadFile(folder, path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		contents[path] = string(raw)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return contents
}
