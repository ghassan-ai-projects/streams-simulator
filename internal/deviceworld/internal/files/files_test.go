package files

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestReadBindingsReturnsTheCatalogBytesAndNamesAMissingOne(t *testing.T) {
	t.Parallel()
	data, err := ReadBindings(testsupport.DeviceWorldBindings())
	if err != nil || len(data) == 0 {
		t.Fatalf("bindings = %d bytes, %v", len(data), err)
	}
	_, err = ReadBindings(filepath.Join(t.TempDir(), "absent.json"))
	if !errors.Is(err, fs.ErrNotExist) || !strings.HasPrefix(err.Error(), "read world bindings: ") {
		t.Fatalf("a missing catalog: err = %v", err)
	}
}
