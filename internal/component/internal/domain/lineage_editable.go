package domain

import "slices"

func editableOf(plant PlantLineage, example bool) Editable {
	switch {
	case plant.Composition != nil && plant.Composition.Status == SourceInSync:
		return editable(VerdictEditable, ReasonComposition, RouteEditComposition, "Edit the composition "+plant.Composition.File+", then compose it again.")
	case plant.Twin != nil && plant.Twin.Status == SourceInSync:
		return editable(VerdictReadOnly, ReasonTwin, RouteOpenDeviceBehavior, "Generated from the twin choices "+plant.Twin.File+": change its readings and responses in the device editor, then create the twin again.")
	case plant.Composition != nil:
		return editable(VerdictEditable, ReasonCompositionOutOfSync, RouteRecompose, "The composition "+plant.Composition.File+" no longer produces this domain file: edit it and compose it again to replace the file.")
	case plant.Twin != nil:
		return editable(VerdictReadOnly, ReasonTwinOutOfSync, RouteRescaffoldTwin, "The twin choices "+plant.Twin.File+" no longer produce this domain file: change its readings and responses in the device editor and create the twin again.")
	case example:
		return editable(VerdictReadOnly, ReasonExample, RouteCreateFromExample, "A shipped example: create a plant from it under a new id, then edit that composition.")
	}
	return editable(VerdictReadOnly, ReasonHandAuthored, RouteCopyDomainFile, "Authored as a domain file: copy the file as a new domain id and edit the copy.")
}

func editable(verdict, reason, route, sentence string) Editable {
	return Editable{Verdict: verdict, Reason: reason, Route: Route{Code: route, Sentence: sentence}}
}

func deviceKindsOf(found *index, all, selected []parsedDomain, onlySelected bool) []DeviceKindLineage {
	entries := []DeviceKindLineage{}
	for _, kind := range found.kinds {
		if onlySelected && !implementsKind(selected, kind.name) {
			continue
		}
		entries = append(entries, deviceKindOf(kind, found.twins, all))
	}
	return entries
}

func implementsKind(domains []parsedDomain, kind string) bool {
	return slices.ContainsFunc(domains, func(d parsedDomain) bool { return slices.Contains(d.implements, kind) })
}

func deviceKindOf(kind kindInfo, twins []twinInfo, domains []parsedDomain) DeviceKindLineage {
	entry := DeviceKindLineage{Kind: kind.name, File: kind.file, Digest: kind.digest, TwinsIncluding: []TwinMembership{}, ImplementedBy: []string{}}
	for _, twin := range twins {
		if slices.Contains(twin.kinds, kind.name) {
			entry.TwinsIncluding = append(entry.TwinsIncluding, TwinMembership{Twin: twin.id, File: twin.file, Kinds: len(twin.kinds)})
			entry.HasOwnTwin = entry.HasOwnTwin || len(twin.kinds) == 1
		}
	}
	for _, domain := range domains {
		if slices.Contains(domain.implements, kind.name) {
			entry.ImplementedBy = append(entry.ImplementedBy, domain.compiled.Spec.ID)
		}
	}
	return entry
}
