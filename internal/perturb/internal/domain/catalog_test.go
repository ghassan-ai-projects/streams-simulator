package domain

import "testing"

func TestCatalogNamesAreUniqueAndACopy(t *testing.T) {
	t.Parallel()
	names := CatalogNames()
	seen := map[string]bool{}
	for _, name := range names {
		if seen[name] {
			t.Fatalf("%s listed twice", name)
		}
		seen[name] = true
		if !isName(name) {
			t.Fatalf("%s is cataloged but not admitted", name)
		}
	}
	names[0] = "bogus"
	if isName("bogus") || CatalogNames()[0] == "bogus" {
		t.Fatal("mutating the returned slice changed the catalog")
	}
}
