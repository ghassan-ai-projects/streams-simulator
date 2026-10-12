package domain

import layer "github.com/ghassan-ai-projects/streams-simulator/internal/domain/internal/domain"

type Catalog struct {
	catalog *layer.Catalog
}

func NewCatalog(domains []*Compiled) *Catalog {
	return &Catalog{catalog: layer.NewCatalog(domains)}
}
