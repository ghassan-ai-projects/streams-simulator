// Package testsupport holds the fixtures and builders shared by tests: where
// the shipped domains and adapters live, and short unix-socket paths. Only
// _test.go files import it; a production package never does.
package testsupport

import (
	"os"
	"path/filepath"
	"testing"
)

func RepositoryRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		panic("testsupport: working directory: " + err.Error())
	}
	for !hasGoMod(dir) {
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("testsupport: no go.mod above the working directory")
		}
		dir = parent
	}
	return dir
}

func hasGoMod(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "go.mod"))
	return err == nil
}

func DomainsDir() string { return filepath.Join(RepositoryRoot(), "domains") }

func Domain(id string) string { return filepath.Join(DomainsDir(), id+".domain.json") }

func Example() string {
	return filepath.Join(RepositoryRoot(), "examples", "aquaculture-pond.domain.json")
}

func StructuredDomain() string {
	return filepath.Join(RepositoryRoot(), "internal", "testsupport", "testdata", "structured-plant.domain.json")
}

func TransferDomain() string {
	return filepath.Join(RepositoryRoot(), "internal", "testsupport", "testdata", "transfer-plant.domain.json")
}

func UtilityDomain() string {
	return filepath.Join(RepositoryRoot(), "internal", "testsupport", "testdata", "utility-plant.domain.json")
}

func TwoKindDomain() string {
	return filepath.Join(RepositoryRoot(), "internal", "testsupport", "testdata", "two-kind-plant.domain.json")
}

func Contract(stem string) string {
	return filepath.Join(RepositoryRoot(), "contracts", stem+".schema.json")
}

func AdaptersDir() string { return filepath.Join(RepositoryRoot(), "adapters") }

func Adapter(id string) string { return filepath.Join(AdaptersDir(), id+".adapter.json") }

func DeviceCapabilities() string {
	return filepath.Join(RepositoryRoot(), "internal", "device", "testdata", "thermal_capability_catalog.json")
}

func DeviceWorldBindings() string {
	return filepath.Join(RepositoryRoot(), "internal", "deviceworld", "testdata", "thermal.bindings.json")
}

func DeviceConformance(validity, name string) string {
	return filepath.Join(RepositoryRoot(), "internal", "device", "contract", "conformance", "v1", validity, name)
}

func SocketPath(tb testing.TB) string {
	tb.Helper()
	//nolint:usetesting // t.TempDir paths exceed the unix socket path limit on macOS
	dir, err := os.MkdirTemp("", "ss")
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			tb.Error(err)
		}
	})
	return filepath.Join(dir, "d.sock")
}

func ReportRun() string {
	return filepath.Join(RepositoryRoot(), "internal", "report", "testdata", "run")
}

func ReportGolden(name string) string {
	return filepath.Join(RepositoryRoot(), "internal", "report", "testdata", name)
}

func PumpBoardDomain() string {
	return filepath.Join(RepositoryRoot(), "internal", "testsupport", "testdata", "pump-board.domain.json")
}

func CoupledDomain() string {
	return filepath.Join(RepositoryRoot(), "internal", "testsupport", "testdata", "coupled-plant.domain.json")
}

func LoopDomain() string {
	return filepath.Join(RepositoryRoot(), "internal", "testsupport", "testdata", "loop-plant.domain.json")
}

func ComponentsDir() string {
	return filepath.Join(RepositoryRoot(), "internal", "component", "internal", "files", "builtin")
}

func PilotComposition() string {
	return filepath.Join(RepositoryRoot(), "examples", "compositions", "chiller-pilot.composition.json")
}

func FactoryComposition() string {
	return filepath.Join(RepositoryRoot(), "examples", "compositions", "factory-utilities.composition.json")
}
