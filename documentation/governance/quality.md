# Contributor documentation and quality

The repository’s contributor rules are in [`CONTRIBUTING.md`](../../CONTRIBUTING.md). The root [`AGENTS.md`](../../AGENTS.md) is an agent-only instruction file, not a public governance policy. This page records code and documentation review checks.

## Code organization checks

Every Go file, including tests, is limited to 300 total lines. Name files and functions by simulator responsibility. Entry points should read as named operations with concrete mechanics one level below. Production functions, methods and anonymous functions are limited to 15 physical body lines, including braces, comments and blanks. Test functions are excluded. The source migration is complete. `make function-length` strictly checks all production Go bodies, including build-tagged files, with no exemptions. The check also runs in `make ci-check`, GitHub CI and the local pre-commit hook.

Run `go test ./test/architecture` to check file size, legacy reviewed function lengths and direct package dependencies. The [refactoring bar](../../docs/refactoring/clean-code-20261002/BAR.md) and [review decisions](../../docs/refactoring/clean-code-20261002/REVIEW.md) describe the checks and their limits. Runtime correctness still requires the full repository gate and tests.

For structural work, use Enola's local baseline-and-delta gate: run `make architecture-baseline` before editing, then `make architecture` afterward. The latter enforces new cycles and layer findings; it requires the local tool and a comparable pre-change baseline. Keep generated `.enola/` state local. The [Enola review](../../docs/refactoring/clean-code-20261002/ENOLA.md) records current findings, checked candidates and extraction limits.

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
