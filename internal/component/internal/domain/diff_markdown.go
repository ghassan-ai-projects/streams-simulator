package domain

import (
	"fmt"
	"strings"
)

func RenderDiffMarkdown(diff *Diff) []byte {
	var out strings.Builder
	out.WriteString("# Domain diff\n\n")
	writeSides(&out, diff)
	for _, kind := range diff.Kinds {
		writeKind(&out, kind)
	}
	writeChanges(&out, "Relations and other members", diff.DomainChanges)
	writeUnchanged(&out, diff)
	return []byte(out.String())
}

func writeSides(out *strings.Builder, diff *Diff) {
	out.WriteString("| | A | B |\n|---|---|---|\n")
	row := func(name string, a, b string) { fmt.Fprintf(out, "| %s | %s | %s |\n", name, cell(a), cell(b)) }
	row("Domain", diff.A.ID, diff.B.ID)
	row("Version", diff.A.Version, diff.B.Version)
	row("Digest", diff.A.Digest, diff.B.Digest)
	row("Source", diff.A.Source.Kind+" "+diff.A.Source.Label, diff.B.Source.Kind+" "+diff.B.Source.Label)
	if diff.Identity.SameIDAndVersionDifferentDigest {
		out.WriteString("\nThe model changed and its version did not: same id, same version, another digest.\n")
	}
	fmt.Fprintf(out, "\n%d change(s): %d kind(s) changed, %d added, %d removed.\n", diff.Summary.Changes, diff.Summary.KindsChanged, diff.Summary.KindsAdded, diff.Summary.KindsRemoved)
}

func writeKind(out *strings.Builder, kind KindDiff) {
	fmt.Fprintf(out, "\n## Kind %s (%s)\n\n", kind.Kind, kind.Status)
	fmt.Fprintf(out, "Instances: %s; now %s.\n", describeInstances(kind.Instances.Before), describeInstances(kind.Instances.After))
	for _, relation := range kind.Relations {
		fmt.Fprintf(out, "Relation %s (%s to %s): %s.\n", relation.Type, relation.FromKind, relation.ToKind, strings.Join(relation.Links, ", "))
	}
	out.WriteString("\n")
	writeList(out, kind.Changes)
}

func describeInstances(summary InstanceSummary) string {
	return fmt.Sprintf("%d %s", summary.Count, strings.Join(namesOf(summary), ", "))
}

func writeChanges(out *strings.Builder, title string, changes []Change) {
	if len(changes) == 0 {
		return
	}
	fmt.Fprintf(out, "\n## %s\n\n", title)
	writeList(out, changes)
}

func writeList(out *strings.Builder, changes []Change) {
	for _, change := range changes {
		fmt.Fprintf(out, "- `%s` %s\n", change.Code, change.Sentence)
	}
}

func writeUnchanged(out *strings.Builder, diff *Diff) {
	same := diff.Unchanged
	fmt.Fprintf(out, "\n## Unchanged\n\nkinds %d, readings %d, dynamics %d, commands %d, faults %d, states %d, instance sets %d, links %d, link types %d.\n",
		same.Kinds, same.Readings, same.Dynamics, same.Commands, same.Faults, same.States, same.InstanceSets, same.Links, same.LinkTypes)
}

func cell(text string) string {
	return strings.NewReplacer("|", `\|`, "\n", " ").Replace(text)
}
