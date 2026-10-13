# P5a review record — domain-spec v0.2, kinds and instances

- **Reviewed**: the P5a working-tree diff against `07d0ad6` (2026-10-12)
- **Reviewer**: one independent subagent, per the owner's direction to keep reviews light.
- **Outcome**: 0 High, 2 Medium, 4 Low. Resolved below or carried in [FINDINGS.md](../../../governance/FINDINGS.md).

| # | Sev | Finding | Disposition |
|---|---|---|---|
| M1 | Med | Part 1 allows one kind, but the world change that several kinds need has no owning phase | Carried to the new phase P5a2 (multi-kind world), added to PLAN by the owner |
| M2 | Med | The R2-02 id check matches every explicit id against the instance set's `id_template`, so it will reject the named instances (`M1`, `AC1`) P5b adds | Carried to P5b: named instances declare their ids, and the check admits exactly those ids as well as template ids |
| L1 | Low | Comments added in the diff against the owner's no-comments rule | Fixed: removed from `api.go`, `operations.go`, `files.go`, `instances.go`, `validation.go` and `model/domain.go`; only the `determinism-safe` marker and `//go:embed` remain |
| L2 | Low | The equivalence table could miss a ninth shipped domain | Fixed: `TestTheEquivalenceTableNamesEveryShippedDomain` compares the table keys with the files in `domains/` |
| L3 | Low | `compactValue` rendered a JSON `null` through the default branch | Fixed: an explicit `nil` case renders `null` |
| L4 | Low | `domain format` wrote the file in place, so a failure mid-write could truncate a domain | Fixed: the formatted bytes go to a temp file in the same directory, take the original mode, then replace the file by rename; the test checks the mode and that no staging file remains |
| L5 | Low | Objects that fit in 110 columns are inlined, which allows long lines | Accepted as cosmetic: the layout is deterministic and idempotent, and the width is a constant in one place |
