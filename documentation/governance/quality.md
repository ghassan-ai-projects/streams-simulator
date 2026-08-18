# Contributor documentation and quality

The repository’s contributor rules are in [`CONTRIBUTING.md`](../../CONTRIBUTING.md). The root [`AGENTS.md`](../../AGENTS.md) is an agent-only instruction file, not a public governance policy. This page adds the documentation-specific review bar.

## Before changing docs

- Read [the quality bar](../QUALITY_BAR.md).
- Check the working tree and preserve unrelated changes.
- Identify the authoritative source for every fact you will publish.
- Label designed, partial, historical, and superseded material.
- Add or update a `Next reads` section on every substantive page.

## Documentation check

Run the repository-native documentation smoke check first:

```bash
make docs-check
```

Then run this review from the repository root:

```bash
git diff --check
rg -n '\[[^]]+\]\([^)]+' documentation README.md
rg -n '^```' documentation
go run ./cmd/streamsim help
go run ./cmd/streamsim catalog list
go run ./cmd/streamsim adapter list
```

Manually verify every changed relative link, every copyable command, every Mermaid diagram, and every claim that depends on code or tests. `docs-check` intentionally does not fetch network URLs or replace human review of link targets and diagrams; it is a smoke check, not a full Markdown parser or link checker.

## Review questions

- Can a new reader find the first successful run in under five minutes?
- Does every public command match the current CLI implementation?
- Are domain/adapter counts separated into working-tree, committed, and release inventories?
- Does each of the nine non-negotiables link to evidence?
- Could a reader mistake a design record for an implemented guarantee?
- Are security reporting, support, license, contribution, and conduct paths discoverable?

## Next reads

- [Quality bar](../QUALITY_BAR.md)
- [Limitations](../limitations.md)
- [Repository contribution rules](../../CONTRIBUTING.md)
