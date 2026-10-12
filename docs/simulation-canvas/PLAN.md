# Plan — Simulation canvas goal

- **Goal**: [GOAL.md](GOAL.md) · **Bar**: [QUALITY_BAR.md](QUALITY_BAR.md) · **Design**: [canvas-design.md](canvas-design.md) · **Findings**: [FINDINGS.md](FINDINGS.md)
- **Branch**: `canvas-slice` (P0–P2d merged to `main` in ghassan-ai-projects/streams-simulator#20)
- **Decisions in force**: REVIEW-2026-10-12 §1 and D1–D7 (accepted as recommended on
  2026-10-12); no backward compatibility (pre-release); the device path stays intact and gated.
- **Preflight (2026-10-12)**: DEFERRED D-08, D-15 and D-17 (director/run concurrency) are fixed
  with regression tests by the hardening program (`docs/refactoring/hardening-20261010/PLAN.md`);
  the register in `DEFERRED.md` is stale and is not relied on.

- **Working agreements (owner, 2026-10-12)**: make it work first and keep it simple; Sonnet 5.5
  subagents implement, Haiku 5.5 runs lint passes; one review per phase, run in parallel with
  the gate; findings below High are carried to a named phase instead of iterated; the full
  local gate (`make ci-check`) runs in at most 2 minutes.

Order: simulator foundations the canvas consumes (P2e–P4), then the data model (P5–P6), then
the plant wire (P7), then the canvas (P8–P11), then reference worlds and hardening (P12–P13).
The canvas starts after the structures it displays exist, so no view is built twice (confirmed
by the owner on 2026-10-12: finish the preparation before working on the UI).

Execution order: P3 → P2e → P4 → P5a → P5a2 → P5b → P6a–e → P7a–d → P8a → P8b → P9 → P10 →
P11a–c → P12 → P13.

## Phases

| # | Phase | Spec step | Delivers | Exit criteria (in addition to the standing gates) |
|---|---|---|---|---|
| P0 | Goal, bar, plan | — | GOAL, QUALITY_BAR, PLAN, FINDINGS with a disposition for every review item, findings gate, P0 review record, status banners on superseded design text | Two reviews resolved; findings gate green |
| P1 | Structured CLI output | `docs/design/CLI_SURFACE.md` and `cli-output-v0.1`: success documents per command, one error document shape with stable codes and issues (JSON pointer when known), exit codes, exempt long-running commands (`mcp`, `device serve`), the deliberate stdout change | Every non-exempt command prints exactly one schema-valid JSON document on stdout for success and failure; typed validation errors in the `domain`, `adapter` and `run` modules; `godox` and `nolintlint`; `make phase-gate` and `make behaviour-pin-check`; a pinned device-serve transcript (acceptance test) and failure documents in the behaviour pin; consumer-knowledge gate | Every command and every error class tested against the schema; pin diff recorded as deliberate |
| P2 | Scenario file | `scenario-v0.1` and semantics; terminology in `UBIQUITOUS_LANGUAGE.md` (scenario file vs profile vs suite scenario) | `streamsim run --scenario <file>`; flags build the same scenario in memory and both run through the run module's command executor; checks against domain and world before anything runs; `streamsim scenario from-run <run.json> --out <file>`; unknown profiles refused by the run module on every surface; flags accepted after positional arguments; fuzz target | Equivalence test (flags vs file); round trip run → export → run gives an identical command-log digest |
| P2b | Operator truth leak (defect found while specifying P3) | `MCP_SURFACE.md` §4.3 response shape restated as a contract the operator must not exceed | `sim.effector.invoke` returns only `{accepted, simulated, world_id, command_id, effect_eta_ns?, reason?}`: no failure mode, no `effect_applied`, no latency | Differential test: responses for `ok` and `silent_no_effect` (and every mode that acknowledges) are byte-identical; every operator response checked against an allow-list of fields |
| P2c | CLI to wiring only (structural; found while specifying P3) | `docs/refactoring/modularity-20261010/STANDARD.md` surface rules; `cli_wiring_test.go` | `internal/cli` holds flags, scenario from flags, command registry, JSON output and failure documents; MCP serving moves to `mcp` (`transport` edge), device serving to `device`, device-world wiring to `deviceworld`, the score workflow to the new `grading` module, the release manifest to the new `release` module, catalog and adapter directory loading to `domain` and `adapter`, input resolution and scenario export to `run`, reference-consumer orchestration to `refconsumer`; workflow tests to `test/acceptance`; a gate that fails when the CLI regrows logic | No behaviour change: `make behaviour-pin-check` unchanged; the CLI wiring gate proven to fail when a business function is added; every moved module has its own tests; `make phase-gate` |
| P2d | Director out of the MCP surface (structural; found while specifying P3) | `docs/refactoring/modularity-20261010/STANDARD.md` surface rules; `mcp_protocol_test.go` | New `internal/director` module owns the world registry and identities, world commands, run lifecycle, operator view and refusal taxonomy behind typed inputs; `internal/mcp` keeps tool definitions, schemas, argument decoding into typed values and transport, and reaches the simulator only through `director` and foundations; a gate fails when `mcp` regrows logic or imports `run`, `world`, `score`, `truth` or `audit` | No behaviour change: `make behaviour-pin-check` unchanged; MCP tool responses and error codes identical (the existing MCP strictness, operator and closed-loop tests pass unchanged in intent); the role gate proven to fail on an injected file; `make phase-gate` |
| P2e | Fast gate (owner request) | — | `make ci-check` runs the suite once (race, simdet build, coverage together) plus a small wall-clock run; small generated suites in tests; concurrent bounded fuzz; long work in an explicit slow tier | `make ci-check` at most 2 minutes on the development machine; every oracle (determinism, replay, cross-check, truth boundary, behaviour pin, coverage floors) still runs |
| P3 | Run journal | Journal spec and `run-journal-v0.1`: every record type tagged truth-bearing or not; who may read an in-progress run folder; sealed truth stream; unblinded (Explore) stream; pacing and lifecycle records; a reserved, versioned wire-exchange record type that P7 fills; crash behaviour | Journal appended during CLI and director runs; truth-bearing records only in the sealed stream until sealing; `verify` checks the journal against replay; fuzz target for the journal reader | Journal identical across replays; partial journal after an injected crash valid to its end; **canary test**: a distinctive injected fault in an Evaluate run appears in no consumer-readable file before sealing; Explore run stamped unblinded and refused by scoring |
| P4 | Report | `run-report-v0.1` and Markdown rules; consumer-side digests included only when the run recorded them | `streamsim report <run-dir>` (JSON and Markdown) | Pure function of the run folder; schema-valid; golden tests |
| P5a | Domain-spec v0.2: kinds and instances | v0.2 semantics part 1; `DECISIONS.md` deviation entries (incl. D-03, D-04, D-14 where affected) | Schema, loader, compiler and digest for kinds and instances compiling to the runtime world; the eight domains rewritten to v0.2; `streamsim domain format` (canonical formatting); manifest, goldens and schema suite updated | Every rewritten domain compiles to a runtime world identical to its v0.1 one (oracles, determinism, replay unchanged); every validation issue carries a JSON pointer |
| P5a2 | Multi-kind world (found in the P5a review) | v0.2 semantics part 1 amended: several kinds and instance sets per domain; namespacing of channel, fault and effector names and of profile fault weights; RNG draw order across kinds | Runtime spec with a per-kind set; each entity references its kind (state, channels, faults, effectors); one id template, count and churn per instance set; truth solver and analytic cross-check loop over kinds; `maxItems: 1` lifted | Every single-kind domain bit-identical (runtime digest test and behaviour pin); a two-kind world passes the analytic cross-check and replay identity |
| P5b | v0.2: overrides, links, system entities, board | v0.2 semantics part 1 (cont.) | Add-or-narrow overrides; typed links; system entities; board (`boot_id` owner) | Override, link and board rules proven at the domain layer; device path gate green |
| P6a | Vocabulary: `store`, `flow` | v0.2 semantics part 2 for these forms | Both forms with an independent second implementation | Analytic cross-check across swept parameters; closed-grammar test (no free-text expression anywhere in the schema) |
| P6b | Vocabulary: `regime`, `schedule`, block `peak` (`peak_mtd`, ratchet) | Same | Same | Crossing times derived independently and agreeing; same |
| P6c | Vocabulary: `boundary`, `trip`, latched `lockout`, `anti_recycle`, LOTO isolation | Same | Same | Same |
| P6d | Vocabulary: `responder` (closed comparator set; `precondition` landed in P6c as effector `requires`), `fallback_on_lease_loss`, plant sequencer entity | Same | Same | Same; no expression language |
| P6e | Event clock | Event-clock semantics with an RNG draw-order table | Sub-step events (trips, short sags) on an event clock | v0.1-equivalent worlds bit-identical; draw order proven |
| P7a | Plant wire: spec and contract | Plant-wire spec v2 applying REVIEW §4 (each item: node, reference host or next goal), D1–D5, M5 placement, the pipeline amendment (perturbation after sequencing), quiescence barrier and clock mode, `authority_epoch` decision, coexistence with the device path | `contracts/plant-wire/` schemas and fixtures (valid, invalid, scripted exchanges, lying results per command class, injection strings, must-ignore); `make contract-sync`, pinned digest, safety-field floor test | Contract reviewed; fixtures validate against the schemas |
| P7b | Plant wire: node sessions | — | Node module: node dials host (D2), handshake (version range, contract digest, environment marker), births from kinds with current values, one sequence per direction, lanes, `bd_seq`, heartbeat, rebirth; private socket | Session fixtures pass; owner-only socket; host refuses an environment mismatch |
| P7c | Plant wire: commands | — | Commands, receipts and results, freshness and dedup rules, perturbation placement, node-side digest gate for the bench (D4) | Command fixtures pass; replay reproduces the wire exchange |
| P7d | Plant wire: reference host and journal | — | Minimal conformant reference host; wire records in the journal; quiescence barrier | A slow reference host is not outrun; replay identical; device path gate green |
| P8a | Canvas foundation | Technology decision record ([canvas-technology.md](canvas-technology.md), decided 2026-10-12: TypeScript, React + Vite + React Flow, Go bridge `streamsim canvas`, HTTP JSON + Server-Sent Events), `canvas/README.md` architecture | `canvas/` toolchain and every QUALITY_BAR §3 gate on an empty shell; `make canvas-check`; canvas CI job; Go isolation gate; `AGENTS.md` canvas section; customer-workspace marker, ignore rules and check (REVIEW I3) | All gates green on the shell |
| P8b | Bridge | Bridge protocol and threat model | Bridge: director MCP client, workspace-scoped file service with atomic writes and conflict detection, live stream, token and Origin/Host checks; e2e harness running the real binary; stress fixture | One test per threat; first e2e passes against the real binary |
| P9 | Playback and review | Layout and connections schemas; performance budgets | Workspace open, run list, plant view (kinds, instances, links), timeline, playback with seek, report view, verification badge, true vs reported after sealing, reload on file change | e2e with no `streamsim` on PATH plays back a run; crashed-run and failing-verify fixtures show the right badges; budgets enforced |
| P10 | Live run and control | Bridge command set, live-state model, action registry, `.mcp.json` exposure decision (which role, no unblind for coding agents) | Director-driven runs; play, pause, step, speed, reset; live view with wire traffic and party status from session records; fault, environment and perturbation panel; Explore / Evaluate; export session via `scenario from-run`; `.mcp.json` | Parity test: every bridge action maps to a documented CLI or MCP command; Evaluate canary through bridge and page; registered MCP server cannot reveal truth on an open run |
| P11a | Builder: overview, kinds, state, behaviour | Builder interaction spec (views, panels, honesty rule) | Kind view and these panels, validate as you edit | e2e: edits produce binary-valid domains |
| P11b | Builder: sensors, actions, faults | — | These panels | Same |
| P11c | Builder: scenarios and tools | — | Scenarios panel, preview run, fault-quality check, templates, diff before save, layout sidecar | e2e: a world built in the canvas is byte-identical to the authored JSON after `domain format` |
| P12 | Reference worlds | Reference scenario v2 (REVIEW §5 items in this goal, D3) | Factory utilities domain; thermal bench twin domain | Both validate, preview and run over the plant wire with the reference host; the canvas shows wire traffic, session status and environment marker; twin event stream compared with a recorded bench run; factory delivered as "validated and run, headroom not measured" |
| P13 | Hardening and documentation | — | Accessibility audit, final budgets, public documentation in `documentation/`, `limitations.md`, D6 open-decision note, idea.md dispositions (REVIEW §8) | Every GOAL item demonstrated; no open finding for this goal |

## Progress log

The commit hash of a phase is recorded by the first commit of the next phase.

| Phase | Status | Commit | Evidence | Review |
|---|---|---|---|---|
| P0 | Done | `3e73ece` | [evidence/P0.md](evidence/P0.md) | [reviews/P0.md](reviews/P0.md) |
| P1 | Done | `511ac42` | [evidence/P1.md](evidence/P1.md) | [reviews/P1.md](reviews/P1.md) |
| P2 | Done | `ed05e15` | [evidence/P2.md](evidence/P2.md) | [reviews/P2.md](reviews/P2.md) |
| P2b | Done | `87f49e3` | [evidence/P2b.md](evidence/P2b.md) | [reviews/P2b.md](reviews/P2b.md) |
| P2c | Done | `31b6a18` | [evidence/P2c.md](evidence/P2c.md) | [reviews/P2c.md](reviews/P2c.md) |
| P2d | Done | `75d26e7` | [evidence/P2d.md](evidence/P2d.md) | [reviews/P2d.md](reviews/P2d.md) |
| P3 | Done | `d1d4db8` | [evidence/P3.md](evidence/P3.md) | [reviews/P3.md](reviews/P3.md) |
| P2e | Done | `63881c1` (merged in `f8f9d5a`) | `make ci-check` 1:24 after merging P3 (was about 4 min); details in `.agents/context/testing.md` and `docs/DECISIONS.md` D-13 | not separately reviewed (test-only; every oracle still runs) |
| P4 | Done | `1657a47` | [evidence/P4.md](evidence/P4.md) | [reviews/P4.md](reviews/P4.md) |
| P5a | Done | `2509ee6` | [evidence/P5a.md](evidence/P5a.md) | [reviews/P5a.md](reviews/P5a.md) |
| P5a2 | Done | `91b0eec` | [evidence/P5a2.md](evidence/P5a2.md) | [reviews/P5a2.md](reviews/P5a2.md) |
| P5b | Done | `0853114` | [evidence/P5b.md](evidence/P5b.md) | [reviews/P5b.md](reviews/P5b.md) |
| P6a | Done | `2e1e5e0` | [evidence/P6a.md](evidence/P6a.md) | [reviews/P6a.md](reviews/P6a.md) |
| P6b | Done | `c918425` | [evidence/P6b.md](evidence/P6b.md) | [reviews/P6b.md](reviews/P6b.md) |
| P6c | Done | `85329ba` | [evidence/P6c.md](evidence/P6c.md) | [reviews/P6c.md](reviews/P6c.md) |
| P6d | Done | `51587e3` | [evidence/P6d.md](evidence/P6d.md) | [reviews/P6d.md](reviews/P6d.md) |
