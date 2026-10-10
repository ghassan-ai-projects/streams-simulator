package domain

import layer "github.com/ghassan-ai-projects/streams-simulator/internal/domain/internal/domain"

// Catalog is the installed set of compiled domains, listed and described by
// id and reported on by axis coverage.
type Catalog struct {
	catalog *layer.Catalog
}

// NewCatalog builds a catalog over compiled domains.
func NewCatalog(domains []*Compiled) *Catalog {
	return &Catalog{catalog: layer.NewCatalog(domains)}
}
