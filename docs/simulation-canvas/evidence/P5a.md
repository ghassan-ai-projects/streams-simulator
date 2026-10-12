# P5a evidence — domain-spec v0.2, kinds and instances

- **Date**: 2026-10-12
- **Change**: domain specs are written as **kinds** (the entity template) and **instances** (the
  population that follows a kind). The document compiles to the existing runtime form, so the
  world, perturbation, adapter, truth, score and run code are unchanged. Spec:
  [`docs/design/DOMAIN_SPEC_V0_2.md`](../../design/DOMAIN_SPEC_V0_2.md); contract
  `docs/contracts/domain-spec-v0.2.schema.json` (v0.1 removed).
- **Not in this phase** (P5b): add-or-narrow overrides, typed links, system entities, the board.
  Part 1 is one kind and one instance set per domain (`maxItems: 1`), named instances arrive with
  overrides.

## What was built

| Area | Change |
|---|---|
| Contract | `domain-spec-v0.2.schema.json` (`kinds[]`, `instances[]`, v0.1 sections moved verbatim into the kind); embedded copy in `internal/schemas`, byte parity in `TestEmbeddedMatchCommitted`; the schema suite validates all eight shipped domains |
| `internal/model` | `DomainDocument`, `Kind`, `InstanceSet` (the v0.2 document); `DomainSpec` is the runtime form it lowers to; `EntityCount` named; `ContractDomainSpec` is `domain-spec-v0.2`; `SimVersion` `0.3.0` |
| `internal/domain` | `lowering.go` (document to runtime form, kind and instance name must agree), `pointer.go` (runtime pointer to v0.2 pointer), `instances.go` (instance count bounds, `Compiled.CheckEntityIDs`), `format.go` and `files.Format` (canonical formatting); every cross-check issue carries a pointer; digest is of the v0.2 document |
| `internal/jsonschema` | "missing required property" and "additional property" issues point at the member (R1-01) |
| `internal/run` | `newRunWorld` refuses explicit entity ids that do not fit the domain's instances (R2-02) |
| `internal/cli` | `domain format <path>`; `domain_format` in `cli-output-v0.1`; usage text |
| `domains/`, `docs/examples/` | Eight domains rewritten to v0.2 and formatted with `domain format`; the documented example is the same bytes as the shipped pond |
| Tests | `testsupport.AsKinds` lifts a flat (runtime-shaped) test document into kinds and instances, so inline domain fixtures and the world test spec builder keep their shape |
| Fixtures | The pinned report run (`internal/report/testdata/run`): `domain_spec` is the v0.2 pond, and the recorded digests that depend on the domain digest or `sim_version` (`journal.jsonl`, `world_digest`, `journal_digest`) and the two report goldens are updated. Its trace digest is unchanged and `streamsim verify` reproduces it |
| Docs | `DOMAIN_SPEC_V0_2.md`, `DECISIONS.md` (D-03, D-04, D-14 deviation entries, D-15), `CLI_SURFACE.md`, `documentation/{guides/authoring-domains,reference/cli,reference/contracts,architecture/domains-and-adapters}.md`, `internal/domain/UBIQUITOUS_LANGUAGE.md`, `CHANGELOG.md`, `PLAN.md`, `FINDINGS.md` |
| Release manifest | Regenerated with `make manifest` (same author and reviewer identities): new domain file hashes, commit `07d0ad6`, go1.27.2. It stays a historical record, as `documentation/limitations.md` says |

## Decisions the spec left open

| Question | Decision | Why |
|---|---|---|
| Shape of a kind | A named object with the five v0.1 sections; `name` becomes `entity_type` | The v0.1 sections already describe one entity; v0.1's optional `entity_type` was the kind name |
| How many kinds | Exactly one (`maxItems: 1`) | The world holds one state, channel and fault set per domain; several kinds are a world change, which this part must not make |
| Named instances and per-instance parameters | None in part 1 | v0.1 allowed no variation between entities; variation is the P5b override |
| Digest | RFC 8785 digest of the v0.2 document | Same rule as v0.1; formatting never changes it |
| Runtime form | `model.DomainSpec`, unchanged | Everything below the loader keeps reading it |
| R2-02 | Explicit ids: count within `count.min..max` and each id matches the `id_template` (`{n}` digits, other placeholders `[A-Za-z0-9_.-]+`), checked in `run` when the world is created | One place covers CLI flags, scenario files, the director and replay; ids matching the template are all the template can render |
| R1-01 | Missing and additional property issues point at the member; cross-checks carry pointers into the v0.2 document | The canvas pins errors to the element |
| `sim_version` | `0.3.0` | Every run's `domain_digest`, `world_digest` and `domain_spec` change |

## Runtime equivalence

| Proof | Result |
|---|---|
| `TestEveryShippedDomainCompilesToItsV01RuntimeWorld` | The RFC 8785 digest of each rewritten domain's compiled runtime form equals the digest of its v0.1 compiled form, recorded before the rewrite (`aquaculture-pond efb583f5…`, one row per domain in the test) |
| Behaviour pin, 8 domains x 2 adapters (`seed 7`, 600 s) and the scripted fault-and-perturbation run | `trace.jsonl`, `ledger.jsonl`, `world_state_history.jsonl`, every `verify/*` and `replay/*` result, `suite/*` and `adapter-verify/*`, `catalog-list`, `catalog-coverage` are byte-identical before and after |
| Pinned report run | A recorded run with faults, drop and duplicate perturbations and `silent_no_effect` calls replays under the v0.2 pond with the same trace digest (`sha256:8dc7384c…`) |

## Behaviour pin changes (all deliberate)

| Pin line | Why it changes |
|---|---|
| `domain-validate/<id>` (8) | The printed `digest` is the digest of the v0.2 document |
| `catalog-describe/<id>` (8) | Same digest; the printed `spec` (the runtime form) is identical |
| `run/<id>/<adapter>/run.json` (16), `scripted/run.json` | `domain.digest`, `domain_spec` (now the v0.2 document), `world_digest` (includes the domain digest and `sim_version`), `sim_version` `0.3.0`, `journal_digest` |
| `run/<id>/<adapter>/journal.jsonl` (16), `scripted/journal.jsonl` | The `run.opened` record carries the domain digest and `sim_version` |
| `manifest-file` | The manifest embeds the hashes of the domain files, which were rewritten |
| `failure/domain-invalid/stdout`, `…/stderr` | The contract is `domain-spec-v0.2` and the "missing required property" issues now point at their members |
| `failure/unknown-command/stderr` | The usage text lists `domain format` |

Unchanged: `trace.jsonl`, `ledger.jsonl`, `world_state_history.jsonl`, `sealed/journal.jsonl`
(the director stream does not carry the domain digest), every `verify/*` and `replay/*`,
`suite/*`, `adapter-*`, `catalog-list`, `catalog-coverage`, and the other `failure/*` documents.

## Findings closed

| ID | Closed by |
|---|---|
| R1-01 | `internal/jsonschema` points a missing or additional property at the member; every domain cross-check issue is located (`TestEveryCrossReferenceIssueCarriesAPointerIntoTheV02Document`, `TestAMissingRequiredMemberIsReportedAtTheMember`, `TestACrossReferenceViolationIsADomainIssueAtItsPointer`, the corrected expectations in the scenario, run-artifact and device conformance tests) |
| R2-02 | `Compiled.CheckEntityIDs` in `newRunWorld` (`TestExplicitEntityIDsMustFitTheInstanceTemplateAndCountBounds`, `TestAScenarioWithEntityIdsOutsideTheDomainInstancesIsAViolation`, `TestCreateWorldRefusesEntityIdsThatDoNotFitTheDomainInstances`) |
| R0-10 | `streamsim domain format` (`TestDomainFormatRewritesCanonicallyAndIsIdempotent`, `TestEveryShippedDomainIsAlreadyCanonicallyFormatted`, `TestFormatKeepsTheDigestAndTheRuntimeWorldWhateverTheLayout`) |
| S15 | `DECISIONS.md` D-03, D-04 and D-14 deviation entries, D-15 |

## Gates

| Gate | Result |
|---|---|
| `make phase-gate` | exit 0, 1:26 wall clock (`ci-check`, `architecture`, `behaviour-pin-check`, `git diff --check HEAD`) |
| `make ci-check` | passes, 1:15 wall clock (limit 2:00) |
| `make architecture` | pass against the baseline pinned before the first structural edit; no cycle or layer finding |
| `make behaviour-pin-check` | matches the updated `BEHAVIOUR_PIN.txt` |
| Device path | `internal/device`, `internal/deviceworld` and `test/acceptance` (including the pinned `device serve` transcript) pass; the two device conformance expectations changed only to carry the new pointer prefix (`validate command record: /operation: missing required property …`) |
| `golangci-lint` | 0 issues |

## Known limits

- The director's `catalog describe` and the `sim://domain/<id>` MCP resource print the runtime form,
  not the v0.2 document; the canvas reads the file.
- Part 1 refuses a second kind or instance set. Adapter, scenario-flag and not-JSON issues remain
  unlocated (they are outside the domain contract).
- Run artifacts recorded before this change (v0.1 `domain_spec`) no longer replay; there is no
  compatibility, by the owner's decision.
