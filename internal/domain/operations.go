package domain

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/domain/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func Load(path string) (*Compiled, error) {
	return files.Load(path)
}

func LoadAll(dir string) ([]*Compiled, error) {
	return files.LoadAll(dir)
}

func NewCatalogFromDir(dir string) (*Catalog, error) {
	inner, err := files.LoadCatalog(dir)
	if err != nil {
		return nil, err
	}
	return &Catalog{catalog: inner}, nil
}

func Format(path string) (*FormatResult, error) {
	return files.Format(path)
}

func Parse(raw []byte, src string) (*Compiled, error) {
	return layer.Parse(raw, src)
}

func (c *Catalog) List(group string) []Entry {
	if c == nil || c.catalog == nil {
		return nil
	}
	return c.catalog.List(group)
}

func (c *Catalog) Describe(id string) (*Compiled, error) {
	if c == nil || c.catalog == nil {
		return nil, ErrNoCatalog
	}
	return c.catalog.Describe(id)
}

func (c *Catalog) Coverage() CoverageReport {
	if c == nil || c.catalog == nil {
		return CoverageReport{}
	}
	return c.catalog.Coverage()
}

func Render(c *Compiled) ([]byte, error) {
	return layer.Format(c)
}

func InstanceRenders(set model.InstanceSet, id string) bool {
	return layer.InstanceRenders(set, id)
}

func Summarize(c *Compiled) Summary {
	return layer.Summarize(c)
}

func NewCatalogWith(base *Catalog, compiled *Compiled) (*Catalog, error) {
	if base == nil || base.catalog == nil {
		return nil, ErrNoCatalog
	}
	inner, err := layer.CatalogWith(base.catalog, compiled)
	if err != nil {
		return nil, err
	}
	return &Catalog{catalog: inner}, nil
}
