# Architecture and quality review — five initial lenses, seven second-pass perspectives

Status: advisory review record (input to planning; not a release gate)
Date: 2026-09-29; second pass: 2026-09-30
Snapshot: commit `4e26cc1`; first-pass clean-tree statement corrected in §8.1
Method: five independent staff-level reviews (read-only), each with a distinct lens, followed by a coordinator verification pass over every P0/P1 citation and a sample of P2/P3 citations
Owner: Streams Simulator maintainers

**Current assessment:** read §8 first. The second pass corrects evidence and severity,
adds six findings with acceptance tests, and supersedes the original remediation order.
The reports in §§1–7 record the first pass and are retained for traceability.

Lens 4 executed the suite: `go test ./... -count=1` (~40 s), `go vet ./...`,
`golangci-lint run`, and `go test -race` over `run`/`world`/`mcp` all pass at this
snapshot. Every finding below is therefore an architecture or evidence finding, not a
failing build.

## 1. Executive summary

The measurement core — world → perturb → adapter → sink, wrapped by the run/replay/ledger
machinery and the two MCP roles — is architecturally sound and unusually disciplined.
Determinism is enforced by construction (named RNG substreams, a total-ordered event queue,
an AST lint that fails on map iteration in output code, a `simdet` build tag that removes
the wall clock), truth sealing is structural (the operator view has no field that could
leak it), and the delivery ledger, quiescence barrier, injection-probe neutrality,
`silent_no_effect` scoring, and trivial-baseline audit are all proven on production paths.
The original seven-green assessment is superseded by the qualified gate assessment in §8.2.

The concerns concentrate in four places, and none of them require rework of the pipeline:

1. **The device subsystem embeds consumer knowledge in the binary and was never
   reconciled with the rules or the decision record** (R1). The second pass rates this as P1 governance debt; no P0 is established.
2. **Two Level-2 claims are weaker than they look today**: G2 has supported, shipped
   dynamic/noise forms with no analytic oracle (R2), and the release manifest is
   write-only, unsigned, and reconciled by nothing (R4) — while CI runs only `-short`
   tests, so the strongest cross-process determinism evidence never executes in the
   pipeline (R3).
3. **The offline path (`streamsim score`) is the weak seam**: it silently drops effector
   evidence and skips the contract enforcement the MCP path performs (R7).
4. **The agent-facing authority documents describe a 7-package repository that has been
   21 packages since August** (R5), and one archive line still claims a current-tree test
   failure that does not exist (R6).

Priority findings (consolidated and deduplicated across lenses; details in §4 and §5):

| ID | Sev | Finding | Lens IDs |
| --- | --- | --- | --- |
| R1 | P0 | Device emulator embeds the Agentic Stream contract in the binary; no decision record, rule text not scoped | B1, MC3, C4 |
| R2 | P1 | G2 gap: `saturation`, `hysteresis`, `random_walk` (and link-delay `lognormal`/`store_and_forward`) have no analytic oracle | D1, T3 |
| R3 | P1 | CI runs `-short` only; cross-process determinism and suite byte-identity never run in the pipeline | T1 |
| R4 | P1 | `release-manifest.json` is write-only, unsigned, stale, and verified by no test | T2 |
| R5 | P1 | AGENTS.md / `.agents/context/` describe a 7-package architecture; the tree has 21; dependency line is wrong and self-contradicts | B2, C1, C2, C5, C6, C7 |
| R6 | P1 | `docs/README.md` still claims a current-tree domain-inventory test failure (resolved by `db198bc`) | C3 |
| R7 | P1 | Offline score path: effector calls silently empty; verdict/ground-truth loaded without contract validation | B5, MC4 |
| R8 | P1 | Two JSON Schema engines enforce the same contracts with no conformance bridge | MC2 |
| R9 | P1 | `sim-event` wire format is unversioned; the stable contract has no recorded evolution policy | MC1 |
| R10 | P1 | Adapter digest — a determinism-tuple member — is computed over `encoding/json` bytes, not RFC 8785 canonical JSON | D2 |

Severity rubric used throughout: **P0** = a non-negotiable invariant violated on a
production path today; **P1** = contradicts the repo's own written bar or will cause real
defects; **P2** = design debt worth scheduling; **P3** = nit.

## 2. Cross-cutting themes

1. **Enforcement over convention is the repo's superpower — protect it.** Where a
   guarantee has an enforcement point (the determinism lint, the `simdet` tag, byte-identity
   schema tests, structural truth sealing), it holds. Where it does not (adapter digest
   canonicalization, the dormant `wall` seam with zero importers, the offline score path,
   the unverified manifest), the guarantee degrades to convention. The recurring fix is
   small: give the guarantee an enforcement point.
2. **The device subsystem is the one place code outran the rules.** It is quarantined,
   capabilities are data, and it never touches the evidence path — but it ships consumer
   names, schemas, and protocol constants in the binary, and neither AGENTS.md's
   no-consumer-knowledge rule nor DECISIONS.md was updated to accommodate it. Fix the
   rules first (R1); the code is otherwise well-built.
3. **CI parity is the gap between "proven" and "trusted".** The full suite is fast (~40 s)
   and green, but `ci-check` runs only `-short`, skipping the subprocess determinism test
   and the entire suite-generation package. Level-2 claims stay permanently
   "requires revalidation" because the pipeline cannot see the evidence (R3, R4).
4. **Documentation authority has forked.** The curated `documentation/` tree is current
   and honest about evidence status; `AGENTS.md`, `.agents/context/`, and one line of
   `docs/README.md` are stale. `documentation/architecture/overview.md` is the only page
   with a fully correct package map — it should be the source the agent-facing docs are
   regenerated from (R5, R6).

## 3. Consolidated gate assessment

Judged from code and tests at this snapshot (lenses 2, 3, 4; lens 4 ran the suites).

| Gate | Verdict | Basis and caveats |
| --- | --- | --- |
| G1 Determinism | Green locally; **CI parity gap** | Five enforcement layers (substream PRNG, total-order queue, AST lint, `simdet` tag, observed-time guard) plus byte-replay/tamper/metamorphic tests. But the cross-process test and suite byte-identity skip under `-short`, so CI never runs them (T1). |
| G2 Analytic cross-check | **At risk** | Real independent-oracle architecture, but `saturation`, `hysteresis`, `random_walk`, and the `lognormal`/`store_and_forward` delay models have none (D1/T3). |
| G3 Reference consumer | Green at protocol boundary | Out-of-process loop over streamable HTTP with real SDK client; director-tool absence asserted. Caveat: the `streamsim refconsumer` process mode itself is never spawned by a test (T4). |
| G4 Delivery ledger | Green | Append-only, terminal state fixed at write, fsync at end, conservation >10k, crash recovery, soak target. |
| G5 Quiescence | Green | Injectable clock, park hooks, stale-report rejection, timeout logged before waiting; zero sleeps. Exemplary. |
| G6 Sealed oracle | Green | Structural capability scoping (no field to leak), monotonic lifecycle, defensive clones, unblind-gated scoring. |
| G7 Injection probe | Green | Producer-controlled strings traced end-to-end; control responses are fixed text/marshaled values; http-push URL/headers code-fixed; forged citations detected against the ledger. |
| G8 `silent_no_effect` | Green | First-class shadow mode, full-tuple action scoring, consumer never trusts acks. |
| G9 Trivial baseline | Green | Hindsight-fitted panel over the immutable delivered stream with trivial/non-trivial fixtures. |

## 4. Priority findings (consolidated)

### R1 (P0) — Consumer knowledge is embedded in the binary by the device emulator, without a recorded deviation

- Evidence: `internal/device/contract.go:26-27` (`//go:embed contract/schemas/*.json`),
  `contract.go:31-35` (`maxFrameBytes` "matching the Agentic Stream codec"),
  `internal/device/device.go:407` and `codec.go:14` (Agentic Stream protocol behaviour in
  comments/constants), `internal/cli/device.go:20-22` (`device serve` exists to speak to
  the Agentic Stream serial effector), vendored `internal/device/contract/SOURCE.md`.
  Violates AGENTS.md Forbidden Changes ("Do not embed consumer knowledge (names, schemas,
  fields, behaviours) in the binary") and `.agents/context/architecture.md:41`.
- Why it matters: the invariant is what keeps adapters data and cross-consumer comparison
  honest. An unreconciled exception is precedent. `docs/DECISIONS.md` stops at D-14 with
  no device entry; the subsystem is absent from `docs/design/TECHNICAL_DESIGN.md` and
  `CHANGELOG.md`.
- Recommendation: do **not** rewrite the pipeline — the emulator is quarantined in
  `internal/device` + one CLI subcommand and never touches the evidence path, with
  capabilities loaded as data. Record an explicit deviation decision (D-15) naming the
  exception, its scope (actuation side only), and review conditions; scope the invariant
  text in AGENTS.md/`.agents/context/architecture.md` to the evidence path
  (world → perturb → adapter → sink plus both MCP surfaces) with the device emulator's own
  invariant stated (vendored, digest-pinned contract only; never domain or stream-consumer
  knowledge). Stronger follow-up if the letter of the invariant must hold: build-tag or
  separate-binary the device stack so the default instrument binary is consumer-free.

### R2 (P1) — G2 gap: supported dynamic/noise forms without analytic oracles

- Evidence: `docs/contracts/domain-spec-v0.1.schema.json:136` declares `saturation` and
  `hysteresis` dynamics; `:309` declares `random_walk` noise; `:329` declares `lognormal`
  and `store_and_forward` link delays. All are implemented (`internal/world/dynamics.go:219-235`,
  `internal/world/observe.go:213-218`, `:308-318`). A repo-wide test grep finds no oracle
  for any of them (`internal/world/oracle_test.go` covers rc_network, dead_time, fault
  envelopes, gaussian/quantization noise, availability; first-order lag is covered by the
  closed-form sweep in `world_test.go:97-135`). `saturation`/`hysteresis` are used by the
  shipped `domains/host-system-health.domain.json`. `docs/QUALITY_BAR.md:113-118` records
  commit `2977d15` as "independent oracles for every dynamic/noise/availability form".
- Why it matters: a domain author declaring `hysteresis` gets integrator code no
  independent oracle has ever checked — the "confident wrong answer" failure mode this
  repo exists to prevent, inside the sealed truth itself.
- Recommendation: add closed-form/recurrence oracles for `saturation` and `hysteresis`
  (both are exact discrete recurrences, same pattern as the existing `dead_time` oracle)
  and for `random_walk` drift (hand-rolled walker over the same named substream). For the
  link-delay models, either add an oracle or record an explicit scope decision in
  `docs/QUALITY_BAR.md` so G2's wording matches what is proven. Alternatively narrow the
  schema enums to oracled forms until proven.

### R3 (P1) — CI green is weaker than suite green: `-short` skips the strongest determinism evidence

- Evidence: `Makefile:141` — `ci-check` runs `test-short`, never the full `test` target;
  `.github/workflows/ci.yml` runs only `make ci-check`.
  `internal/cli/subprocess_test.go:21-23` skips `TestCrossProcessDeterminism` in short
  mode; `internal/suite/suite_test.go:25,81,122` skip all suite tests, including
  `TestGenerateIsByteIdentical`.
- Why it matters: the only test proving process isolation (real binary, byte-compared
  artifacts) and the suite byte-identity gate — core G1/Level-2 evidence — never execute
  in the automated pipeline. The QUALITY_BAR's "revalidation required" caveat never expires
  because CI cannot see these tests.
- Recommendation: add the full `go test -count=1 $(PKGS)` run to `ci-check` after
  `test-short` (or as a non-PR-gating CI job). The full run costs ~40 s locally.

### R4 (P1) — The release manifest is write-only, unsigned, and reconciled by nothing

- Evidence: the only Go reference is the writer, `internal/cli/manifest.go:25`
  (`cmdManifest`); zero test references (`grep -rn manifest --include='*_test.go' internal/`
  is empty); no `verify` subcommand exists; the committed `release-manifest.json` has
  `"signature": ""` and `"commit": "523fb77"` (behind HEAD). AGENTS.md states the schema
  suite reconciles manifest + domains + adapters; it reconciles schemas ↔ domains only.
  (Domain/adapter digests do happen to match the tree today, verified by hand.)
- Why it matters: the manifest is what makes a benchmark claim auditable. An unverifiable,
  unsigned manifest is a claim no one can check — the "instrument less trustworthy than
  the system it measures" failure the repo warns about.
- Recommendation: add `streamsim manifest --verify <file>` (recompute file digests, check
  `sim.commit` against the build, verify the ed25519 signature when non-empty), plus one
  table-driven test over the committed manifest; regenerate and sign for the release
  commit.

### R5 (P1) — The agent-facing authority documents describe a repository that no longer exists

- Evidence: AGENTS.md (Architecture Overview) lists 7 `internal/` packages; the tree has
  21 (`audit`, `canonical`, `cli`, `device`, `deviceworld`, `domain`, `jsonschema`,
  `model`, `randutil`, `refconsumer`, `schemas`, `score`, `suite`, `wall` undocumented —
  including the homes of non-negotiables: `refconsumer`, `audit`, `canonical`). AGENTS.md's
  dependency line `cmd -> mcp -> world/.../ledger/truth` is wrong twice: `cmd` imports only
  `internal/cli` (`cmd/streamsim/main.go:8`), and no `ledger` package exists — the same
  document says so one paragraph earlier. `.agents/context/architecture.md:17-27` lists a
  7-package "Target Structure" plus a `test/` directory that does not exist;
  `.agents/context/go-style.md` carries scaffold layering text (`server`/`service`/`store`
  packages that do not exist); `.agents/context/testing.md:18` understates what
  `make ci-check` runs (7 of 10 steps). Only `documentation/architecture/overview.md:31-39`
  is correct and current.
- Why it matters: AGENTS.md is the canonical agent entrypoint and forbids "invent[ing]
  architecture outside the documented design" while its own map omits two thirds of the
  tree. This drift is the mechanism that let R1 land without a decision record, and it
  will cause agents to misplace code or treat accepted structure as a violation.
- Recommendation: regenerate AGENTS.md's Architecture Overview and
  `.agents/context/architecture.md` from `documentation/architecture/overview.md` (the
  verified source), one line per package; fix the dependency chain to
  `cmd -> cli -> {mcp, run, suite, score, refconsumer}`; replace the go-style layering
  section with this repo's actual rules (thin MCP handlers, business logic in
  world/perturb/adapter/truth, persistence seam in `internal/run`); update
  `testing.md` and the `ci.yml` header to the real `ci-check` step list. Documentation-only.

### R6 (P1) — `docs/README.md` claims a current-tree test failure that does not exist

- Evidence: `docs/README.md:14-17` — "The current working tree has a known
  domain-inventory test failure, and manifest signing is optional." The full suite passes
  at this snapshot (lens 4); `documentation/limitations.md:47` marks the cold-chain item
  "Done"; `documentation/reference/claims-matrix.md` records the suite green at
  snapshot `523fb77`. The reconciliation commit `db198bc` (2026-09-13) touched
  `docs/README.md` and left this line behind. ("Manifest signing is optional" remains
  true.)
- Why it matters: `docs/README.md` is the archive index AGENTS.md points to as the
  evidence archive. A contributor or agent will hunt for a nonexistent failure or
  misreport repository status — the "confident wrong answer" failure mode pointed inward.
- Recommendation: replace the sentence with the reconciled formula used everywhere else
  (historical Level-2 evidence at a snapshot commit; re-run `make ci-check` and
  `streamsim manifest` for the exact commit; no benchmark publication without a fresh
  manifest).

### R7 (P1) — The offline score path is the weak seam: silent evidence loss and skipped contract enforcement

- Evidence: `internal/cli/cli.go:616-622` — `var calls []world.EffectorCall` is declared
  and passed to `score.Offline` never populated, so offline scorecards always see zero
  effector calls while the online path sees the real log (the command inputs and count exist, but outcome evidence is not persisted in the run
  artifact). The same command loads `verdict.json` and `label.json` with bare
  `json.Unmarshal` and no validation (`cli.go:604-614`), while the MCP path enforces the
  same documents twice (SDK schema at the tool boundary, `internal/mcp/schemas.go:133,173`;
  `model.ValidateVerdict` at `internal/run/run.go:813`). There is no
  `ValidateGroundTruth` in `internal/model`; `truth.Store.Seal` validates nothing beyond
  nil (`internal/truth/truth.go:97-112`). `score.go:15-18` states online and offline
  scoring MUST produce identical results.
- Why it matters: the scorecard is "the only artifact anyone outside the project reads"
  (`score.go:20`). A silently wrong action-fidelity metric, or a scorecard produced under
  different contract guarantees depending on entry point, is exactly the confident wrong
  answer the instrument must not emit.
- Recommendation: persist and validate the effector outcome log for `cmdScore` (the artifact
  contains command inputs/counts, not outcomes; see §8.2); add `model.ValidateGroundTruth` and
  call it plus `model.ValidateVerdict` in `cmdScore`. Longer term (P2): make `Offline` the
  single scoring implementation with `Score(r)` extracting the evidence bundle and
  delegating.

### R8 (P1) — Two JSON Schema engines enforce the same contracts with no conformance bridge

- Evidence: `internal/jsonschema` is a hand-rolled draft-2020-12 subset that "ignores"
  unknown keywords (`internal/jsonschema/jsonschema.go:1-14`) and validates domain,
  adapter, verdict, and artifact documents (`internal/domain/domain.go:41`,
  `internal/adapter/adapter.go:37`, `internal/model/model.go:78,95`). The MCP SDK's engine
  enforces the same ground-truth and consumer-verdict contracts as nested tool input
  schemas (`internal/mcp/schemas.go:30-40,133,173`), with its own corner-cutting
  (`$schema`/`$id`/`title` stripped, `schemas.go:35-38`). A verdict is validated by engine
  A at the tool boundary and engine B at `SubmitVerdict` (`run.go:813`).
- Why it matters: the moment a future contract uses a keyword the subset engine ignores,
  the two engines diverge and the applicable guarantee depends on the entry point.
- Recommendation: keep both engines; add one cross-engine conformance test running a table
  of valid/invalid fixtures (covering every keyword each contract uses) through both,
  asserting agreement. No new dependency.

### R9 (P1) — `sim-event` wire format is unversioned; no recorded evolution policy

- Evidence: `internal/model/event.go:8-19` has no version field;
  `docs/contracts/sim-event-v0.1.schema.json:6-8` pins `required` with
  `additionalProperties: false` and no version property. Replay is protected (the artifact
  carries `sim_version`), but a stored `trace.jsonl` cannot self-identify its envelope
  version out of band. Notably the agentic-stream *adapter* adds `storage_schema_version`
  in its preamble — versioning was solved for consumer formats but not the native stream.
- Why it matters: `sim-event` is the stable multi-consumer contract; when v0.2 arrives,
  existing traces become ambiguous out of band.
- Recommendation: no change to the stable v0.1 contract now. Record the policy
  (native stream version = `sim_version` of the producing run; traces are interpretable
  adjacent to `run.json`) in `documentation/reference/contracts.md`, and pre-commit the
  v0.2 move: an optional const version field plus reject-unknown-version rules.

### R10 (P1) — Adapter digest computed over non-canonical JSON

- Evidence: `internal/run/run.go:1020-1023` — `raw, _ := json.Marshal(a); return
  canonical.DigestBytes(raw)`, with the marshal error swallowed. Contrast the domain
  digest (`internal/domain/domain.go:52`, `canonical.Digest`) and world digest
  (`run.go:1042`). AGENTS.md: "RFC 8785 canonical JSON everywhere a digest is computed."
- Why it matters: within this binary it is deterministic and replay-verification is
  self-consistent, so nothing diverges today — but a third party verifying artifacts per
  the spec computes a different digest (`encoding/json` HTML-escapes and formats floats
  differently from RFC 8785), and the swallowed error discards the failure signal. A
  determinism-tuple member digested over non-canonical bytes contradicts a stated
  non-negotiable.
- Recommendation: canonicalize an agreed JSON-shaped adapter representation and propagate
  errors. `canonical.Digest(a)` rejects the adapter struct; see §8.2 and N1 for
  conformance and legacy-artifact compatibility requirements. (Lens 2 rated this P2 on practical impact; it is consolidated at P1
  because it contradicts the written bar.)

## 5. Lens reports

### 5.1 Lens 1 — Module boundaries & dependency direction

**Verdict.** The core measurement-path architecture is sound and unusually disciplined:
the import graph was verified acyclic and strictly downward (`go list`), `world` is free of
persistence, `sink` imports only stdlib (it cannot know consumers), `model` is free of
transport types, MCP handlers are genuinely thin, and state ownership is singular at every
level. Approvable as-is; the P0 is the device emulator's unreconciled consumer knowledge,
and the authority documents no longer describe the real package tree.

**Strengths.** Verified downward-only acyclic graph; `sink` cannot know consumers (zero
internal imports); the operator boundary is designed, not incidental
(`internal/mcp/operator.go:4-9,86-97`); handlers are translation-only
(`internal/mcp/director.go:3-6,297-307`); single-point data-driven seams (new sink = one
switch arm `internal/run/run.go:197-213`; new domain/adapter = JSON through the common
load path `internal/cli/cli.go:119-157`); dependency inversion at the device seam
(`internal/deviceworld/plant.go:50-84` routes actuation through `world.InvokeEffector`).

**Findings.**

- **B1 (P0) Consumer knowledge embedded by device emulator, no recorded deviation** — see R1.
- **B2 (P1) Authority documents no longer describe the package tree** — see R5.
- **B3 (P2) The `simdet` wall seam is dormant: `internal/wall` has zero importers**, so
  `make test-simdet` cannot fail on wall-clock misuse today; D-13 in DECISIONS.md
  describes a mechanism not yet wired (`internal/wall/wall.go`; `internal/cli/device.go:133-134`
  uses `time.Since` directly; `Makefile:117-119`). Route the documented callers through
  `wall.Now` or amend D-13 to state the seam is reserved.
- **B4 (P2) `device serve` composes a bare world inside `cli`** (`internal/cli/device.go:101`),
  creating a second world-owning path with no command-log artifact, no ledger, and no
  replay — the one mode producing no reproduction evidence; `cli` becomes a
  three-layer composer. Document the out-of-artifact status and dump a session artifact,
  or extract the gateway wiring if HIL grows.
- **B5 (P2) Scoring exists in two shapes kept identical by discipline; the offline path
  silently loses effector evidence** — see R7.
- **B6 (P3) Wall-clock command IDs in the CLI run path** — `internal/cli/cli.go:708,354`
  (`time.Now().UnixNano()` → `cli-%d`). Opaque today so determinism holds; make it a
  counter or route through `wall.Now` (see D5).
- **B7 (P3) A domain file name is hard-coded as a flag default** —
  `internal/cli/device.go:40` (`domains/cold-chain-transit.domain.json`). Require the flag
  when `--world` is set.
- **B8 (P3) The scenario-perturbation shape is defined three times**
  (`internal/suite/suite.go:30-35`, `internal/audit/audit.go:58-64`, plus run/model
  fields). Leave until a fourth appears; then hoist to `model`.

### 5.2 Lens 2 — Determinism & correctness core

**Verdict.** An unusually strong determinism core: guarantees are enforced by
architecture (build-tag wall-clock removal, single named-substream PRNG, total-ordered
event queue, an AST lint failing on map iteration in output code, a fail-closed
observed-time guard), replay verifies every input digest before execution, and truth
sealing is structural. Concerns are narrow: the G2 claim overstates coverage (R2), one
determinism-tuple member is hashed over non-canonical bytes (R10), and lazy F1 integration
makes solver and runtime trajectories agree only to grid-alignment tolerance (D3).

**Strengths.** Determinism lint that cannot be forgotten
(`internal/world/determinism_test.go:19-60`; markers at `randutil.go:109`,
`effectors.go:390`); wall clock physically removed under `simdet`
(`internal/wall/wall.go:14`); total order with explicit tiebreakers
(`internal/world/queue.go:41-46`, `world.go:286-299`, observed-time guard
`internal/run/run.go:263-266`); single RNG source with metamorphic isolation
(`internal/randutil/randutil.go:93-95`, `TestSubstreamIsolation`); fail-closed replay
(`internal/run/replay.go:89-95,120-122`); structural truth sealing
(`internal/mcp/operator.go:5-9`, `internal/truth/truth.go:123-157`); real analytic
oracles (`internal/world/oracle_test.go`); quiescence without sleeps
(`internal/run/run.go:401-407,728-736`).

**Findings.**

- **D1 (P1) G2 gap: four declared forms have no oracle** — see R2.
- **D2 (P2→R10) Adapter digest over `encoding/json` bytes** — see R10.
- **D3 (P2) Lazy F1 integration makes the trajectory depend on read times** —
  `internal/world/dynamics.go:194-207` re-anchors the RK4 grid at every read; the run path
  reads at every event time (`run.go:271-274`) while the truth solver reads on a 60 s scan
  grid (`internal/truth/solver.go:93-105`). Same class: the stochastic fault envelope
  advances lazily on read with no monotonic-read enforcement (`dynamics.go:101-110`).
  Determinism holds (the read sequence is a function of the command log), but sealed
  `first_observable_time_ns` is computed on a different grid than the scored trajectory.
  Anchor integration to an absolute grid (`floor(t/dt)*dt`) or document the tolerance in
  the SNR-based truth definition.
- **D4 (P2) `firstDivergentRecord` does not find the first divergent record** —
  `internal/run/replay.go:227-237` compares only ledger length against `art.Counts.Emitted`;
  a same-length content divergence returns -1, so the operator gets a digest mismatch with
  no bisecting hint (the comment partially discloses this). Add a rolling per-record
  digest chain to the ledger/artifact, or align the doc comment and QUALITY_BAR wording
  with the actual behaviour.
- **D5 (P3) Wall-clock values enter artifact/command-log scaffolding, bypassing the
  `wall` seam** — `internal/cli/cli.go:354,708`; `internal/run/run.go:834,943` (metadata
  only, verified not digested). Make command IDs a counter; route run-layer timestamps
  through `wall.Now` (pairs with B3).
- **D6 (P3) `resultDigest` is neither a digest nor canonical** —
  `internal/world/effectors.go:331-339` stores `json.Marshal(map[string]any{...})` output
  in a field named `ResultDigest`. Rename or route through `canonical`.
- **D7 (P3) Map-typed function parameters escape the determinism lint** —
  `internal/truth/solver.go:128`, `internal/audit/audit.go:188` range over a
  `map[string]int64` parameter; the lint collects struct fields/assignments only. No live
  bug (single-entry literals today), but teach the lint or mark the sites before a future
  multi-fault injection makes fault ordering map-iteration-dependent.

**Code-level notes (route to a code-review pass, not architecture):** `internal/sink/sink.go:52-58`
— the `MkdirAll` error is swallowed when the path's dir is `.` (inverted compound
condition) and the `return nil, nil` is unreachable; `internal/run/run.go:642-657` — an
inner `if err != nil` is always true with an unreachable `return nil, nil` in
`InvokeEffector`. Both verified by the coordinator.

### 5.3 Lens 3 — MCP protocol & contract design

*(Finding IDs prefixed MC to avoid clashing with severity labels.)*

**Verdict.** The two-role trust boundary is genuinely architectural: director and operator
are disjoint `mcp.Server` instances, the operator surface is a narrow `OperatorView` with
no field capable of holding truth/fault/perturbation state, and per-world capability
tokens (32 crypto-random bytes) gate every operator call. All six v0.1 contracts are
enforced at their load boundaries from one embedded, byte-identity-tested schema set.
Risks are forward-looking: the native wire format is unversioned (R9), two schema engines
validate the same contracts with no bridge (R8), the offline score path skips enforcement
(R7), and the no-consumer-knowledge rule is now false as written because of
`internal/device` (R1).

**Strengths.** Disjoint tool sets with the absence of director tools asserted in
`tools/list` (`internal/mcp/operator_e2e_test.go:58-77`); the operator view has nothing to
leak (`internal/mcp/operator.go:88-97`) with deliberately coarse refusals
(`operator.go:142`, `effectors.go:47-49`); contracts enforced at every boundary from one
embedded, byte-tested source (`internal/schemas/schemas.go`, `internal/domain/domain.go:41-47`,
`internal/adapter/adapter.go:37-43`, `run.go:813`); stateful truth lifecycle
(`director.go:380-386,398-399,456-458`); verdict citations verified against the ledger
(`internal/score/score.go:302-312`); G8 first-class (`effectors.go:27,123`,
`score.go:401-434`).

**Findings.**

- **MC1 (P1) sim-event unversioned; no evolution seam** — see R9.
- **MC2 (P1) Two JSON Schema engines, no conformance bridge** — see R8.
- **MC3 (P1, consolidated into R1's P0) The no-consumer-knowledge invariant and
  `internal/device` have diverged; the canonical rule is false as written** — the emulator
  is a documented deliberate component (`documentation/architecture/overview.md:53-54`,
  `documentation/limitations.md:32-39`), but the governing rule was never scoped to
  accommodate it. Fix the rule text (see R1).
- **MC4 (P2, consolidated into R7) Offline score path bypasses contract enforcement** — see R7.
- **MC5 (P2) Error taxonomy is stable in code but weakly encoded on the wire** — 15 typed
  codes (`internal/mcp/errors.go:15-46`) whose wire form is a text prefix clients parse
  (`internal/refconsumer/mcpclient.go:51-58`); no test asserts any code's wire text;
  director-side mapping collapses most failures to `domain_invalid`
  (`director.go:165,304,337,465,489`) and non-quiescence errors to `clock_backwards`
  (`:271-274`). Add a table-driven wire-prefix test; introduce a `replay_failed`-class
  code where the collapse is worst.
- **MC6 (P2) Operator HTTP endpoint posture is entirely caller-configured** —
  `internal/cli/cli.go:463-480` binds whatever `--operator-addr` says; bearer tokens ride
  plaintext HTTP; nothing warns on a non-loopback bind. Warn or refuse without an explicit
  `--allow-nonloopback` (two lines).
- **MC7 (P3) One handler breaks the thin-handler pattern** — `sim.entity.retire` reaches
  three layers down (`internal/mcp/server.go:170-179`); add `Director.RetireEntity`
  mirroring the other methods.
- **MC8 (P3) Reference consumer's core is typed on simulator internals** —
  `internal/refconsumer/refconsumer.go:18-25` returns `*world.InvokeResult`; G3 is
  honestly demonstrated at the protocol boundary, but promoting the consumer to a
  standalone binary later requires unforking. Accept and document, or give the client
  local wire shapes like `nameplateShape` already does.

### 5.4 Lens 4 — Testing & evidence architecture

**Verdict.** The gate machinery is real, not decorative: all nine gates have identifiable
tests and most are proven against production paths. Full suite passes in ~40 s; vet, lint,
and race are clean; quiescence testing is a model of deterministic clock injection. Three
claims are weaker than they look today: CI runs only `-short` (R3), the manifest is
write-only (R4), and G2 has a hole (R2).

**Commands run (all pass).** `go test ./... -count=1` (~40 s; 19 packages with tests);
`go vet ./...`; `golangci-lint run` ("0 issues"); `go test -race` over
run/world/mcp (~77 s); coverage 66.4% total; manifest domain/adapter digests verified
against the tree by hand (10/10 match).

**Gate traceability matrix.** G1 byte-replay proven on `run.New/Advance/End`
(`internal/run/run_test.go`); G1 cross-process proven but CI-skipped
(`internal/cli/subprocess_test.go`); G1 metamorphic: entity/seed isolation differential,
suite byte-identity CI-skipped; G2 proven except the four forms (R2); G3 proven at the
protocol boundary (`internal/mcp/operator_e2e_test.go`), process mode untested (T4); G4
proven (`internal/run/ledger_test.go` >10k + crash recovery; soak gated); G5 proven
(`internal/run/quiescence_test.go`, six tests, zero sleeps); G6 proven at store and
production-Director layers (`internal/truth/truth_test.go`, `internal/mcp/mcp_test.go`); G7
proven (`internal/run/probe_test.go`, `internal/perturb/perturb_test.go`,
`internal/audit/audit_test.go`); G8 proven (`internal/score/score_test.go`, full tuple); G9
proven (`internal/audit/audit_test.go`, both buckets). Hygiene scan: no `t.Setenv`, no
sleep-based synchronization in deterministic tests; golden machinery deterministic and
CI-run; fixture reconciliation real for schemas and domains
(`TestEmbeddedMatchCommitted`, `TestEmbeddedDomainSchemaValidatesAllShippedDomains`).

**Strengths.** G5 quiescence is the best-in-repo pattern (injectable `QuiescenceClock`,
park hooks, stale-report rejection); layered determinism evidence (in-process byte-replay,
AST lint, `simdet`, subprocess byte-equality); data/contract reconciliation genuinely
enforced for schemas and domains; fuzz/soak targets wired and fail-closed
(`Makefile:147-165`).

**Findings.**

- **T1 (P1) CI `-short` parity gap** — see R3.
- **T2 (P1) Manifest write-only/unsigned/unverified; AGENTS.md claim about it is false** — see R4.
- **T3 (P1) G2 hole: `saturation`/`hysteresis` un-oracled despite being shipped** — see R2.
- **T4 (P2) G3's process-mode consumer binary is untested** — `cmdRefconsumer`
  (`internal/cli/cli.go:482`, including the `--mcp` close-the-loop path) is 0.0% in the
  coverage profile; the e2e golden test runs `refconsumer.New` linked in-process. Extend
  the e2e harness (or a short-skipped sibling) to exec `streamsim refconsumer` against the
  running operator endpoint.
- **T5 (P2) The CLI command layer has ~zero in-process coverage** — `internal/cli` 9.3%;
  `cmdRun`/`cmdReplay`/`cmdMCP`/`cmdManifest`/`cmdScore` 0.0%. One table-driven test
  invoking the command functions with `t.TempDir()` outputs; also gives R4's verifier a home.
- **T6 (P2) G1 metamorphic coverage is asymmetric** — command-order, fault, and
  perturbation *isolation* have no differential test; only seed and entity isolation are
  differentially proven (`internal/perturb/perturb_test.go:204` is repeatability, not
  isolation). One run-level table-driven test varying {fault id, perturbation name,
  command order} and asserting byte-identity of unrelated substreams, following
  `TestSubstreamIsolation`'s shape.
- **T7 (P3) Zero `t.Parallel()` despite the style rule** — mark pure-function tests
  opportunistically; not load-bearing at 40 s total.
- **T8 (P3) E2E park-watchdog uses fixed 5 s wall bounds** — watchdogs, not
  synchronization (`internal/mcp/operator_e2e_test.go:163-167`); accept, and derive from
  `t.Deadline()` if it ever flakes under loaded CI.

### 5.5 Lens 5 — Documentation & design–code alignment

**Verdict.** The curated public tree is in unusually good shape: status headers are
consistent, every documented CLI command and flag matches actual `--help` output, and the
release-status reconciliation (`db198bc`) held across `limitations.md`, `claims-matrix.md`,
`release.md`, `evidence.md`, and `roadmap.md`. The drift is concentrated in the
agent-facing layer: AGENTS.md and `.agents/context/` describe a 7-package architecture
that has been 21 packages since August, with a wrong and self-contradicting dependency
line; one residue in `docs/README.md` claims a current-tree test failure that does not
exist; and the device subsystem exists entirely outside the design/decision record.

**Drift inventory.**

| # | Claim | Where | Reality | Sev |
| --- | --- | --- | --- | --- |
| 1 | 7 internal packages listed | AGENTS.md (Architecture Overview) | 21 packages; 14 undocumented incl. `refconsumer`/`audit`/`canonical` (homes of non-negotiables) | P1 |
| 2 | `cmd -> mcp -> .../ledger/truth` | AGENTS.md (Dependency direction) | `cmd -> cli -> {mcp,run,suite,score,refconsumer}`; no `ledger` package (self-contradiction in the same file) | P1 |
| 3 | "Current working tree has a known domain-inventory test failure" | `docs/README.md:14-17` | Suite green at `523fb77` and at this snapshot; item marked Done in `limitations.md:47` | P1 |
| 4 | Device subsystem exists in no design/decision record | `docs/DECISIONS.md` (ends D-14, 2026-08-18); `TECHNICAL_DESIGN.md` §2 | Added 2026-08-30 (`3693d8f`,`45ce943`,`5bd300b`); only public `reference/cli.md` + `limitations.md:32-39` describe it | P1 |
| 5 | `ci-check` = 7 steps | `.agents/context/testing.md:18` | 10 steps (`Makefile:141` adds `docs-check`, `test-simdet`, `fuzz-soak`) | P2 |
| 6 | Layering = `server`/`service`/`store` | `.agents/context/go-style.md` | No such packages; scaffold template text | P2 |
| 7 | Target structure includes `test/` dir; 14 packages absent | `.agents/context/architecture.md:17-27` | No top-level `test/`; e2e lives in-package | P2 |
| 8 | §7 index omits `DECISIONS.md` and `IMPLEMENTATION_READINESS_REVIEW.md` | `docs/README.md` §7 | Both exist; DECISIONS.md is the declared deviations authority | P2 |
| 9 | "Last verified: 2026-08-17" stamps | `documentation/architecture/overview.md:3`, `documentation/reference/cli.md:3` | Both edited later (08-30, 08-31) with post-stamp content (content is current; stamps are false) | P2 |
| 10 | "Latest archive material: 2026-08-11"; manifest toolchain `go1.27.1` unnamed anywhere | `docs/README.md:7`, `release-manifest.json:27` | Material dated to 2026-09-13; benign (manifest labeled historical) | P3 |

**Strengths.** CLI reference genuinely verified command-by-command; the status
reconciliation held with one consistent formula across five pages; `scripts/docs-check`
enforces the copyable smoke paths mechanically; archive status-labeling is disciplined
(historical material marked historical).

**Findings.**

- **C1 (P1) AGENTS.md package map is obsolete** — see R5. Highest-leverage doc fix in the
  repo; regenerate from `documentation/architecture/overview.md`.
- **C2 (P1) AGENTS.md dependency chain is wrong and self-contradicts** — see R5.
- **C3 (P1) `docs/README.md` stale test-failure claim** — see R6.
- **C4 (P1) The device subsystem has no design record and no decision record** — see R1
  (the record is the fix; the code needs no rewrite).
- **C5 (P2) `.agents/context/testing.md` understates `ci-check`** — see R5.
- **C6 (P2) `go-style.md` layering rules are scaffold text** — see R5.
- **C7 (P2) `.agents/context/architecture.md` names a phantom `test/` and leaves the
  target-vs-actual gap unnavigable** — see R5.
- **C8 (P2) The archive's own index omits two of its files** — add §7 rows for
  `DECISIONS.md` and `IMPLEMENTATION_READINESS_REVIEW.md`.
- **C9 (P2) "Last verified" stamps are false on two pages edited after their stamp** —
  re-verify and re-stamp; consider a `docs-check` warning when a stamp commit precedes the
  page's last content commit.
- **C10 (P3) Minor stale metadata** — see drift table rows 10.

## 6. Recommended remediation order

**Quick wins (small, mechanical, high trust yield):**
1. R6 — fix the one stale sentence in `docs/README.md`.
2. R5 — regenerate AGENTS.md / `.agents/context/` from `documentation/architecture/overview.md`.
3. R10 — adapter digest via `canonical.Digest` (one line).
4. R7 — populate `calls` from the artifact; add `ValidateGroundTruth` + `ValidateVerdict` to `cmdScore`.
5. R2 — two oracle tests (`saturation`, `hysteresis`), plus a scope decision for the link-delay models.
6. R3 — run the full suite in `ci-check`.
7. R1 — write decision D-15 and scope the invariant text (no code change required).
8. R4 — `manifest --verify` + one test, then regenerate/sign.

**Decisions to record (no code yet):**
- R9 — native-stream versioning policy; pre-commit the v0.2 seam shape.
- MC5 — wire-format error-code contract test; director-side code granularity.
- MC6 — non-loopback operator bind policy.
- D3 — absolute-grid integration vs documented SNR tolerance.

**Structural / schedule when the need is concrete:**
- T4 — process-mode reference-consumer e2e.
- T6 — run-level metamorphic isolation for faults/perturbations/command order.
- B3/D5 — wire the dormant `wall` seam through its documented callers.
- B4 — device-gateway session artifact (and package extraction if HIL grows).
- C8/C9 — archive index rows; verified-stamp hygiene.
- Code-level notes — `sink.NewFile` error handling, `InvokeEffector` dead branch, D4, D6, B6, B7.

## 7. Verification log

The coordinator independently re-verified every P0/P1 citation and a sample of P2/P3
citations against the snapshot tree: the device `go:embed` and Agentic Stream constants
(`internal/device/contract.go`); the absence of any world test referencing
`saturation`/`hysteresis`/`random_walk`/`lognormal`/`store_and_forward` while all are
declared in `docs/contracts/domain-spec-v0.1.schema.json`; the non-canonical adapter
digest with swallowed error (`internal/run/run.go`); the `ci-check`/`test-short`/`testing.Short`
skip chain (`Makefile`, `internal/cli/subprocess_test.go`, `internal/suite/suite_test.go`);
the manifest's zero test references, empty signature, and stale commit
(`release-manifest.json`); the never-populated `calls` slice (`internal/cli/cli.go`);
`firstDivergentRecord`'s length-only comparison (`internal/run/replay.go`); the stale
`docs/README.md` claim; the AGENTS.md package/dependency drift; and both code-level
dead-code reports (`internal/sink/sink.go`, `internal/run/run.go`). Full-suite green at
this snapshot was established by lens 4's runs, not assumed.

## 8. Second-pass review — 2026-09-30

This section supersedes the original severity rubric, gate verdicts and remediation order
where they conflict. The first-pass lens reports remain available as historical evidence.
This pass was performed by one reviewer across seven perspectives, not seven independent
reviewers. Scope: improve the architecture review and its implementation guidance; no
production changes, design exceptions, contract migrations or release claims are approved.

### 8.1 Evidence and quality bar

Reviewed commit: `4e26cc10cbaaa852326d2c4e95c09dff3c40c1da`.
The starting worktree was **not clean**: `docs/README.md` was modified and
`docs/reviews/` was untracked. Those existing changes were preserved.

A finding must identify a trigger, an enforcement point, a consequence, a confidence
level, a minimal repair, and an acceptance test. A passing suite proves only the behaviors
it exercises. Missing tests are evidence gaps, not proof of incorrect model output.

Priority is based on consequence:

- P0: demonstrated broad measurement corruption or a critical active security failure
  requiring immediate containment. **No P0 established by this pass.**
- P1: reproducible incorrect scoring, interoperability failure, evidence loss, or a
  material gap in a mandatory correctness gate.
- P2: bounded architecture, compatibility, security-posture or evidence debt.
- P3: local cleanup with no demonstrated effect on instrument trust.

The original R1 P0 conflated a governance violation with demonstrated corruption. Device
contract coupling remains a real rule violation, but is P1 governance debt with no proven
contamination of the main measurement pipeline. R5 and R6 are P2 documentation debt.
R8 and R9 are P2 preventive contract work until a concrete validation divergence or
compatibility failure is demonstrated. R2, R3, R4, R7 and R10 remain P1, with the
qualifications below. Original per-lens priority labels are historical.

### 8.2 Corrections to the first review

| Original assertion or proposed fix | Second-pass correction | Implementation consequence |
| --- | --- | --- |
| G3 is proven out of process by `operator_e2e_test.go`. | The endpoint and SDK client use real HTTP but run in the same Go test process; `refconsumer.New` is linked in-process. | Mark the external-process requirement **not established**. Spawn the consumer binary and verify a closed loop with only documented inputs and operator capabilities. |
| `simdet` physically removes wall-clock dependence. | `internal/wall` has no production importer; the tag cannot prohibit direct `time.Now` calls elsewhere. | Retain byte-replay evidence; qualify the build-tag claim. Add a targeted static rule for the deterministic core or wire and test the intended seam. Allow metadata and real timeout clocks explicitly. |
| Effector outcomes already exist in `RunArtifact`. | `internal/model/run.go:15-39` has a command log and call count, but no outcome log. Commands cannot prove refusal, applied effect, or `silent_no_effect` outcomes. | R7 needs persisted outcome evidence or an explicit unavailable-metric policy. It is not simply a slice assignment. |
| `canonical.Digest(a)` fixes R10 in one line. | `canonical.writeValue` rejects `*model.Adapter`; the helper accepts JSON-shaped values, not structs. A probe returns `unsupported value type *model.Adapter`. | Decode the agreed semantic adapter representation with number preservation, then canonicalize and propagate errors. Specify whether it includes defaults. Plan artifact compatibility before changing digest semantics. |
| AGENTS.md promises automated manifest reconciliation. | It requires coordinated inventory/test/manifest updates; it does not state that the schema suite verifies manifest digests. | Correct the claim while retaining the independently confirmed absence of a manifest verifier. |
| Optional unsigned signing and a historical manifest are defects themselves. | Both are explicitly documented policies. The confirmed gaps are no automated verifier and incomplete provenance: the manifest describes suite parameters and consumer version, not exact suite/consumer artifacts. | Add verification and exact evidence identities first. Decide authentication separately; never silently rewrite historical evidence or fabricate review identities. |
| A version field is required immediately in native events. | v0.1 is already a named closed schema. Detached-trace interpretation and future evolution need a policy; no current incompatibility was reproduced. | Record schema identity in a bundle/sidecar first. Adding a field also requires changing a schema that currently rejects unknown fields. |
| All six contracts are enforced at every boundary. | `cmdScore` uses plain decoding and suppresses artifact read errors. Direct truth-store sealing validates only nil/lifecycle conditions. | Audit raw input boundaries, then add semantic cross-file checks in addition to schema validation. |
| G2 has four missing forms. | Five distinct forms are named: two dynamics, one noise form and two delay forms. G2 explicitly requires dynamic/noise coverage; delay coverage needs an explicit scope decision. | Generate a support-to-oracle matrix rather than relying on a keyword search alone. Do not claim the implementations are wrong merely because oracle tests are absent. |

G1 therefore has useful local byte-replay evidence with CI and enforcement caveats; G2 is
incomplete; G3 has protocol-boundary evidence with process isolation unproven. G4 has
happy-path conservation/recovery tests but the durability/lifecycle issues below need
repair. G5–G9 retain the first-pass test evidence, subject to the scoring and input-boundary
qualifications. There is no defensible aggregate statement that seven gates are fully green
for release from this review alone.

### 8.3 New findings

#### N1 — P1: the canonical helper does not implement the claimed RFC 8785 byte contract

**Perspective:** correctness, interoperability and compatibility. **Confidence:** reproduced.

`internal/canonical/canonical.go:204-207` escapes U+2028/U+2029, while RFC 8785
§3.2.2.2 requires these code points to be emitted as-is. `canonicalNumber:225-245`
preserves arbitrary integer literals instead of applying the specified IEEE-754/ECMAScript
serialization. String values also accept invalid UTF-8 and silently emit replacement
characters, although object keys are checked.

Probe outputs: U+2028 produces hex `225c753230323822`, instead of `22e280a822`;
`json.Number("9007199254740993")` remains that integer, whereas the JCS numeric
representation rounds through binary64; a string containing byte `ff` becomes
`22efbfbd22` with no error. Existing `canonical_test.go:101-105` explicitly expects
extra escaping, so a green suite locks in the discrepancy.

**Root cause:** tests repeat implementation decisions rather than independently testing
standard conformance. R10's call-site repair alone cannot fix this shared trust primitive.

**Minimal repair:** introduce published conformance vectors and invalid-input cases;
decide how nanosecond integers and seeds retain exact precision at digest boundaries
(strings, restricted numeric ranges, or an explicitly named non-JCS legacy format).
Then correct canonicalization with a versioned compatibility plan. Preserve legacy artifact
verification rather than reinterpreting existing hashes.

**Acceptance:** RFC Appendix B vectors, UTF-16 ordering, U+2028/U+2029, invalid Unicode,
and equivalent numeric spellings have expected independent bytes; legacy artifacts either
verify through an explicit legacy path or fail with a clear compatibility error.

Source: [RFC 8785 §§3.1, 3.2.2 and Appendix B](https://www.rfc-editor.org/rfc/rfc8785).

#### N2 — P1: offline judgment changes when detections are reordered

**Perspective:** measurement correctness. **Confidence:** reproduced.

`internal/score/offline.go:46` compares absolute timestamp `t` to relative duration
`DetectionLatencyNS`. After selecting one detection, an earlier valid detection normally
cannot replace it. The online path instead compares against `bestAt`
(`internal/score/score.go:367-390`).

A probe with onset at `model.DefaultStartTimeNS` and detections ordered +20 s (wrong label),
then +10 s (correct label), returns latency 20 s and `label_correct=false`. Expected:
10 s and true. This is an actual online/offline disagreement, independent of missing
actuation evidence.

**Root cause:** duplicate judgment algorithms drifted; their parity tests do not challenge
input ordering.

**Minimal repair:** track an absolute best timestamp in the offline path, then share the
small judgment calculation if that removes duplication without widening package coupling.
Define equal-time tie behavior deliberately.

**Acceptance:** permutations of distinct detection times produce identical judgment;
wrong entity, pre-observable detection, negative class and label changes are tested;
online/offline comparison uses the same complete evidence and agrees field-for-field.

#### N3 — P1: offline scoring succeeds even when its run artifact cannot be decoded

**Perspective:** integrity, trust boundaries and operability. **Confidence:** reproduced.

`internal/cli/cli.go:618-621` discards `loadJSON` failures and continues with an empty
perturbation list. It neither uses `run.LoadArtifact` nor carries artifact eligibility
flags into `score.Offline`. The latter initializes neither `Reproducible` nor `Unblinded`.
Schema validation alone will not establish run/verdict/label association.

A diagnostic probe supplied `{}` for verdict and label, an empty ledger file, and
`not JSON` for `run.json`; `cmdScore` returned success and printed a scorecard. Offline
scoring also has no guard corresponding to `Director.Score`'s unblinded refusal.

**Root cause:** optional evidence is conflated with required provenance. Missing data is
represented by ordinary zero-valued metrics, which can be interpreted as measured outcomes.
This extends R7 rather than replacing it.

**Minimal repair:** fail closed on required artifact/contract errors; check run identity,
domain and seed association where represented; preserve unblinded/incomplete/reproducibility
status. Decide which incomplete runs can receive diagnostic scores and distinguish unavailable
metrics from measured false/zero results. Persist outcome evidence separately or in a
versioned bundle for full parity.

**Acceptance:** unreadable/malformed artifact, malformed verdict/label, mismatched run,
wrong domain/seed, unblinded artifact and missing outcome evidence produce explicit rejection
or documented diagnostic/unavailable status. A complete bundle yields online/offline parity.

#### N4 — P2: durable ledger descriptors have no terminal owner

**Perspective:** reliability and long-lived process scalability. **Confidence:** reproduced.

`run.New` opens `ledgerFile` (`internal/run/run.go:174-184`); successful `End` flushes
and syncs it but never closes it (`:879-887`). Initialization errors after opening it also
have no cleanup. There is no run abort/close API to own abandoned resources.

A probe configured `LedgerPath`, ended the run successfully, then successfully called
`r.ledgerFile.Stat()`: the descriptor remained open. Repeated runs can exhaust descriptors
in a long-lived director; the threshold depends on deployment limits and retained objects.

**Root cause:** finalization owns evidence publication but does not consistently own resource
release. Normal tests check artifacts, not post-terminal resource state.

**Minimal repair:** one explicit run resource cleanup owner covering successful finalization,
initialization failure, cancellation and abort. Preserve errors from flush/sync/close and
avoid closing resources while commands are active. No new lifecycle framework is needed.

**Acceptance:** successful end and each initialization/finalization failure close owned
resources exactly once; repeated creation/end remains bounded; failure artifacts remain
available with the original error preserved.

#### N5 — P1: failed artifact publication leaves a terminal run with no retry path

**Perspective:** failure recovery and evidence durability. **Confidence:** source-confirmed;
crash/power-loss behavior not experimentally measured.

`internal/run/run.go:866` sets `finished=true` before output-directory creation and all
artifact writes. If writing `run.json` then fails, `End` returns no artifact; retry returns
`run: already finished`. JSON sidecars are written directly, and `writeJSONL:993-1017`
ignores marshaling and close errors. Durable ledger append errors are deferred to the
buffer flush; that deferral needs documented acknowledgment semantics, not a claim that
all errors vanish.

**Trigger:** unwritable output, quota/disk exhaustion, serialization failure, or interruption
between sidecar publication and the final artifact. No atomic complete-bundle boundary is
established; existing crash-recovery tests do not prove power-loss durability of the bundle.

**Root cause:** stopping simulation and committing evidence are one irreversible operation.

**Minimal repair:** retain an immutable finalized snapshot and separate stopping from
publication/retry. Write files to temporary paths, check encoding/write/flush/close errors,
and publish a completion record last. Define fsync guarantees explicitly; rename alone
is not a power-loss guarantee. Keep incomplete bundles distinguishable.

**Acceptance:** inject failure at each publication stage, retry without repeating effects,
recover the same final digest and counts, and reject partial bundles as complete evidence.
A real filesystem fault test must complement injectable I/O unit tests if stronger crash
claims are made.

#### N6 — P2: file delivery does not bound the run's memory use

**Perspective:** scale, performance and maintenance cost. **Confidence:** source-confirmed;
no capacity limit or peak-memory benchmark measured in this pass.

`onEmit` appends a state-map snapshot for every native event; `appendLedger` retains every
ledger row even with durable persistence. `sink.File.Close` reads the entire trace into
memory, `End` retains it, and `Trace()` copies it. `HTTPPush` also buffers successful bytes.
Thus file output still requires memory proportional to the full run, with additional
per-event/per-state history and duplicate-delivery costs. Director world retention should
also be measured before describing long-running sessions as bounded.

**Root cause:** the sink interface requires `Close() ([]byte,error)`, coupling transport
finalization to complete trace materialization; diagnostic history is unconditional.

**Minimal repair:** measure first. Establish an event/byte/history budget and peak-memory
curve; then use incremental hashing and streamed persistence where the measured need
justifies it. Keep explicit in-memory mode for tests and small runs. Do not add a database
or silently cap evidence. Make any history reduction an explicit evidence-mode decision.

**Acceptance:** scale fixtures compare trace digest, ledger conservation and replay identity
across storage modes; benchmark peak memory across event counts and duplicate rates;
publish a supported budget and clear resource-exhaustion behavior.

### 8.4 Additional bounded perspectives

**Security.** The operator/director tool split is a valuable capability boundary. It does
not establish safe transport on arbitrary bind addresses: `serveOperatorEndpoint` exposes
plaintext HTTP at caller-selected addresses, with a 10-second header timeout. Default
loopback is appropriate for local use. Define remote deployment assumptions, TLS termination
and bearer-token transport explicitly before adding a remote mode. Unknown delivery after
HTTP response loss is another trust boundary: `HTTPPush.Write` classifies a failed request
as unsuccessful even if the receiver consumed it. Record confirmed, failed and unknown
transport outcomes or clearly constrain what “delivered” means. A receiver accepting a
record then dropping the response is the required test; no occurrence was reproduced here.
Do not infer SSRF from director-controlled URLs without an untrusted-input path.

**Concurrency.** Mutations are serialized by `commandMu`, but several run getters and
world reads are not covered by the same lock. Director map lookup locking does not provide
a coherent run snapshot. Existing race tests cover exercised schedules only. Before
supporting concurrent director commands and operator reports, add adversarial tests for
advance/read/end/unblind/score overlap and document which reads require quiescence. This is
a risk to investigate, not a reproduced race finding.

**Numerical fidelity.** D3 should be treated as a measurement question rather than an
immediate absolute-grid rewrite. Compare the same horizon under alternate observation
schedules against an independent solution, including coupled inputs and threshold crossings.
Then define the allowed error in state and first-observable time. A grid change affects
historical replay and requires an explicit version boundary. RNG-driven recurrence oracles
must independently check update timing and parameters; copying the production recurrence
and sampling schedule proves little.

**Maintainability.** Preserve the pipeline and existing package seams. Package count is
not a quality metric; document responsibility, state ownership and permitted dependencies.
The highest-value simplification is shared scoring calculations and explicit evidence
inputs, not a wholesale reorganization. R1 must receive a real design decision: either
retain a reviewed, bounded emulator exception or isolate it in a separate build artifact.
Editing AGENTS.md alone does not make the exception accepted.

**Observability and reproducibility.** Keep D4's digest-mismatch diagnostic honest. Comparing
ledger length with `Counts.Emitted` is not a reliable first-divergence index when duplicate
rows exist. Add original per-record comparison evidence only if needed; otherwise return
“index unavailable” and report input identities, completion state and trace hashes.
Lifecycle failures should identify the publication stage and recovery path. Do not log
capability tokens or private operator arguments to make diagnostics easier.

### 8.5 Root-cause synthesis — five whys

Why can green tests coexist with a wrong offline score? The offline implementation differs
from the online one. Why can it differ? Judgment and eligibility are duplicated and input
binding is weaker. Why did tests miss it? Existing parity examples do not permute detections
or exercise corrupt/mismatched bundles. Why are these inputs not modeled? Evidence arrives
as optional parallel values rather than one explicit validated scoring input. Why did this
persist? Gate descriptions are broader than the independently demonstrated cases.

The corrective principle is to put validation, evidence availability and scoring semantics
at enforceable boundaries, then challenge them with independent negative tests. This also
explains the canonicalizer and finalization gaps: tests validate ordinary operation while
interoperability and terminal failure states escape the asserted contract.

### 8.6 Ordered improvement programme

Each slice should include a reproducing/expectation test, the smallest repair, self-review,
appropriate package tests and repository gates, and updated evidence. These are proposed
implementation slices; none has been implemented by this review.

| Order | Slice | Required result before closing |
| --- | --- | --- |
| 1 | Correct offline judgment (N2). | Detection permutation test and online/offline judgment parity. |
| 2 | Validate scoring provenance and eligibility (N3/R7). | Required artifact errors fail closed; cross-file identity checks; eligibility and unavailable metrics are explicit. |
| 3 | Define and persist full scoring evidence (R7). | Outcome log and bundle contract; valid complete offline result equals online result. |
| 4 | Correct canonicalization and adapter identity (N1/R10). | Independent RFC vectors; numeric precision decision; error propagation; explicit old-artifact compatibility. |
| 5 | Repair terminal resource ownership and publication (N4/N5). | Cleanup and failure/retry tests; publication completion semantics; stable replay after recovery. |
| 6 | Close mandatory oracle/process/CI gaps (R2/R3/G3). | Support-to-oracle matrix; actual external consumer process; full deterministic tests in CI without redundant full-suite runs. |
| 7 | Add manifest verifier and exact evidence provenance (R4). | Missing/extra/changed inventory and wrong identity rejected; trusted public-key policy if authentication is required; exact suite/consumer evidence references. |
| 8 | Reconcile documented architecture and emulator scope (R1/R5/R6). | Reviewed decision, accurate responsibilities/dependencies, no stale failure claims. This low-risk documentation work may run alongside earlier slices. |
| 9 | Address measured scale and protocol needs (N6/R8/R9). | Capacity evidence, cross-validator keyword coverage/fail-closed policy, documented contract evolution. |

Contract changes, hash migrations and device exceptions require explicit design records.
Do not regenerate a signed release manifest from this advisory working tree. The first-pass
“quick wins” list is superseded: measurement correctness and evidence integrity come before
cosmetic architecture cleanup or speculative abstractions.

### 8.7 Validation and reproducible evidence

- `go test ./... -count=1`: first restricted run failed to bind local Unix/TCP sockets;
  rerun with local socket permission passed all packages (suite package approximately 38 s).
  This is current local evidence, not a remote-CI or release result.
- Four temporary diagnostic tests reproduced N1–N4 and the unsupported adapter-struct
  recommendation. All passed as **current-behavior probes**, not regression tests for a
  corrected implementation. They were removed from `internal/` after execution.
- The exact tests are retained in [SECOND_PASS_PROBES.patch](evidence/SECOND_PASS_PROBES.patch).
  On this snapshot, apply with `git apply docs/reviews/evidence/SECOND_PASS_PROBES.patch`,
  run `go test ./internal/canonical ./internal/score ./internal/cli ./internal/run
  -run '^TestArchitectureReviewProbe$' -count=1 -v`, then remove with
  `git apply -R docs/reviews/evidence/SECOND_PASS_PROBES.patch`.
  Use a disposable checkout if those filenames already exist. These probes intentionally
  assert observed defects and must become expectation-setting regression tests during repair.
- `go vet ./...`: passed.
- `go test -race ./internal/run ./internal/world ./internal/mcp ./internal/score -count=1`:
  passed with local socket permission. This is schedule-specific evidence, not proof that
  all concurrent access is safe.
- `make docs-check`, `git diff --check`, a whitespace check of the untracked review,
  and `git apply --check docs/reviews/evidence/SECOND_PASS_PROBES.patch`: passed.
  Full `make ci-check`, lint, soak, performance measurements,
  remote CI and release manifest verification were not executed by this second pass;
  no corresponding success is claimed.
