package domain

import (
	"slices"
	"strings"
)

func (lib *Library) plantOf(domain parsedDomain, found *index, examples []NamedFile) PlantLineage {
	spec := domain.compiled.Spec
	plant := PlantLineage{Domain: PlantDomain{ID: spec.ID, Version: spec.Version, Title: spec.Title, Digest: domain.compiled.Digest, File: domain.file},
		DeviceKinds: plantKinds(domain.implements, found.kinds), Runs: runsOf(domain, found.runs)}
	plant.Composition = lib.compositionSource(domain, found.compositions)
	plant.Twin = lib.twinSource(domain, found)
	plant.Editable = editableOf(plant, plant.Composition == nil && plant.Twin == nil && lib.isExample(domain, examples))
	return plant
}

func plantKinds(implements []string, kinds []kindInfo) []PlantKind {
	plantKinds := make([]PlantKind, 0, len(implements))
	for _, name := range implements {
		if at := slices.IndexFunc(kinds, func(k kindInfo) bool { return k.name == name }); at >= 0 {
			plantKinds = append(plantKinds, PlantKind{Kind: name, Status: KindFound, File: kinds[at].file, Digest: kinds[at].digest})
			continue
		}
		plantKinds = append(plantKinds, PlantKind{Kind: name, Status: KindMissing})
	}
	return plantKinds
}

func runsOf(domain parsedDomain, runs []runInfo) []RunLineage {
	lineage := []RunLineage{}
	for _, run := range runs {
		if run.domain == domain.compiled.Spec.ID {
			run.SameAsCurrent = run.DomainDigest == domain.compiled.Digest
			lineage = append(lineage, run.RunLineage)
		}
	}
	return lineage
}

func (lib *Library) compositionSource(domain parsedDomain, compositions []compositionInfo) *SourceLineage {
	at := slices.IndexFunc(compositions, func(c compositionInfo) bool { return c.id == domain.compiled.Spec.ID })
	if at < 0 {
		return nil
	}
	info := compositions[at]
	composed, err := ComposeDomain(lib, info.raw, info.file)
	source := &SourceLineage{File: info.file, ID: info.id, Version: info.version}
	if err != nil {
		source.Status, source.Reason = SourceDoesNotCompose, oneLine(err.Error())
		return source
	}
	return withProduced(source, composed.Domain.Digest, domain.compiled.Digest)
}

func withProduced(source *SourceLineage, produced, current string) *SourceLineage {
	source.ProducesDigest, source.Status = produced, SourceOutOfSync
	if produced == current {
		source.Status = SourceInSync
	}
	return source
}

func (lib *Library) twinSource(domain parsedDomain, found *index) *SourceLineage {
	at := slices.IndexFunc(found.twins, func(t twinInfo) bool { return t.id == domain.compiled.Spec.ID })
	if at < 0 {
		return nil
	}
	return lib.scaffoldSource(found.twins[at], found, domain.compiled.Digest)
}

func (lib *Library) scaffoldSource(info twinInfo, found *index, current string) *SourceLineage {
	source := &SourceLineage{File: info.file, ID: info.id, Version: info.version}
	kinds, missing := found.kindDocuments(info.kinds)
	if len(missing) > 0 {
		source.Status, source.Reason = SourceKindFilesMissing, "no kind file defines: "+strings.Join(missing, ", ")
		return source
	}
	composed, err := ScaffoldTwin(lib, kinds, info.raw, info.file)
	if err != nil {
		source.Status, source.Reason = SourceDoesNotScaffold, oneLine(err.Error())
		return source
	}
	return withProduced(source, composed.Domain.Digest, current)
}

func (x *index) kindDocuments(names []string) (map[string]map[string]any, []string) {
	documents, missing := map[string]map[string]any{}, []string{}
	for _, name := range names {
		if at := slices.IndexFunc(x.kinds, func(k kindInfo) bool { return k.name == name }); at >= 0 {
			documents[name] = x.kinds[at].doc
		} else {
			missing = append(missing, name)
		}
	}
	return documents, missing
}

func (lib *Library) isExample(domain parsedDomain, examples []NamedFile) bool {
	for _, example := range examples {
		composed, err := ComposeDomain(lib, example.Raw, example.Name)
		if err == nil && composed.Domain.Spec.ID == domain.compiled.Spec.ID && composed.Domain.Digest == domain.compiled.Digest {
			return true
		}
	}
	return false
}
