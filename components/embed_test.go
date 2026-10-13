package components

import (
	"io/fs"
	"testing"
)

func TestTheEmbeddedLibraryHoldsTheVocabularyEquipmentInstrumentsAndCompositions(t *testing.T) {
	t.Parallel()
	for pattern, want := range map[string]int{"vocabulary.json": 1, "equipment/*.component.json": 20,
		"instruments/*.instrument.json": 17, "compositions/*.composition.json": 1} {
		matches, err := fs.Glob(FS(), pattern)
		if err != nil || len(matches) < want {
			t.Errorf("%s matches %v (%v), want at least %d files", pattern, matches, err, want)
		}
	}
}
