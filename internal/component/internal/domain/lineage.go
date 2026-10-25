package domain

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	specs "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const foldersDomains = "domains/"

type parsedDomain struct {
	file       string
	compiled   *specs.Compiled
	implements []string
}

func DescribeLineage(lib *Library, workspace Workspace, only string) (*Lineage, error) {
	domains, err := parseDomains(workspace.Domains)
	if err != nil {
		return nil, err
	}
	selected, err := selectDomain(domains, only)
	if err != nil {
		return nil, err
	}
	return validatedLineage(lib.lineageOf(workspace, domains, selected, only != ""))
}

func (lib *Library) lineageOf(workspace Workspace, domains, selected []parsedDomain, filtered bool) *Lineage {
	found := lib.indexWorkspace(workspace)
	lineage := &Lineage{SchemaVersion: lineageVersion, Plants: make([]PlantLineage, 0, len(selected)), Unreadable: found.unreadable}
	for _, domain := range selected {
		lineage.Plants = append(lineage.Plants, lib.plantOf(domain, found, workspace.Examples))
	}
	lineage.DeviceKinds = deviceKindsOf(found, domains, selected, filtered)
	return lineage
}

func validatedLineage(lineage *Lineage) (*Lineage, error) {
	raw, err := json.Marshal(lineage)
	if err != nil {
		return nil, fmt.Errorf("component: encode lineage: %w", err)
	}
	if err := model.ValidateDomainLineage(raw); err != nil {
		return nil, fmt.Errorf("component: %w", err)
	}
	return lineage, nil
}

func parseDomains(files []NamedFile) ([]parsedDomain, error) {
	domains := make([]parsedDomain, 0, len(files))
	for _, file := range files {
		domain, err := parseDomain(file)
		if err != nil {
			return nil, err
		}
		domains = append(domains, domain)
	}
	slices.SortFunc(domains, func(a, b parsedDomain) int {
		return strings.Compare(a.compiled.Spec.ID+"\x00"+a.file, b.compiled.Spec.ID+"\x00"+b.file)
	})
	return domains, nil
}

func parseDomain(file NamedFile) (parsedDomain, error) {
	if file.Err != nil {
		return parsedDomain{}, fmt.Errorf("component: read %s%s: %w", foldersDomains, file.Name, file.Err)
	}
	compiled, err := specs.Parse(file.Raw, foldersDomains+file.Name)
	if err != nil {
		return parsedDomain{}, fmt.Errorf("component: %w", err)
	}
	return parsedDomain{file: foldersDomains + file.Name, compiled: compiled, implements: implementedKinds(compiled.Spec)}, nil
}

func implementedKinds(spec *model.DomainSpec) []string {
	names := []string{}
	for i := range spec.Kinds {
		if name := spec.Kinds[i].Implements; name != "" && !slices.Contains(names, name) {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}

func selectDomain(domains []parsedDomain, only string) ([]parsedDomain, error) {
	if only == "" {
		return domains, nil
	}
	selected := slices.DeleteFunc(slices.Clone(domains), func(d parsedDomain) bool { return d.compiled.Spec.ID != only })
	if len(selected) == 0 {
		return nil, fmt.Errorf("component: lineage: %w %q", specs.ErrUnknownDomain, only)
	}
	return selected, nil
}
