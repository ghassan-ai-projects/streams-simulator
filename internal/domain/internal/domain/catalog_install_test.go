package domain

import (
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func loadShipped(t *testing.T, id string) *Compiled {
	t.Helper()
	compiled, err := Load(testsupport.Domain(id))
	if err != nil {
		t.Fatal(err)
	}
	return compiled
}

func TestCatalogWithAddsADomainWithoutChangingTheOriginal(t *testing.T) {
	t.Parallel()
	base := NewCatalog([]*Compiled{loadShipped(t, "rotating-machinery")})
	grown, err := CatalogWith(base, loadShipped(t, "aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	if got := len(base.List("")); got != 1 {
		t.Fatalf("the original catalog lists %d domains, want 1", got)
	}
	if got := grown.List(""); len(got) != 2 || got[0].ID != "aquaculture-pond" || got[1].ID != "rotating-machinery" {
		t.Fatalf("grown catalog = %+v, want both domains sorted by id", got)
	}
}

func TestCatalogWithTheSameDigestIsIdempotent(t *testing.T) {
	t.Parallel()
	pond := loadShipped(t, "aquaculture-pond")
	base := NewCatalog([]*Compiled{pond})
	same, err := CatalogWith(base, pond)
	if err != nil || same != base {
		t.Fatalf("With(same digest) = %p, %v, want the catalog itself and no error", same, err)
	}
}

func TestCatalogWithAnotherDigestUnderAnInstalledIDIsRefused(t *testing.T) {
	t.Parallel()
	pond := loadShipped(t, "aquaculture-pond")
	other := *pond
	other.Digest = "sha256:other"
	if _, err := CatalogWith(NewCatalog([]*Compiled{pond}), &other); !errors.Is(err, ErrDomainIDTaken) {
		t.Fatalf("error = %v, want ErrDomainIDTaken", err)
	}
}
