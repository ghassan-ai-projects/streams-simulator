# Documentation organization plan

> Status: Completed reorganization plan and maintenance record. Authority: repository inventory, Tamoz pattern review, and independent documentation reviews. Verified by: final documentation validation loop. Last verified: 2026-08-17.

This plan defines how the repository documentation will be reorganized. It is also the maintenance contract for the new documentation set.

## Outcome

Create a curated, open-source-ready documentation set under `documentation/` while preserving the engineering record under `docs/`.

The root `README.md` remains the product entry point. `documentation/README.md` becomes the detailed documentation hub. `docs/README.md` remains the working archive and design source record until a future project decision changes that boundary.

## Design principles

1. **Readers follow intent.** Organize around what a person is trying to do, not around the order in which documents were written.
2. **One home for each public fact.** The README or a hub page summarizes; the linked page owns the detail.
3. **Working evidence stays available.** Plans, audits, reviews, and research are not deleted; they are linked as archive material and labeled by status.
4. **Implemented and designed are different states.** Public pages must say whether a behavior is implemented, partial, proposed, or historical.
5. **Contracts stay machine-readable.** JSON schemas remain authoritative artifacts; prose explains how and when to use them.
6. **Examples are executable evidence.** Every quickstart command must correspond to the current CLI and shipped files.
7. **The quality bar is a gate.** Reviewers use `QUALITY_BAR.md`, not taste, to decide whether the set is complete.
8. **Claims are reconciled before publication.** Every material claim is mapped to code, a test/evidence artifact, or an explicitly labeled design record.

## Target information architecture

```text
documentation/
├── README.md                         documentation hub and reading paths
├── QUALITY_BAR.md                    acceptance criteria and review evidence
├── limitations.md                    honest scope and known limitations
├── roadmap.md                        shipped, in progress, and planned work
├── overview/
│   ├── product.md                    what it is, is not, and who it serves
│   ├── concepts.md                   mental model and vocabulary
│   └── compatibility.md              Go, platform, dependency, and version policy
├── getting-started/
│   ├── install.md                    prerequisites and build setup
│   └── quickstart.md                 first catalog, run, replay, and score
├── architecture/
│   ├── overview.md                   components and dependency direction
│   ├── pipeline.md                   world → perturbation → adapter → sink → ledger
│   ├── determinism.md                seeds, command logs, artifacts, replay
│   ├── domains-and-adapters.md       data-defined extension model
│   ├── truth-and-scoring.md          sealing, verdicts, ledger, and scoring
│   ├── security-model.md             roles, capabilities, injection boundary
│   └── invariants.md                 nine non-negotiables and their evidence
├── design/
│   └── README.md                     public design summaries and authoritative archive map
├── guides/
│   ├── consumer-integration.md      reference consumer and operator workflow
│   ├── authoring-domains.md         validate and extend domain JSON
│   ├── authoring-adapters.md        build and verify adapter JSON
│   ├── replay-and-debugging.md      reproduce failures and inspect artifacts
│   └── benchmark-workflow.md        suites, baseline audit, verdicts, scores
├── benchmark/
│   ├── README.md                     benchmark model, scope, and evidence posture
│   ├── methodology.md                scenario validity and scoring methodology
│   └── evidence.md                   release evidence and reproducibility expectations
├── operations/
│   ├── operations.md                run directories, artifacts, and failure handling
│   ├── troubleshooting.md           common failures and evidence-first diagnosis
│   └── release.md                   manifest, validation, and publication rules
├── reference/
│   ├── cli.md                       complete CLI command and flag reference
│   ├── mcp.md                       director/operator tools and role boundary
│   ├── contracts.md                 schema catalog and compatibility rules
│   └── artifacts.md                 run, ledger, verdict, and score outputs
├── adr/
│   └── README.md                    curated architecture decisions with archive links
├── research/
│   ├── README.md                    published research context and provenance
│   ├── transport.md                 control/evidence transport rationale
│   ├── prior-art.md                 adopted and rejected prior art
│   └── benchmark-validity.md        benchmark validity controls
└── governance/
    ├── quality.md                   contributor documentation and review checks
    └── claims-matrix.md             public-claim authority and evidence map
```

The repository root remains the home for policy files such as `CONTRIBUTING.md`, `SECURITY.md`, `LICENSE`, `CODE_OF_CONDUCT.md`, `SUPPORT.md`, and `CHANGELOG.md`. If one is absent, the public documentation must say so rather than imply that a policy exists.

## Migration map

| Existing material | New public home | Treatment |
| --- | --- | --- |
| `docs/README.md` | `docs/README.md` | Rewrite as the engineering archive policy; the public hub is `documentation/README.md`. |
| `docs/QUALITY_BAR.md` | `documentation/QUALITY_BAR.md` | Keep the new public acceptance bar here; retain the implementation gate in the archive until reconciled. |
| `docs/design/TECHNICAL_DESIGN.md` | `architecture/overview.md`, `architecture/pipeline.md`, `architecture/determinism.md`, `architecture/security-model.md` | Split by reader question; link to the design record for normative detail. |
| `docs/design/MCP_SURFACE.md` | `reference/mcp.md` and `guides/consumer-integration.md` | Separate protocol reference from workflow guidance. |
| `docs/design/CONSUMERS.md` | `guides/consumer-integration.md` | Make the reference consumer path executable and explicit about the operator boundary. |
| `docs/design/IMPLEMENTATION_PLAN.md`, reviews, and readiness notes | `design/README.md` | Publish a status-aware design reading order; keep detailed execution history in the archive. |
| `docs/design/GROUND_TRUTH_AND_SCORING.md` | `architecture/truth-and-scoring.md` and `guides/benchmark-workflow.md` | Explain model first, procedure second. |
| `docs/design/DOMAIN_CATALOG.md` | `architecture/domains-and-adapters.md` and `guides/authoring-domains.md` | Keep coverage/catalog detail in the archive and expose the extension path publicly. |
| `docs/contracts/` | `reference/contracts.md` | Keep JSON schemas authoritative; add a prose index and usage rules. |
| `docs/adapters/`, `docs/examples/` | `architecture/domains-and-adapters.md` | Treat as historical or test fixtures; link public readers to `adapters/` and `domains/`. |
| `adapters/vendor/`, `.github/`, `.agents/`, `release-manifest.json` | `reference/contracts.md`, `benchmark/`, `governance/quality.md` | Document their role without creating duplicate public sources of truth. |
| `docs/research/` | `research/README.md` | Curate research context and label non-normative material. |
| `docs/DECISIONS.md` | `adr/README.md` | Curate the small set of public architecture decisions and link to the working record. |
| `docs/*PLAN*`, `docs/*REVIEW*`, audits, and handover notes | `docs/` archive | Preserve as engineering history; link only when it helps a reader verify a claim. |
| root `CONTRIBUTING.md`, `SECURITY.md`, `LICENSE`, `AGENTS.md` | `governance/quality.md` and documentation hubs | Keep root files authoritative; link to them instead of copying policy. |

## Status labels

Every page that discusses behavior or a contract must use one of these labels when the status is not obvious:

- **Implemented** — present in the current code and covered by tests or a documented verification command.
- **Partially implemented** — some path exists, but the stated design is not fully available.
- **Designed** — specified in the working archive but not yet implemented or independently verified.
- **Proposed** — a future idea or decision candidate; it is not a current promise.
- **Historical** — retained to explain how the project arrived at the current design.
- **Superseded** — replaced by a later decision or implementation; retained only for traceability.

Substantive pages should also state `Authority`, `Implementation status`, `Verified by`, `Last verified`, and `Supersedes` when those fields affect how the page should be trusted.

## Review sequence

1. Inventory the repository, existing docs, and Tamoz pattern.
2. Review this plan with independent agents for structure, scope, and missing reader journeys.
3. Build the hubs and core public pages first.
4. Add workflow guides and references from the implemented CLI, MCP, schemas, and tests.
5. Run a completeness/style/diagram review and a code-alignment review.
6. Fix findings, then repeat gap and alignment reviews until the quality bar passes.
7. Run `make docs-check`, then validate links, Markdown structure, commands, repository tests, and `git diff --check`.

The handoff must include a claim-to-source matrix covering product claims, commands, versions, domain/adapter counts, MCP tools, artifact files, contracts, and release status. The source column must point to code, tests, schemas, or a clearly labeled design/archive document.

## Non-goals

- Do not delete engineering history from `docs/` during this reorganization.
- Do not claim benchmark readiness, production safety, or complete implementation merely because a design document exists.
- Do not duplicate machine-readable schemas into prose or maintain two conflicting contract copies.
- Do not add a docs website, generator, or external dependency unless the repository already requires it.
- Do not call the documentation complete while the claim-to-source matrix contains an unresolved contradiction.

## Next reads

- [Public documentation hub](../documentation/README.md)
- [Public quality bar](../documentation/QUALITY_BAR.md)
- [Engineering archive](README.md)
