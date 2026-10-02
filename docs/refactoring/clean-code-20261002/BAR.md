# Clean-code and architecture bar

Scope: all repository Go files and production package boundaries. Baseline:
`4e26cc1`, branch `code-improvements-1`, 103 Go files, 30 over 300 lines.
Existing modifications to `docs/README.md` and `docs/reviews/` are excluded.

## Completion criteria

1. Every Go file has at most 300 physical lines, including tests, comments and
   blanks. A source-walking test enforces this without generated-file exemptions.
2. File names identify responsibilities; related declarations stay together.
   There are no numbered overflow files or new packages created only for size.
3. Entry points state intent and delegate concrete mechanics one level down.
   Each changed function has one responsibility. Functions over 60 lines receive
   an explicit review decision; cohesive declarative tables or numerical
   algorithms may remain together with a recorded reason.
4. An executable import allowlist protects the business package graph. The core
   cannot import CLI, MCP, run orchestration or device transport. Device-world
   integration is a leaf bridge. Schema/model helpers remain foundation packages.
5. Each modified production package includes a meaningful regression test added
   or strengthened in this program. Relocated declarations preserve their bodies
   and existing tests remain intact. Intentional corrections are identified.
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
