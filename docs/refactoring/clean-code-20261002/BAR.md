# Clean-code and architecture bar

Scope: all repository Go files and production package boundaries. Baseline:
`4e26cc1`, branch `code-improvements-1`, 103 Go files, 30 over 300 lines.
Existing modifications to `docs/README.md` and `docs/reviews/` were excluded from
refactoring rounds; a concurrent publication commit later included them.

## Completion criteria

1. Every Go file has at most 300 physical lines, including tests, comments and
   blanks. A source-walking test enforces this without generated-file exemptions.
2. File names identify responsibilities; related declarations stay together.
   There are no numbered overflow files or new packages created only for size.
3. Entry points state intent and delegate concrete mechanics one level down.
   Every production function, method and anonymous function has at most 15
   physical body lines, including braces, comments and blanks. Tests are exempt
   from the function limit and remain unchanged during this migration. Extract
   named responsibilities without compressing statements or hiding complexity.
4. An executable import allowlist protects the business package graph. The core
   cannot import CLI, MCP, run orchestration or device transport. Device-world
   integration is a leaf bridge. Schema/model helpers remain foundation packages.
5. Each behavior or workflow change includes a meaningful regression test added
   or strengthened in its package. Pure relocation preserves declaration bodies
   and existing oracles, with source-comparison evidence. Intentional corrections
   are identified and covered by regressions.
6. Focused tests, self-review and whitespace checks precede every code commit.
   Final evidence includes `make ci-check`, full tests with coverage, build/vet/
   lint, and hook checks if available. Environment failures are not passes.
7. Preserve determinism, replay, analytic oracles, role separation, JSON wire
   shapes, command order, locks, RNG draws, effects and digest inputs.

## Architecture decision

One Go module contains business packages: domain/catalog, world, perturbation,
adapter, sink, run/replay/ledger, truth, score, audit, suite and reference consumer.
The CLI and MCP are application surfaces. Device and deviceworld are bounded
emulation and integration packages. Model, schemas, canonical, jsonschema,
randutil and wall are shared foundations with specific responsibilities.

Keep these packages. Large files show missing internal organization, not evidence
for another Go module or a generic service/store abstraction. Partition files by
responsibility first. Introduce a package only when a distinct ownership boundary
and current caller justify it; document a design change before moving ownership.

The architecture test records direct imports, not transitive reachability, because
run necessarily composes the pipeline. Go itself rejects import cycles. Trust,
determinism and behavior remain covered by the existing runtime tests.

## Root-cause analysis

Why are responsibilities difficult to read? Large files contain loading, validation,
execution and reporting together. Why? New operations accumulated in the original
package entry files. Why did this persist? Quality checks addressed correctness and
lint but neither file size nor package direction. Why would splitting alone fail?
Large functions still mix orchestration and mechanics, and duplicated scoring
policies can drift. Why this bar? Named files, stepdown functions, one owner for
shared scoring rules, and executable structure checks address those causes while
retaining the documented architecture.

## Round plan

- Establish the bar and dependency guard.
- Partition production and test files by responsibility; enforce 300 lines.
- Simplify scoring into shared policies and prove online/offline parity.
- Refactor run lifecycle and delivery orchestration without changing ordering.
- Refactor remaining entry-point hotspots and add boundary regression evidence.
- Review all exceptions, run complete gates, and commit final evidence.

Known correctness and release findings in `docs/reviews/ARCHITECTURE_REVIEW.md`
remain separate from readability completion. This program must not imply all prior
release risks were resolved through file splitting.

## Stricter function limit — 2026-10-03

The user tightened the function limit from the former reviewed 60-line threshold
to 25, then selected a hard maximum of **15 lines**. The earlier round-10
completion claim applies to the old bar. Rounds 12–39 completed the stricter
source migration; final gate evidence is recorded in [HANDOFF.md](HANDOFF.md).

The initial source-only audit found 281 named production functions and seven
anonymous functions over 15 lines; the completed audit has zero violations.
Run `make function-length` for the strict AST-based check. It counts all production Go files, including build-tagged sources, and
excludes test files. It runs in the full CI gate and local pre-commit hook.
No allowances or frozen-body exemptions apply.

Refactor by owning package, preserving expression order, RNG consumption, errors,
locking and digest representations. Keep test files unchanged; existing replay,
analytic, protocol and fixture oracles provide behavioral evidence. Commit each
validated round. Enola baselines precede structural edits; receipt comparability
and new coupling are reviewed afterward.
