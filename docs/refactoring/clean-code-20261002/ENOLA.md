# Enola architecture review

Date: 2026-10-03. Repository: Streams Simulator. Starting commit: `ecd5b98`.
This round follows Tamoz's generate → pin → edit → regenerate → compare workflow.
The baseline includes the existing working-tree review documents; it is a local
architecture reference, not a clean release artifact. It was pinned after round
10 and does not retrospectively grade rounds 1–10.

## Project integration

- [mcp-arch.yaml](../../../mcp-arch.yaml) owns this repository's configuration.
  Built-in extractors, explainers and ignore defaults are inherited. No new
  filters, suppressions, runtime dependencies or installed hooks are introduced.
- `make architecture-baseline` snapshots and pins the pre-change architecture.
- `make architecture` compares the working tree against that baseline and fails
  on new cycle/layer findings at confidence 0.8, matching Tamoz's policy. It is
  read-only and does not replace the baseline. `ENOLA` can select a local binary.
- The generated `.enola/` directory is ignored. Durable evidence belongs in this
  document and [ROUNDS.md](ROUNDS.md), not links to generated files.
- [AGENTS.md](../../../AGENTS.md) specifies the pre-edit impact review, baseline,
  post-edit receipt comparison and structural diff. These local targets complement
  `make ci-check`; they are not added to fresh CI jobs that lack a pre-change pin.

## Findings

Enola 0.4.25, extractor version `v276`, indexed 23 package/module nodes: 21 internal
packages, the command package and the root. The initial scan produced 2,484 facts
and 143 heuristic insights with zero parse errors.

| Check | Result and limit |
| --- | --- |
| Dependency cycles | No cycle insights at any confidence. These are package/module dependency cycles; ordinary recursive JSON traversal and runtime feedback are different concepts. |
| Inferred layers | Recognizes `go-standard` at 0.95 confidence; 22 of 23 nodes classified, one import between ordered layers, none against their order. The inference groups all internal packages into one layer. |
| Business package boundaries | `go test ./test/architecture` passes. Its explicit direct-import allowlist checks ownership within `internal`, beyond Enola's generic layer inference. |
| Domain dynamics | Domain tests pass, including dependency-cycle rejection and shipped domain validation. |
| Structural delta | Receipts are comparable; the setup introduces no new cycle/layer findings. The initial delta adds an AGENTS heading and its documentation edges, not runtime imports. |
| Extraction limits | Receipt comparison reports no quality regression. The refreshed receipt also reports 316 unresolved extraction references; absence of a finding is not proof of complete call-graph resolution. |

The server announces 0.4.26 with changed extractors. This work uses 0.4.25 in both
CLI and MCP. An upgrade needs a coverage/comparability review before replacing a
baseline; mixing extractor versions cannot establish a clean structural delta.

Initial candidate counts: performance 48, high fan-in 25, dead-code 20, hotspots
20, complexity 15, dependency depth 10, exported surface 2, package metrics 2, and
one positive layer-pattern finding. These counts include documentation link nodes
and static-analysis heuristics. They are review leads, not 143 confirmed defects.

## Candidate decisions

| Candidate | Source review |
| --- | --- |
| `auditCapture.sampleGrid` reported as cubic | The per-channel delivery index only advances. Actual work is proportional to channel × grid points plus consumed samples; loop nesting alone overstates the complexity. Preserve the delivered-evidence grid. |
| `score.matchesActions` pairwise matching | Actual worst-case action × effector-call scanning. This is a valid large-run scaling candidate already recorded in REVIEW.md; optimization needs a workload and identity/multiplicity regressions. |
| `mcp.errTool`, time formatting/decoding helpers have high fan-in | Small shared boundary helpers with intentional reuse. Fan-in alone does not justify another module or duplicate implementations. |
| `mcp.toolSchema`, `canonical.writeValue` have high complexity | Closed declarative/type dispatch tables with the existing bounded review decisions in REVIEW.md. Splitting by branch count would scatter one contract. |
| JSON document structs reported as dead-code candidates | Decoder/reflection usage and type references are not fully represented by static call edges. Keep wire structs; corroborate real unreachable behavior with the Go deadcode tool and tests. |
| Deep dependency chains / public model surface | Application composition eventually reaches domain/model/schema/canonical foundations. No reverse dependency or cycle is reported; shared wire records intentionally expose fields/types. No speculative package extraction is justified. |

## Validation

`make architecture` passes with an explicit `cycles,layers` policy, confidence 0.8
and warnings-only mode disabled. This is not a report-only exit zero.
`compare_receipts` confirms equivalent extraction inputs; `diff_snapshot` is
reviewed for documentation changes as well as runtime coupling.

Disposable source fixtures verify the installed gate's failure behavior:

| Introduced condition | Observed outcome |
| --- | --- |
| New `internal/a` ↔ `internal/b` import cycle | Regression, exit 1. |
| New `internal/a` → `cmd/control` upward dependency | Regression, exit 1. |
| No pinned baseline | Error, exit 2. |
| Changed ignore globs between pin and check | Incomparable, exit 3. |

The fixtures use temporary directories and disabled history, leaving this
repository's pin intact. Architecture/domain tests, documentation smoke/link
checks, and whitespace checks pass. There is no runtime source change in this
round, so the previously completed full race/coverage gate is not repeated.
